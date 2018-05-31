package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
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
	traceFile  = pflag.StringP("trace", "t", "", "Trace to the given file")
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

type tracer struct {
	file io.Writer
}

func (t *tracer) Read(p []byte) (int, error) {
	n, err := os.Stdin.Read(p)
	t.file.Write(p[:n])
	return n, err
}

func (t *tracer) Write(p []byte) (int, error) {
	t.file.Write(p)
	return os.Stdout.Write(p)
}

func lspInit() (*lsp.Server, error) {
	srv, err := lsp.NewServer()

	if err != nil {
		return nil, err
	}

	err = srv.Start(&lsp.ServerOpts{
		Path: *cqueryPath,
		Args: []string{},
	})

	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cache, err := filepath.Abs(".cache")
	if err != nil {
		return nil, err
	}

	opts := cquery.InitializationOptions{
		CacheDirectory: cache,
	}

	if err := lsp.Initialize(srv, cwd, opts); err != nil {
		return nil, err
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
		ptr, err := unix.Mmap(fd, 0, int(s.Size), unix.PROT_READ, unix.MAP_FILE)
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

	if *traceFile != "" {
		f, err := os.OpenFile(*traceFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", PROGNAME, err)
			os.Exit(1)
		}

		conn = cscope.Conn{
			In:  &tracer{f},
			Out: &tracer{f},
		}

		defer f.Close()
	}

	// TODO(jpeach): If we need to restart the LSP server, we also need
	// to re-initialize it. This probably means that we need to drive the
	// restart from the main loop here rather than automatically in the
	// lsp.Server.

	srv, err := lspInit()
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

			srv, err = lspInit()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: failed to start %s: %s\n",
					PROGNAME, *cqueryPath, err)
				os.Exit(1)
			}

		default:
			conn.Out.Write([]byte(fmt.Sprintf("%s: %s\n", PROGNAME, err)))
			os.Exit(1)
		}
	}
}
