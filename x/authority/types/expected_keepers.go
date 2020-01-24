// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

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
