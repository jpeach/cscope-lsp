package ccls

// CacheOptions collects options for caching the symbol index.
type CacheOptions struct {
	Directory string `json:"directory"`

	HierarchicalPath bool `json:"hierarchicalPath"`

	// Format can be "binary" or "json".
	Format string `json:"format"`

	RetainInMemory int `json:"retainInMemory"`
}

// InitializationOptions carries initialization options for the
// ccls language server.
//
// See https://github.com/MaskRay/ccls/wiki/Customization#initialization-options
type InitializationOptions struct {
	// Cache directory for indexed files.
	Cache CacheOptions `json:"cache"`
}
