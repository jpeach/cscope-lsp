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
	RootURI *string `json:"rootUri"`

	InitializationOptions interface{} `json:"initializationOptions"`

	// The capabilities provided by the client (editor or tool)
	Capabilities ClientCapabilities `json:"capabilities"`

	// The initial trace setting. If omitted trace is disabled ('off').
	//	trace?: 'off' | 'messages' | 'verbose';
	Trace *string `json:"trace"`

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
