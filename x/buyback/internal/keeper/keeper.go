package keeper

import (
	"time"

	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/buyback/internal/types"
	market "github.com/e-money/em-ledger/x/market/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	supply "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

type Keeper struct {
	cdc      *codec.Codec
	storeKey sdk.StoreKey

	marketKeeper  MarketKeeper
	supplyKeeper  SupplyKeeper
	stakingKeeper StakingKeeper
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, mk MarketKeeper, sk SupplyKeeper, stakingKeeper StakingKeeper) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: key,

		marketKeeper:  mk,
		supplyKeeper:  sk,
		stakingKeeper: stakingKeeper,
	}
}

func (k Keeper) GetBuybackAccount(ctx sdk.Context) supply.ModuleAccountI {
	return k.supplyKeeper.GetModuleAccount(ctx, types.AccountName)
}

func (k Keeper) CancelCurrentModuleOrders(ctx sdk.Context) {
	var (
		account = k.GetBuybackAccount(ctx)
		orders  = k.marketKeeper.GetOrdersByOwner(ctx, account.GetAddress())
	)

	for _, order := range orders {
		result, err := k.marketKeeper.CancelOrder(ctx, account.GetAddress(), order.ClientOrderID)
		if err != nil {
			panic(err)
		}

		ctx.EventManager().EmitEvents(result.Events)
	}
}

func (k Keeper) SendOrderToMarket(ctx sdk.Context, order market.Order) (*sdk.Result, error) {
	return k.marketKeeper.NewOrderSingle(ctx, order)
}

func (k Keeper) GetMarketData(ctx sdk.Context) []market.MarketData {
	return k.marketKeeper.GetInstruments(ctx)
}

func (k Keeper) GetStakingTokenDenom(ctx sdk.Context) string {
	return k.stakingKeeper.BondDenom(ctx)
}

func (k Keeper) UpdateBuybackMarket(ctx sdk.Context) bool {
	var (
		lastUpdated = &time.Time{}
		blockTime   = ctx.BlockTime()
	)

	store := ctx.KVStore(k.storeKey)
	if bz := store.Get(types.GetLastUpdatedKey()); bz != nil {
		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, lastUpdated)
		if err != nil {
			panic(err)
		}
	}

	if blockTime.Sub(*lastUpdated) < time.Hour {
		return false
	}

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(blockTime)
	if err != nil {
		panic(err)
	}

	store.Set(types.GetLastUpdatedKey(), bz)
	return true
}

func (k Keeper) BurnStakingToken(ctx sdk.Context) error {
	moduleAccount := k.GetBuybackAccount(ctx)

	stakingBalance, _ := util.SplitCoinsByDenom(moduleAccount.GetCoins(), k.stakingKeeper.BondDenom(ctx))
	if stakingBalance.IsZero() {
		return nil
	}

	return k.supplyKeeper.BurnCoins(ctx, types.ModuleName, stakingBalance)
}
