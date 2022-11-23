package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type (
	AccountKeeper interface {
		GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	}

	BankKeeper interface {
		GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
		InputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error
		SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
		GetSupply(ctx sdk.Context, denom string) sdk.Coin
		GetPaginatedTotalSupply(ctx sdk.Context, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error)
		AddBalanceListener(l func(sdk.Context, []sdk.AccAddress))
	}
)
