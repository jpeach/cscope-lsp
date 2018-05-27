package cquery

import (
	"context"

	"github.com/jpeach/cscope-lsp/pkg/lsp"
)

func Callers(s *lsp.Server, file string, line int, col int) ([]lsp.Location, error) {
	var loc []lsp.Location

	pos := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.FileToURI(file),
		},
		Position: lsp.Position{
			Line:      line,
			Character: col,
		},
	}

	if err := s.Call(context.Background(), "$cquery/callers", pos, &loc); err != nil {
		return nil, err
	}

	return loc, nil
}
