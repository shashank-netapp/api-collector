package main

import "fmt"

type ClientCapabilities struct {
}

// ServerCapabilities lists capabilities the server provides
type ServerCapabilities struct {
	TextDocumentSync                 interface{}                      `json:"textDocumentSync,omitempty"` // Can be either TextDocumentSyncOptions or a number
	HoverProvider                    bool                             `json:"hoverProvider,omitempty"`
	CompletionProvider               *CompletionOptions               `json:"completionProvider,omitempty"`
	SignatureHelpProvider            *SignatureHelpOptions            `json:"signatureHelpProvider,omitempty"`
	DefinitionProvider               bool                             `json:"definitionProvider,omitempty"`
	TypeDefinitionProvider           interface{}                      `json:"typeDefinitionProvider,omitempty"` // Can be either bool or a structured option
	ImplementationProvider           interface{}                      `json:"implementationProvider,omitempty"` // Can be either bool or a structured option
	ReferencesProvider               bool                             `json:"referencesProvider,omitempty"`
	DocumentHighlightProvider        bool                             `json:"documentHighlightProvider,omitempty"`
	DocumentSymbolProvider           bool                             `json:"documentSymbolProvider,omitempty"`
	WorkspaceSymbolProvider          bool                             `json:"workspaceSymbolProvider,omitempty"`
	CodeActionProvider               interface{}                      `json:"codeActionProvider,omitempty"` // Can be either bool or CodeActionOptions
	CodeLensProvider                 *CodeLensOptions                 `json:"codeLensProvider,omitempty"`
	DocumentFormattingProvider       bool                             `json:"documentFormattingProvider,omitempty"`
	DocumentRangeFormattingProvider  bool                             `json:"documentRangeFormattingProvider,omitempty"`
	DocumentOnTypeFormattingProvider *DocumentOnTypeFormattingOptions `json:"documentOnTypeFormattingProvider,omitempty"`
	RenameProvider                   interface{}                      `json:"renameProvider,omitempty"` // Can be either bool or RenameOptions
	DocumentLinkProvider             *DocumentLinkOptions             `json:"documentLinkProvider,omitempty"`
	ColorProvider                    interface{}                      `json:"colorProvider,omitempty"`        // Can be either bool or ColorProviderOptions
	FoldingRangeProvider             interface{}                      `json:"foldingRangeProvider,omitempty"` // Can be either bool or FoldingRangeProviderOptions
	DeclarationProvider              interface{}                      `json:"declarationProvider,omitempty"`  // Can be either bool or a structured option
	ExecuteCommandProvider           *ExecuteCommandOptions           `json:"executeCommandProvider,omitempty"`
	Workspace                        *WorkspaceCapabilities           `json:"workspace,omitempty"`
	Experimental                     interface{}                      `json:"experimental,omitempty"`
}

// ServerResponse represents the structure of the response from the server
type ServerResponse struct {
	Result  InitializeResult `json:"result,omitempty"`
	Error   *ResponseError   `json:"error,omitempty"`
	ID      int              `json:"id"`
	Jsonrpc string           `json:"jsonrpc"`
}

// InitializeResult represents the successful initialization response
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

// ResponseError represents an error in the initialization process
type ResponseError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    InitializeError `json:"data,omitempty"`
}

// InitializeError provides additional information about initialization errors
type InitializeError struct {
	Retry bool `json:"retry"`
}

type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// WorkspaceCapabilities describes capabilities specific to the workspace
type WorkspaceCapabilities struct {
	WorkspaceFolders *WorkspaceFolders `json:"workspaceFolders,omitempty"`
}

// WorkspaceFolders defines support for workspace folders
type WorkspaceFolders struct {
	Supported           bool        `json:"supported,omitempty"`
	ChangeNotifications interface{} `json:"changeNotifications,omitempty"` // Can be either bool or string
}

