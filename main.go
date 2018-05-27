package main

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jpeach/cscope-cquery/pkg/cscope"
	"github.com/jpeach/cscope-cquery/pkg/lsp"
	"github.com/jpeach/cscope-cquery/pkg/lsp/cquery"

	"github.com/spf13/pflag"
)

const (
	// PROGNAME is the program name used in error and log messages.
	PROGNAME = "cscope-cquery"
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

	log.Printf("start")
	err = srv.Start(&lsp.ServerOpts{
		Path: *cqueryPath,
		Args: []string{
			"--log-all-to-stderr",
			"--record=/tmp/cquery",
		},
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

	log.Printf("init")
	if err := lsp.Initialize(srv, cwd, opts); err != nil {
		return nil, err
	}

	log.Printf("done")
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

func convertLocationToResult(l *lsp.Location) (cscope.Result, error) {
	uri, err := url.Parse(l.URI)
	if err != nil {
		return cscope.Result{}, err
	}

	// TODO(jpeach): For Text, we can just read the first line of the
	// position.

	// TODO(jpeach): For Symbol, maybe read from the starting Position until
	// the first whitespace?

	// NOTE: We convert LSP 0-based lines back to Vim 1-based lines.
	r := cscope.Result{
		File:   uri.Path,
		Line:   l.Range.Start.Line + 1,
		Symbol: "-",
		Text:   "-",
	}

	return r, nil
}

func mtime(path string) (int, error) {
	s, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return int(s.ModTime().Unix()), nil
}

func handle(s *lsp.Server, q *cscope.Query) ([]cscope.Result, error) {
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
		return nil, fmt.Errorf("not implemented")

	case cscope.FindDefinition:
		loc, err := lsp.TextDocumentDefinition(s, file, line, col)
		if err != nil {
			return nil, err
		}

		results := make([]cscope.Result, 0, len(loc))
		for _, l := range loc {
			r, err := convertLocationToResult(&l)
			if err != nil {
				return nil, err
			}

			results = append(results, r)
		}

		return results, err

	case cscope.FindCallees:
		return nil, fmt.Errorf("not implemented")

	case cscope.FindReferences:
		return nil, fmt.Errorf("not implemented")

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
		f, err := os.OpenFile(*traceFile, os.O_WRONLY|os.O_CREATE, 0644)
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

	for *lineFlag {
		conn.Prompt()

		query, err := conn.Read()
		if err == io.EOF {
			os.Exit(0)
		}

		if err != nil {
			conn.Out.Write([]byte(fmt.Sprintf("%s: %s\n", PROGNAME, err)))
			continue
		}

		results, err := handle(srv, query)
		if err != nil {
			conn.Out.Write([]byte(fmt.Sprintf("%s: %s\n", PROGNAME, err)))
			continue
		}

		if err = conn.Write(results); err != nil {
			conn.Out.Write([]byte(fmt.Sprintf("%s: %s\n", PROGNAME, err)))
			os.Exit(1)
		}
	}
}
