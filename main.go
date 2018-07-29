package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jpeach/cscope-lsp/pkg/cscope"
	"github.com/jpeach/cscope-lsp/pkg/lsp"
	"github.com/jpeach/cscope-lsp/pkg/lsp/cquery"
	"golang.org/x/sys/unix"

	"github.com/spf13/pflag"
)

const (
	// PROGNAME is the program name used in error and log messages.
	PROGNAME = "cscope-lsp"
)

var (
	traceFile  = pflag.StringP("trace", "t", "", "Trace cscope and LSP messages to the given file")
	cqueryPath = pflag.StringP("cquery", "c", "cquery", "Path to the cquery binary")
	helpFlag   = pflag.BoolP("help", "h", false, "Print this help message")

	// The following flags are require for cscope compatibility. Vim will
	// set them when starting up the line-oriented interface, but we only
	// actually use `lineFlag`.
	lineFlag    = pflag.BoolP("line", "l", false, "Enter cscope line oriented interface")
	dFlags      = pflag.BoolP("noxref", "d", false, "Do not update the cross-reference")
	refFile     = pflag.StringP("reffile", "f", "", "Use reffile as cross-ref file name instead of cscope.out")
	prependPath = pflag.StringP("prepend", "P", "", "Prepend path to relative file names in pre-built cross-ref file")
)

func lspInit(opts []lsp.ServerOption) (*lsp.Server, error) {
	srv, err := lsp.NewServer()

	if err != nil {
		return nil, err
	}

	if err = srv.Start(opts); err != nil {
		return nil, fmt.Errorf("failed to start LSP server: %s", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cache, err := filepath.Abs(".cache")
	if err != nil {
		return nil, err
	}

	init := cquery.InitializationOptions{
		CacheDirectory: cache,
	}

	if err := lsp.Initialize(srv, cwd, init); err != nil {
		return nil, fmt.Errorf("LSP initialization failed: %s", err)
	}

	return srv, nil
}

func parseQueryPattern(spec string) (string, int, int, error) {
	parts := strings.Split(spec, ":")
	if len(parts) != 3 {
		return "", 0, 0, fmt.Errorf("invalid document position")
	}

	file, err := filepath.Abs(parts[0])
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid file '%s': %s", parts[0], err)
	}

	line, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid line number '%s': %s", parts[1], err)
	}

	col, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid column number '%s': %s", parts[1], err)
	}

	// NOTE: Comvert from Vim 1-based indices, to LSP 0-based.
	return file, line - 1, col - 1, nil
}

func uriToPath(wd string, uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		panic(fmt.Sprintf("failed to parse URI '%s': %s", uri, err))
	}

	return strings.TrimPrefix(u.Path, wd+"/")
}

func convertLocationsToResult(wd string, loc []lsp.Location) ([]cscope.Result, error) {
	results := make([]cscope.Result, 0, len(loc))

	for _, l := range loc {
		// TODO(jpeach): For Text, we can just read the first line of the
		// position.

		// TODO(jpeach): For Symbol, maybe read from the starting Position until
		// the first whitespace?

		// NOTE: We convert LSP 0-based lines back to Vim 1-based lines.
		r := cscope.Result{
			File:   uriToPath(wd, l.URI),
			Line:   l.Range.Start.Line + 1,
			Symbol: "-",
			Text:   "-",
		}

		results = append(results, r)
	}

	return results, nil
}

func convertCallsToResult(wd string, calls *cquery.CallHierarchy) ([]cscope.Result, error) {
	results := make([]cscope.Result, 0, len(calls.Children))

	for _, c := range calls.Children {
		// NOTE: We convert LSP 0-based lines back to Vim 1-based lines.
		r := cscope.Result{
			File:   uriToPath(wd, c.Location.URI),
			Line:   c.Location.Range.Start.Line + 1,
			Symbol: "-",
			Text:   c.Name,
		}

		results = append(results, r)
	}

	return results, nil

}

