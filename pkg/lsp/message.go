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

// LineCount returns the length of the range in lines.
func (r Range) LineCount() int {
	return r.End.Line - r.Start.Line
}

// Contains returns true if sub is fully contained by this range.
func (r Range) Contains(sub Range) bool {
	return r.Start.Line <= sub.Start.Line &&
		r.End.Line >= sub.End.Line
}

// After ...
func (r Range) After(r2 Range) bool {
	return r.Start.Line >= r2.Start.Line
}

// Before ...
func (r Range) Before(r2 Range) bool {
	return !r.After(r2)
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

type SymbolKind int

const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

// SymbolInformation represents information about programming constructs
// like variables, classes,
type SymbolInformation struct {
	// Name is the name of this symbol.
	Name string `json:"name"`

	// Kind is the kind of this symbol.
	Kind int `json:"kind"`

	// Deprecated indicates if this symbol is deprecated.
	Deprecated bool `json:"deprecated, omitempty"`

	// Location is the location of this symbol. The location's
	// range is used by a tool to reveal the location in the
	// editor. If the symbol is selected in the tool the range's
	// start information is used to position the cursor. So the
	// range usually spans more then the actual symbol's name and
	// does normally include things like visibility modifiers.
	//
	// The range doesn't have to denote a node range in the sense
	// of a abstract syntax tree. It can therefore not be used
	// to re-construct a hierarchy of the symbols.
	Location Location `json:"location"`

	// ContainerName is the name of the symbol containing this
	// symbol. This information is for user interface purposes
	// (e.g. to render a qualifier in the user interface if
	// necessary). It can't be used to re-infer a hierarchy for
	// the document symbols.
	ContainerName *string `json:"containerName, omitempty"`
}
