// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/bank"
	supply "github.com/cosmos/cosmos-sdk/x/supply/exported"

	"github.com/e-money/em-ledger/types"
)

type (
	AccountKeeper interface {
		GetAccount(sdk.Context, sdk.AccAddress) auth.Account
		AddAccountListener(func(sdk.Context, auth.Account))
	}

	BankKeeper interface {
		InputOutputCoins(ctx sdk.Context, inputs []bank.Input, outputs []bank.Output) error
	}

	SupplyKeeper interface {
		GetSupply(ctx sdk.Context) (supply supply.SupplyI)
	}

	RestrictedKeeper interface {
		GetRestrictedDenoms(sdk.Context) types.RestrictedDenoms
	}
)