func resolveContainerForLocation(s *lsp.Server, results []cscope.Result, loc []lsp.Location) error {
	// Map of file path to all the symbols in that file.
	syms := map[string][]lsp.SymbolInformation{}

	// First, fetch the symbols for each file.
	for _, l := range loc {
		if _, ok := syms[l.URI]; ok {
			continue
		}

		sym, err := lsp.TextDocumentDocumentSymbol(s, l.URI)
		if err != nil {
			return err
		}

		// Make sure the symbols are sorted by their start position.
		sort.Slice(sym, func(i, j int) bool {
			return sym[i].Location.Range.Start.Line < sym[j].Location.Range.Start.Line
		})

		syms[l.URI] = sym
	}

	// Capture a function name from a string of the form
	//"type function(args)".
	matchFunctionName := regexp.MustCompile(`\s([^(\s]+)\s?\(`)

	for i, l := range loc {

		var best *lsp.SymbolInformation

		// TODO(jpeach): use a binary search ...
		for i, sym := range syms[l.URI] {
			if sym.Location.Range.After(l.Range) {
				// We sorted by start position, so
				// now we have gone too far.
				break
			}

			if !sym.Location.Range.Contains(l.Range) {
				continue
			}

			if best == nil {
				best = &syms[l.URI][i]
				continue
			}

			// Take this as the best symbol if it is shorter than
			// the one we have, since the shortest enclosing rance
			// must be the most nested scope.
			if sym.Location.Range.LineCount() < best.Location.Range.LineCount() {
				best = &syms[l.URI][i]
			}

		}

		if best != nil {
			if best.ContainerName == nil {
				results[i].Symbol = best.Name
				continue
			}

			// If we have a ContainerName, we can use that to
			// improve the Symbol. cquery uses ContainerName to
			// report the expanded name for the symbol (i.e. not
			// actually the container that encloses the symbol).
			results[i].Symbol = best.Name

			switch lsp.SymbolKind(best.Kind) {
			case lsp.SymbolKindMethod, lsp.SymbolKindFunction:
				n := matchFunctionName.FindStringSubmatch(*best.ContainerName)
				if len(n) == 2 {
					// n[0] is the entire matched string, n[1]
					// is the first captured group.
					results[i].Symbol = n[1]
				}
			default:
				// If there's no whitespace in the container name,
				// we can take the whole thing without breaking the
				// cscope protocol.
				f := strings.Fields(*best.ContainerName)
				if len(f) == 1 {
					results[i].Symbol = *best.ContainerName
				}
			}
		}
	}

	return nil
}

func resolveTextForResults(results []cscope.Result) error {

	// Map of file path to all the lines in that file.
	lines := map[string][]string{}

	for _, r := range results {
		lines[r.File] = make([]string, 0)
	}

	for f := range lines {
		fd, err := unix.Open(f, unix.O_RDONLY, 0)
		if err != nil {
			return fmt.Errorf("failed to open %s: %s", f, err)
		}

		var s unix.Stat_t
		unix.Fstat(fd, &s)

		// TODO(jpeach): on Linux use unix.MAP_POPULATE to trigger
		// readahead.
		ptr, err := unix.Mmap(fd, 0, int(s.Size), unix.PROT_READ, unix.MAP_FILE|unix.MAP_SHARED)
		if err != nil {
			return fmt.Errorf("failed to mmap %s: %s", f, err)
		}

		// TODO(jpeach): convert ptr to string without copying ...
		lines[f] = strings.Split(string(ptr), "\n")

		defer unix.Munmap(ptr)
		defer unix.Close(fd)
	}

	for i, r := range results {
		results[i].Text = lines[r.File][r.Line-1]
	}

	return nil
}

func mtime(path string) (int, error) {
	s, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return int(s.ModTime().Unix()), nil
}

