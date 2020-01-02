// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supply "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

type (
	AccountKeeper interface {
		GetAccount(sdk.Context, sdk.AccAddress) auth.Account
		AddAccountListener(func(sdk.Context, auth.Account))
	}

	BankKeeper interface {
		SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	}

	SupplyKeeper interface {
		GetSupply(ctx sdk.Context) (supply supply.SupplyI)
	}
)
