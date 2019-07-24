module tmsandbox

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.36.0-rc1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.31.5
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5

// Point directly at github.com/cosmos/cosmos-sdk when PRs 4613 and 4624 make it into a release
replace github.com/cosmos/cosmos-sdk => github.com/e-money/cosmos-sdk v0.28.2-0.20190626070956-b2c07f838bc6