func search(s *lsp.Server, q *cscope.Query) ([]cscope.Result, error) {
	wd, _ := os.Getwd()

	file, line, col, err := parseQueryPattern(q.Pattern)

	// Use the mtime as the file version since it will increment
	// when the file changes
	vers, err := mtime(file)
	if err != nil {
		return nil, err
	}

	// If cquery can't find the symbol, it will crash unless the document
	// is open. Work around that by always opening the doc just in case.
	if err := lsp.TextDocumentDidOpen(s, file, vers); err != nil {
		return nil, err
	}

	defer lsp.TextDocumentDidClose(s, file)

	switch q.Search {
	case cscope.FindSymbol:
		loc, err := lsp.TextDocumentReferences(s, file, line, col)
		if err != nil {
			return nil, err
		}

		r, err := convertLocationsToResult(wd, loc)
		if err != nil {
			return nil, err
		}

		if err = resolveTextForResults(r); err != nil {
			return nil, err
		}

		if err = resolveContainerForLocation(s, r, loc); err != nil {
			return nil, err
		}

		return r, nil

	case cscope.FindDefinition:
		loc, err := lsp.TextDocumentImplementation(s, file, line, col)
		if err != nil {
			return nil, err
		}

		if len(loc) == 0 {
			loc, err = lsp.TextDocumentDefinition(s, file, line, col)
			if err != nil {
				return nil, err
			}
		}

		if len(loc) == 0 {
			loc, err = lsp.TextDocumentTypeDefinition(s, file, line, col)
			if err != nil {
				return nil, err
			}
		}

		r, err := convertLocationsToResult(wd, loc)
		if err != nil {
			return nil, err
		}

		if err = resolveTextForResults(r); err != nil {
			return nil, err
		}

		return r, nil

	case cscope.FindCallees:
		calls, err := cquery.CalleeHierarchy(s, file, line, col)
		if err != nil {
			return nil, err
		}

		return convertCallsToResult(wd, calls)

	case cscope.FindCallers:
		calls, err := cquery.CallerHierarchy(s, file, line, col)
		if err != nil {
			return nil, err
		}

		return convertCallsToResult(wd, calls)

	case cscope.FindTextString:
		return nil, fmt.Errorf("not implemented")

	case cscope.FindEgrepPattern:
		return nil, fmt.Errorf("not implemented")

	case cscope.FindFile:
		return nil, fmt.Errorf("not implemented")

	case cscope.FindIncludingFiles:
		return nil, fmt.Errorf("not implemented")

	default:
		return nil, fmt.Errorf("invalid cscope search type '%d'", q.Search)
	}

}

func main() {
	pflag.Parse()

	if *helpFlag {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION...]\n", PROGNAME)
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		pflag.PrintDefaults()
		os.Exit(0)
	}

	conn := cscope.Conn{
		In:  os.Stdin,
		Out: os.Stdout,
	}

	lspOpts := []lsp.ServerOption{
		lsp.OptPath(*cqueryPath),
	}

	if *traceFile != "" {
		traceFd, err := os.OpenFile(*traceFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", PROGNAME, err)
			os.Exit(1)
		}

		defer traceFd.Close()

		conn = cscope.Conn{
			In:  io.TeeReader(os.Stdin, traceFd),
			Out: io.MultiWriter(os.Stdout, traceFd),
		}

		os.Stderr = traceFd

		lspOpts = append(lspOpts,
			lsp.OptTrace(traceFd),
			lsp.OptArgs([]string{
				"--log-all-to-stderr",
			}),
		)
	}

	// TODO(jpeach): If we need to restart the LSP server, we also need
	// to re-initialize it. This probably means that we need to drive the
	// restart from the main loop here rather than automatically in the
	// lsp.Server.
	srv, err := lspInit(lspOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: failed to start %s: %s\n",
			PROGNAME, *cqueryPath, err)
		os.Exit(1)
	}

	defer srv.Stop()

	for *lineFlag {
		conn.Prompt()

		query, err := conn.Read()
		if err == io.EOF || err == cscope.ErrQuit {
			os.Exit(0)
		}

		if err != nil {
			conn.Out.Write([]byte(fmt.Sprintf("%s: %s\n", PROGNAME, err)))
			continue
		}

		results, err := search(srv, query)

		switch err {
		case nil:
			if err = conn.Write(results); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", PROGNAME, err)
				os.Exit(1)
			}

		// Unfortunately, if we just exit on any error, vim
		// doesn't restart us, so if the LSP server stops for
		// any reason, we just restart it. The user experience
		// will be that one cscope search fails, but subsequent
		// ones will succeed.
		case lsp.ErrStopped:
			if err = conn.Write(results); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", PROGNAME, err)
				os.Exit(1)
			}

			srv, err = lspInit(lspOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: failed to start %s: %s\n",
					PROGNAME, *cqueryPath, err)
				os.Exit(1)
			}

		default:
			// If we get an error from the LSP server, we can show
			// it on stderr, but we still have to emit an empty cscope
			// result so that vim will complete the cscope query.
			fmt.Fprintf(os.Stderr, "%s: %s\n", PROGNAME, err)
			conn.Write([]cscope.Result{})
		}
	}
}
