package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	supply "github.com/cosmos/cosmos-sdk/x/supply/exported"
	market "github.com/e-money/em-ledger/x/market/types"
)

type (
	MarketKeeper interface {
		NewOrderSingle(ctx sdk.Context, order market.Order) (*sdk.Result, error)
		GetOrdersByOwner(ctx sdk.Context, owner sdk.AccAddress) []*market.Order
		GetInstruments(ctx sdk.Context) (instrs []market.MarketData)
		CancelOrder(ctx sdk.Context, owner sdk.AccAddress, clientOrderId string) (*sdk.Result, error)
	}

	SupplyKeeper interface {
		GetModuleAccount(ctx sdk.Context, moduleName string) supply.ModuleAccountI
		BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	}

	StakingKeeper interface {
		BondDenom(sdk.Context) string
	}
)
