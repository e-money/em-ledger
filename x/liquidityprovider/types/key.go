package types

const (
	ModuleName   = "liquidityprovider"
	QuerierRoute = ModuleName
	RouterKey    = ModuleName
	StoreKey     = ModuleName
)

// IAVL Store prefixes
var (
	ProviderKeyPrefix = []byte{0x00}
	// Perhaps needed for future access
	// MintableKeyPrefix   = []byte{0x01}
)
