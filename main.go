package main

import (
	"fmt"
	"io"
	"os"

	"github.com/jpeach/cscope-cquery/pkg/cscope"
	"github.com/spf13/pflag"
)

const (
	// PROGNAME is the program name used in error and log messages.
	PROGNAME = "cscope-cquery"
)

var (
	traceFile  = pflag.StringP("trace", "t", "", "Trace to the given file")
	cqueryPath = pflag.StringP("cquery", "c", "", "Path to the cquery binary")
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

	for *lineFlag {
		conn.Prompt()

		_, err := conn.Read()
		if err == io.EOF {
			os.Exit(0)
		}

		if err != nil {
			conn.Out.Write([]byte(fmt.Sprintf("%s: %s\n", PROGNAME, err)))
			continue
		}

		// TODO(jpeach): send the search to cquery and figure out
		// the result set.
		r := []cscope.Result{}
		if err = conn.Write(r); err != nil {
			conn.Out.Write([]byte(fmt.Sprintf("%s: %s\n", PROGNAME, err)))
			os.Exit(1)
		}
	}
}
