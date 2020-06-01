package cquery

import (
	"context"

	"github.com/jpeach/cscope-lsp/pkg/lsp"
)

func Callers(s *lsp.Server, file string, line int, col int) ([]lsp.Location, error) {
	var loc []lsp.Location
	params := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.FileToURI(file),
		},
		Position: lsp.Position{
			Line:      line,
			Character: col,
		},
	}

	if err := s.Call(context.Background(), "$cquery/callers", params, &loc); err != nil {
		return nil, err
	}

	return loc, nil
}

func CallerHierarchy(s *lsp.Server, file string, line int, col int) (*CallHierarchy, error) {
	var calls CallHierarchy

	params := CallHierarchyParams{
		Callee:       false,
		Levels:       1,
		DetailedName: true,
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.FileToURI(file),
		},
		Position: lsp.Position{
			Line:      line,
			Character: col,
		},
	}

	if err := s.Call(context.Background(), "$cquery/callHierarchy", params, &calls); err != nil {
		return nil, err
	}

	return &calls, nil
}

func CalleeHierarchy(s *lsp.Server, file string, line int, col int) (*CallHierarchy, error) {
	var calls CallHierarchy

	params := CallHierarchyParams{
		Callee:       true,
		Levels:       1,
		DetailedName: true,
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.FileToURI(file),
		},
		Position: lsp.Position{
			Line:      line,
			Character: col,
		},
	}

	if err := s.Call(context.Background(), "$cquery/callHierarchy", params, &calls); err != nil {
		return nil, err
	}

	return &calls, nil
}
