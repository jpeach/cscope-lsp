package cquery

// InitializationOptions carries initialization options for the
// cquery language server.
//
// See https://github.com/cquery-project/cquery/blob/master/src/config.h
type InitializationOptions struct {
	// Cache directory for indexed files.
	CacheDirectory string `json:"cacheDirectory"`

	// Directory containing compile_commands.json.
	CompilationDatabaseDirectory *string `json:"compilationDatabaseDirectory,omitempty"`
}
