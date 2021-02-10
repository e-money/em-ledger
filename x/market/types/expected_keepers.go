// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	types2 "github.com/e-money/em-ledger/x/authority/types"
)

type (
	AccountKeeper interface {
		GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
		AddAccountListener(func(sdk.Context, authtypes.AccountI))
	}

	BankKeeper interface {
		GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
		InputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error
		SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
		GetSupply(ctx sdk.Context) exported.SupplyI
	}

	RestrictedKeeper interface {
		GetRestrictedDenoms(sdk.Context) types2.RestrictedDenoms
	}
)
