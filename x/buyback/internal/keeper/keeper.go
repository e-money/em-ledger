package keeper

import (
	"time"

	"github.com/e-money/em-ledger/x/buyback/internal/types"
	market "github.com/e-money/em-ledger/x/market/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	cdc      *codec.LegacyAmino
	storeKey sdk.StoreKey

	marketKeeper   MarketKeeper
	acccountKeeper AccountKeeper
	stakingKeeper  StakingKeeper
	bankKeeper     BankKeeper
}

func NewKeeper(cdc *codec.LegacyAmino, key sdk.StoreKey, mk MarketKeeper, ak AccountKeeper, stakingKeeper StakingKeeper, bk BankKeeper) Keeper {
	return Keeper{
		cdc:            cdc,
		storeKey:       key,
		marketKeeper:   mk,
		acccountKeeper: ak,
		stakingKeeper:  stakingKeeper,
		bankKeeper:     bk,
	}
}

func (k Keeper) GetBuybackAccountAddr() sdk.AccAddress {
	return k.acccountKeeper.GetModuleAddress(types.AccountName)
}

func (k Keeper) CancelCurrentModuleOrders(ctx sdk.Context) {
	buybackAccount := k.GetBuybackAccountAddr()
	orders := k.marketKeeper.GetOrdersByOwner(ctx, buybackAccount)

	for _, order := range orders {
		result, err := k.marketKeeper.CancelOrder(ctx, buybackAccount, order.ClientOrderID)
		if err != nil {
			panic(err)
		}
		for _, ev := range result.Events {
			ctx.EventManager().EmitEvent(sdk.Event(ev))
		}
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

	updateInterval := k.GetUpdateInterval(ctx)
	if blockTime.Sub(*lastUpdated) < updateInterval {
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
	moduleAccountAddr := k.GetBuybackAccountAddr()
	stakingBalance := k.bankKeeper.GetBalance(ctx, moduleAccountAddr, k.stakingKeeper.BondDenom(ctx))
	if stakingBalance.IsZero() {
		return nil
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeBuyback,
			sdk.NewAttribute(types.AttributeKeyAction, "burn"),
			sdk.NewAttribute(types.AttributeKeyAmount, stakingBalance.String()),
		),
	})

	return k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.Coins{stakingBalance})
}

func (k Keeper) GetUpdateInterval(ctx sdk.Context) time.Duration {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetUpdateIntervalKey())

	var updateInterval time.Duration
	k.cdc.MustUnmarshalBinaryBare(bz, &updateInterval)
	return updateInterval
}

func (k Keeper) SetUpdateInterval(ctx sdk.Context, newVal time.Duration) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(newVal)
	store.Set(types.GetUpdateIntervalKey(), bz)
}
