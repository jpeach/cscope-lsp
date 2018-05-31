package lsp

// String returns a pointer to its argument.
func String(s string) *string {
	return &s
}

// ClientCapabilities ...
type ClientCapabilities struct {
}

// WorkspaceFolder ...
//
// https://microsoft.github.io/language-server-protocol/specification#workspace_workspaceFolders
type WorkspaceFolder struct {
	// The associated URI for this workspace folder.
	URI string `json:"uri"`

	// The name of the workspace folder. Defaults to the
	// uri's basename.
	Name string `json:"name"`
}

const (
	// TraceOff ...
	TraceOff = "off"

	// TraceVerbose  ...
	TraceVerbose = "verbose"

	// TraceMessages ...
	TraceMessages = "messages"
)

// InitializeParams ...
//
// https://microsoft.github.io/language-server-protocol/specification#initialize
type InitializeParams struct {
	// The process Id of the parent process that started the
	// server. Is null if the process has not been started by
	// another process.  If the parent process is not alive then
	// the server should exit (see exit notification) its process.
	//
	ProcessID int `json:"processId"`

	// The rootUri of the workspace. Is null if no folder is open.
	// If both `rootPath` and `rootUri` are set `rootUri` wins.
	RootURI string `json:"rootUri,omitempty"`

	InitializationOptions interface{} `json:"initializationOptions"`

	// The capabilities provided by the client (editor or tool)
	Capabilities ClientCapabilities `json:"capabilities"`

	// The initial trace setting. If omitted trace is disabled ('off').
	//	trace?: 'off' | 'messages' | 'verbose';
	Trace string `json:"trace,omitempty"`

	// The workspace folders configured in the client when the
	// server starts.  This property is only available if the
	// client supports workspace folders.  It can be `null` if the
	// client supports workspace folders but none are configured.
	WorkspaceFolders []WorkspaceFolder `json:"workspaceFolders"`
}

// Position in a text document expressed as zero-based line and
// zero-based character offset. A position is between two characters
// like an ‘insert’ cursor in a editor.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range is a range in a text document expressed as (zero-based)
// start and end positions. A range is comparable to a selection in
// an editor. Therefore the end position is exclusive. If you want to
// specify a range that contains a line including the line ending
// character(s) then use an end position denoting the start of the
// next line.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location inside a resource, such as a
// line inside a text file.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextDocumentIdentifier identifies a text documents using a URI.
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// TextDocumentPositionParams is a parameter literal used in
// requests to pass a text document and a position inside that
// document.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// DocumentSymbolParams ...
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// TextDocumentItem is item to transfer a text document from
// the client to the server.
type TextDocumentItem struct {
	// The text document's URI.
	URI string `json:"uri"`

	// The text document's language identifier.
	LanguageID string `json:"languageId"`

	// The version number of this document (it will increase
	// after each change, including undo/redo).
	Version int `json:"version"`

	// The content of the opened text document.
	Text string `json:"text"`
}

// DidOpenTextDocumentParams is sent from the client to the server to
// signal newly opened text documents. The document’s truth is now
// managed by the client and the server must not try to read the
// document’s truth using the document’s uri. Open in this sense
// means it is managed by the client.
type DidOpenTextDocumentParams struct {
	// The document that was opened.
	TextDocument TextDocumentItem `json:"textDocument"`
}

// DidCloseTextDocumentParams is sent from the client to the server when
// the document got closed in the client. The document’s truth now
// exists where the document’s uri points to (e.g. if the document’s
// uri is a file uri the truth now exists on disk).
type DidCloseTextDocumentParams struct {
	// The document that was closed.
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// ReferenceContext ...
type ReferenceContext struct {
	// Include the declaration of the current symbol.
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// ReferenceParams is sent from the client to the server to resolve
// project-wide references for the symbol denoted by the given text
// document position.
type ReferenceParams struct {
	Context ReferenceContext `json:"context"`

	// ReferenceParams extends TextDocumentPositionParam

	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}
