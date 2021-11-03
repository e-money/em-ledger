package v09

import sdk "github.com/cosmos/cosmos-sdk/types"

type (
	GenesisState struct {
		AuthorityKey     sdk.AccAddress   `json:"key" yaml:"key"`
		RestrictedDenoms RestrictedDenoms `json:"restricted_denoms" yaml:"restricted_denoms"`
		MinGasPrices     sdk.DecCoins     `json:"min_gas_prices" yaml:"min_gas_prices"`
	}

	RestrictedDenoms []RestrictedDenom

	RestrictedDenom struct {
		Denom   string
		Allowed []sdk.AccAddress
	}
)
