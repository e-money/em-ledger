// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	emauthtypes "github.com/e-money/em-ledger/x/authority/types"
)

type (
	AccountKeeper interface {
		GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	}

	BankKeeper interface {
		GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
		InputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error
		SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
		GetSupply(ctx sdk.Context) exported.SupplyI
		AddBalanceListener(l func(sdk.Context, []sdk.AccAddress))
	}

	RestrictedKeeper interface {
		GetRestrictedDenoms(sdk.Context) emauthtypes.RestrictedDenoms
	}
)
