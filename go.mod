module github.com/e-money/em-ledger

go 1.12

require (
	github.com/Workiva/go-datastructures v1.0.50
	github.com/cosmos/cosmos-sdk v0.37.3
	github.com/gorilla/mux v1.7.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.32.6
	github.com/tendermint/tm-db v0.2.0
	github.com/tidwall/gjson v1.3.2
	github.com/tidwall/sjson v1.0.4
)

// replace github.com/cosmos/cosmos-sdk => ./tmpvendor/cosmos-sdk

// replace github.com/tendermint/tendermint => ./tmpvendor/tendermint