type CompletionOptions struct {
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type SignatureHelpOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type CodeLensOptions struct {
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

type DocumentOnTypeFormattingOptions struct {
	FirstTriggerCharacter string   `json:"firstTriggerCharacter"`
	MoreTriggerCharacter  []string `json:"moreTriggerCharacter"`
}

type DocumentLinkOptions struct {
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

type ExecuteCommandOptions struct {
	Commands []string `json:"commands"`
}

type TextDocumentPositionParams struct {
	/**
	 * The text document.
	 */
	TextDocument TextDocumentIdentifier `json:"textDocument"`

	/**
	 * The position inside the text document.
	 */
	Position Position `json:"position"`
}

type TextDocumentIdentifier struct {
	/**
	 * The text document's URI.
	 */
	Uri string `json:"uri"`
}

type Position struct {
	/**
	 * Line position in a document (zero-based).
	 */
	Line int `json:"line"`

	/**
	 * Character offset on a line in a document (zero-based). Assuming that the line is
	 * represented as a string, the `character` value represents the gap between the
	 * `character` and `character + 1`.
	 *
	 * If the character value is greater than the line length it defaults back to the
	 * line length.
	 */
	Character int `json:"character"`
}

type ReferenceContext struct {
	/**
	 * Include the declaration of the current symbol.
	 */
	IncludeDeclaration bool `json:"includeDeclaration"`
}

type InitializedParams struct {
}

type ProgressParams[T any] struct {
	/**
	 * The progress token provided by the client or server.
	 */
	Token ProgressToken `json:"token"`

	/**
	 * The progress data.
	 */
	Value T `json:"value"`
}

type ProgressToken = interface{}

type WorkDoneProgressParams struct {
	/**
	 * An optional token that a server can use to report work done progress.
	 */
	WorkDoneToken ProgressToken `json:"workDoneToken,omitempty"`
}

type PartialResultParams struct {
	/**
	 * An optional token that a server can use to report partial results (e.g.
	 * streaming) to the client.
	 */
	PartialResultToken ProgressToken `json:"partialResultToken,omitempty"`
}

type CallHierarchyItem struct {
	/**
	 * The name of this item.
	 */
	Name string `json:"name"`

	/**
	 * The kind of this item.
	 */
	Kind SymbolKind `json:"kind"`

	/**
	 * More detail for this item, e.g. the signature of a function.
	 */
	Detail string `json:"detail"`

	/**
	 * The resource identifier of this item.
	 */
	Uri string `json:"uri"`

	/**
	 * The range enclosing this symbol not including leading/trailing whitespace
	 * but everything else, e.g. comments and code.
	 */
	Range Range `json:"range"`

	/**
	 * The range that should be selected and revealed when this symbol is being
	 * picked, e.g. the name of a function. Must be contained by the
	 * [`range`](#CallHierarchyItem.range).
	 */
	SelectionRange Range `json:"selectionRange"`
}

type SymbolKind int

// Enumeration of SymbolKind values
const (
	SymbolKindFile SymbolKind = iota + 1
	SymbolKindModule
	SymbolKindNamespace
	SymbolKindPackage
	SymbolKindClass
	SymbolKindMethod
	SymbolKindProperty
	SymbolKindField
	SymbolKindConstructor
	SymbolKindEnum
	SymbolKindInterface
	SymbolKindFunction
	SymbolKindVariable
	SymbolKindConstant
	SymbolKindString
	SymbolKindNumber
	SymbolKindBoolean
	SymbolKindArray
	SymbolKindObject
	SymbolKindKey
	SymbolKindNull
	SymbolKindEnumMember
	SymbolKindStruct
	SymbolKindEvent
	SymbolKindOperator
	SymbolKindTypeParameter
)

type Range struct {
	/**
	 * The range's start position.
	 */
	Start Position `json:"start"`

	/**
	 * The range's end position.
	 */
	End Position `json:"end"`
}

//////////////////////////////////////////////////////////////////////////////////////
/////////////////////// Requests Structs
//////////////////////////////////////////////////////////////////////////////////////

type InitializeRequest struct {
	Jsonrpc string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  InitializeParams `json:"params"`
	ID      int              `json:"id"`
}

type InitializeParams struct {
	ProcessID             *int               `json:"processId"`
	RootPath              *string            `json:"rootPath,omitempty"`
	RootURI               string             `json:"rootUri"`
	InitializationOptions interface{}        `json:"initializationOptions,omitempty"`
	Capabilities          ClientCapabilities `json:"capabilities"`
	Trace                 string             `json:"trace,omitempty"`
	WorkspaceFolders      []WorkspaceFolder  `json:"workspaceFolders,omitempty"`
}

type ReferenceRequest struct {
	Jsonrpc string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  ReferencesParams `json:"params"`
	ID      int              `json:"id"`
}

type ReferencesParams struct {
	TextDocumentPositionParams
	Context ReferenceContext `json:"context"`
}

type CallHierarchyPrepareRequest struct {
	Jsonrpc string                     `json:"jsonrpc"`
	Method  string                     `json:"method"`
	Params  CallHierarchyPrepareParams `json:"params"`
	ID      int                        `json:"id"`
}

type CallHierarchyPrepareParams struct {
	TextDocumentPositionParams
	WorkDoneProgressParams
}

type CallHierarchyOutgoingCallRequest struct {
	Jsonrpc string                           `json:"jsonrpc"`
	Method  string                           `json:"method"`
	Params  CallHierarchyOutgoingCallsParams `json:"params"`
	ID      int                              `json:"id"`
}

type CallHierarchyOutgoingCallsParams struct {
	WorkDoneProgressParams
	PartialResultParams
	Item CallHierarchyItem `json:"item"`
}

type ImplementationRequest struct {
	Jsonrpc string               `json:"jsonrpc"`
	Method  string               `json:"method"`
	Params  ImplementationParams `json:"params"`
	ID      int                  `json:"id"`
}

type ImplementationParams struct {
	TextDocumentPositionParams
	WorkDoneProgressParams
	PartialResultParams
}

func newImplementationRequest(uri string, line, character, id int) *ImplementationRequest {
	return &ImplementationRequest{
		Jsonrpc: "2.0",
		Method:  "textDocument/implementation",
		Params: ImplementationParams{
			TextDocumentPositionParams: TextDocumentPositionParams{
				TextDocument: TextDocumentIdentifier{
					Uri: fmt.Sprintf("file://%s", uri),
				},
				Position: Position{
					Line:      line,
					Character: character,
				},
			},
		},
		ID: id,
	}
}

type CallHierarchyOutgoingCall struct {

	/**
	 * The item that is called.
	 */
	To CallHierarchyItem `json:"to"`

	/**
	 * The range at which this item is called. This is the range relative to
	 * the caller, e.g the item passed to `callHierarchy/outgoingCalls` request.
	 */
	FromRanges []Range `json:"fromRanges"`
}

type Location struct {
	Uri   string `json:"uri"`
	Range Range  `json:"range"`
}

type DocumentSymbolParams struct {
	WorkDoneProgressParams
	PartialResultParams
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DocumentSymbolRequest struct {
	Jsonrpc string               `json:"jsonrpc"`
	Method  string               `json:"method"`
	Params  DocumentSymbolParams `json:"params"`
	ID      int                  `json:"id"`
}

type DocumentSymbolReponse struct {
	/**
	 * The name of this symbol.
	 */
	Name string `json:"name"`

	/**
	 * The kind of this symbol.
	 */
	Kind SymbolKind `json:"kind"`

	/**
	 * Indicates if this symbol is deprecated.
	 *
	 * @deprecated Use tags instead
	 */
	Deprecated bool `json:"deprecated,omitempty"`

	/**
	 * The location of this symbol. The location's range is used by a tool
	 * to reveal the location in the editor. If the symbol is selected in the
	 * tool the range's start information is used to position the cursor. So
	 * the range usually spans more then the actual symbol's name and does
	 * normally include things like visibility modifiers.
	 *
	 * The range doesn't have to denote a node range in the sense of an abstract
	 * syntax tree. It can therefore not be used to re-construct a hierarchy of
	 * the symbols.
	 */
	Location Location `json:"location"`

	/**
	 * The name of the symbol containing this symbol. This information is for
	 * user interface purposes (e.g. to render a qualifier in the user interface
	 * if necessary). It can't be used to re-infer a hierarchy for the document
	 * symbols.
	 */
	ContainerName string `json:"containerName"`
}

func newDocumentSymbolRequest(fileName string, id int) *DocumentSymbolRequest {
	return &DocumentSymbolRequest{
		Jsonrpc: "2.0",
		Method:  "textDocument/documentSymbol",
		Params: DocumentSymbolParams{
			TextDocument: TextDocumentIdentifier{
				Uri: fmt.Sprintf("file://%s", fileName),
			},
		},
		ID: id,
	}
}

type HoverParams struct {
	TextDocumentPositionParams
	WorkDoneProgressParams
}

type HoverRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  HoverParams `json:"params"`
	ID      int         `json:"id"`
}

func newHoverRequest(fileName string, line, character, id int) *HoverRequest {
	return &HoverRequest{
		Jsonrpc: "2.0",
		Method:  "textDocument/hover",
		Params: HoverParams{
			TextDocumentPositionParams: TextDocumentPositionParams{
				TextDocument: TextDocumentIdentifier{
					Uri: fmt.Sprintf("file://%s", fileName),
				},
				Position: Position{
					Line:      line,
					Character: character,
				},
			},
		},
		ID: id,
	}
}
