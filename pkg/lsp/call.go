package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func toURI(path string) string {
	// If this is already a URL, leave it alone.
	if strings.HasPrefix(path, "file://") {
		return path
	}

	// If we can't convert to an absolute path, just keep it
	// and hope for the best.
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Sprintf("file://%s", path)
	}

	return fmt.Sprintf("file://%s", abs)
}

// FileToLanguageID maps a file extension to a TextDocumentItem
// identifier. Currently this handles only C++ files.
func FileToLanguageID(path string) string {
	ident := map[string]string{
		"cpp": "cpp",
		"cc":  "cpp",
		"c++": "cpp",
		"hpp": "cpp",

		"c": "c",
	}

	ext := strings.ToLower(filepath.Ext(path))

	return ident[ext]
}

// Initialize ...
func Initialize(s *Server, root string, options interface{}) error {
	var res json.RawMessage

	if !filepath.IsAbs(root) {
		abs, err := filepath.Abs(root)
		if err != nil {
			return err
		}

		root = abs
	}

	err := s.Call(
		context.Background(),
		"initialize",
		&InitializeParams{
			ProcessID: os.Getpid(),
			RootURI:   root,
			Trace:     TraceMessages,
			WorkspaceFolders: []WorkspaceFolder{
				{
					URI: root,
				},
			},
			InitializationOptions: options,
		},
		&res)

	return err
}

// TextDocumentDefinition returns one or more Locations for the definition of
// the symbol at the given document position. Note that the Ranges in
// the returned locations (at least for cquery) cover the entire symbol
// (e.g. the whole class definition, not just the name).
func TextDocumentDefinition(s *Server, file string, line int, col int) ([]Location, error) {
	var loc []Location

	pos := TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{
			URI: toURI(file),
		},
		Position: Position{
			Line:      line,
			Character: col,
		},
	}

	if err := s.Call(context.Background(), "textDocument/definition", pos, &loc); err != nil {
		return nil, err
	}

	return loc, nil
}

// TextDocumentDidOpen ...
func TextDocumentDidOpen(s *Server, path string, vers int) error {
	u, err := url.Parse(path)
	if err != nil {
		return err
	}

	text, err := ioutil.ReadFile(u.Path)
	if err != nil {
		return err
	}

	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			LanguageID: FileToLanguageID(path),
			URI:        toURI(path),
			Version:    vers,
			Text:       string(text),
		},
	}

	return s.Notify(context.Background(), "textDocument/didOpen", &params)
}

// TextDocumentDidClose ...
func TextDocumentDidClose(s *Server, path string) error {
	params := DidCloseTextDocumentParams{
		TextDocument: TextDocumentIdentifier{
			URI: toURI(path),
		},
	}

	return s.Notify(context.Background(), "textDocument/didClose", &params)
}