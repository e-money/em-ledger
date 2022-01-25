package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	market "github.com/e-money/em-ledger/x/market/types"
)

type (
	MarketKeeper interface {
		NewOrderSingle(ctx sdk.Context, order market.Order) error
		GetOrdersByOwner(ctx sdk.Context, owner sdk.AccAddress) []*market.Order
		GetBestPrice(ctx sdk.Context, src, dst string) *sdk.Dec
		CancelOrder(ctx sdk.Context, owner sdk.AccAddress, clientOrderId string) error
	}

	AccountKeeper interface {
		GetModuleAddress(name string) sdk.AccAddress
	}

	BankKeeper interface {
		GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
		GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
		BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	}

	StakingKeeper interface {
		BondDenom(sdk.Context) string
	}
)
