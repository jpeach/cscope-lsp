package cscope

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// SearchType specified the type of cscope search.
type SearchType int

const (
	// FindSymbol - Find this symbol
	FindSymbol SearchType = 0

	// FindDefinition - Find this function definition
	FindDefinition SearchType = 1

	// FindCallees - Find functions called by this function
	FindCallees SearchType = 2

	// FindReferences - Find references of this function
	FindReferences SearchType = 3

	// FindTextString Find this text string
	FindTextString SearchType = 4

	// Change this text string

	// FindEgrepPattern - Find this egrep pattern
	FindEgrepPattern SearchType = 6

	// FindFile - Find this file
	FindFile SearchType = 7

	// FindIncludingFiles - Find files #including this file
	FindIncludingFiles SearchType = 8
)

// Query is a cscope query.
type Query struct {
	Search  SearchType
	Pattern string
}

// Result is the result of a Query. A Query may have 0 or more results.
type Result struct {
	File   string
	Line   int
	Symbol string
	Text   string
}

// Conn is a connection from a cscope client.
type Conn struct {
	In  io.Reader
	Out io.Writer

	scanner *bufio.Scanner
}

// Prompt writes a cscope prompt to the client.
func (c *Conn) Prompt() error {
	_, err := c.Out.Write([]byte(">> "))
	return err
}

// Read reads a cscope line query from the input. The line protocol is
// very simple and consists of a digit (one of the SearchType constants),
// followed by a pattern, followed by a newline.
func (c *Conn) Read() (*Query, error) {
	if c.scanner == nil {
		c.scanner = bufio.NewScanner(c.In)
	}

	if !c.scanner.Scan() {
		// If we got EOF before reading a complete line,
		// still return EOF to the caller.
		if err := c.scanner.Err(); err != nil {
			return nil, err
		}

		return nil, io.EOF
	}

	str := c.scanner.Text()
	n, err := strconv.Atoi(string(str[0]))
	if err != nil {
		return nil, fmt.Errorf("unknown command '%s'", str)
	}

	switch SearchType(n) {
	case FindSymbol:
	case FindDefinition:
	case FindCallees:
	case FindReferences:
	case FindTextString:
	case FindEgrepPattern:
	case FindFile:
	case FindIncludingFiles:
	default:
		return nil, fmt.Errorf("invalid search type %d", n)
	}

	return &Query{
		Search:  SearchType(n),
		Pattern: str[1:],
	}, nil
}

// Write writes a set of cscope results to the output. The output
// format is a line consisting of the file name, function name, line
// number, and line text, separated by spaces.
func (c *Conn) Write(results []Result) error {
	if _, err := c.Out.Write(
		[]byte(fmt.Sprintf("cscope: %d lines\n", len(results)))); err != nil {
		return err
	}

	for _, r := range results {
		if _, err := c.Out.Write(
			[]byte(fmt.Sprintf("%s %s %d %s\n", r.File, r.Symbol, r.Line, r.Text))); err != nil {
			return err
		}
	}

	return nil
}
