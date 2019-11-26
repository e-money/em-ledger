package types

const (
	ModuleName   = "offer"
	StoreKey     = ModuleName
	RouterKey    = ModuleName
	QuerierRoute = ModuleName
)

var (
	// Parameter key for global order IDs
	GlobalOrderIDKey = []byte("globalOrderID")
)
