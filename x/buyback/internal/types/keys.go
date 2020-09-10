package types

const (
	ModuleName = "buyback"

	StoreKey = ModuleName

	QuerierRoute = ModuleName

	// Module account identifier
	AccountName = ModuleName

	QueryBalance = "balance"
)

var (
	// IAVL Store prefixes
	keysPrefix     = []byte{0x01}
	lastUpdatedKey = []byte("lastUpdated")
)

func GetLastUpdatedKey() []byte {
	return append(keysPrefix, lastUpdatedKey...)
}
