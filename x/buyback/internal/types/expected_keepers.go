package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}
