package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	supply "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

type (
	GasPricesKeeper interface {
		SetMinimumGasPrices(gasPricesStr string) error
	}

	SupplyKeeper interface {
		GetSupply(ctx sdk.Context) (supply supply.SupplyI)
	}
)
