module emoney

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.36.0-rc1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.32.1
)

// replace github.com/cosmos/cosmos-sdk => ./tmpvendor/cosmos-sdk
