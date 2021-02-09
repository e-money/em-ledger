// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
)

type (
	GasPricesKeeper interface {
		SetMinimumGasPrices(gasPricesStr string) error
	}

	SupplyKeeper interface {
		GetSupply(ctx sdk.Context) exported.SupplyI
	}
)
