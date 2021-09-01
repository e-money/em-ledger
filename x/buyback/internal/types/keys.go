package types

const (
	ModuleName = "buyback"

	StoreKey = ModuleName

	QuerierRoute = ModuleName

	// Module account identifier
	AccountName = ModuleName

	QueryBalance = "balances"
)

var (
	// IAVL Store prefixes
	keysPrefix     = []byte{0x01}
	lastUpdatedKey = []byte("lastUpdated")
	updateInterval = []byte("UpdateInterval")
)

func GetUpdateIntervalKey() []byte {
	return append(keysPrefix, updateInterval...)
}

func GetLastUpdatedKey() []byte {
	return append(keysPrefix, lastUpdatedKey...)
}
