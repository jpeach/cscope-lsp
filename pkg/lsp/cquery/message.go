package cquery

import "github.com/jpeach/cscope-lsp/pkg/lsp"

// Progress is a message sent by the cquery "$cquery/progress" notification.
type Progress struct {
	IndexRequestCount      int `json:"indexRequestCount"`
	DoIDMapCount           int `json:"doIdMapCount"`
	LoadPreviousIndexCount int `json:"loadPreviousIndexCount"`
	OnIDMappedCount        int `json:"onIdMappedCount"`
	OnIndexedCount         int `json:"onIndexedCount"`
	ActiveThreads          int `json:"activeThreads"`
}

// CallHierarchyParams ...
type CallHierarchyParams struct {
	Levels       int  `json:"levels"`
	Callee       bool `json:"callee"`
	DetailedName bool `json:"detailedName"`

	TextDocument lsp.TextDocumentIdentifier `json:"textDocument"`
	Position     lsp.Position               `json:"position"`
}

type CallHierarchy struct {
	Name        string          `json:"name"`
	Location    lsp.Location    `json:"location"`
	CallType    int             `json:"callType"`
	NumChildren int             `json:"numChildren"`
	Children    []CallHierarchy `json:"children"`
}
