package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/buyback/internal/types"
	market "github.com/e-money/em-ledger/x/market/types"
	ptypes "github.com/gogo/protobuf/types"
)

type Keeper struct {
	cdc      codec.BinaryMarshaler
	storeKey sdk.StoreKey

	marketKeeper   MarketKeeper
	acccountKeeper AccountKeeper
	stakingKeeper  StakingKeeper
	bankKeeper     BankKeeper
}

func NewKeeper(cdc codec.BinaryMarshaler, key sdk.StoreKey, mk MarketKeeper, ak AccountKeeper, stakingKeeper StakingKeeper, bk BankKeeper) Keeper {
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
		if err := k.marketKeeper.CancelOrder(ctx, buybackAccount, order.ClientOrderID); err != nil {
			panic(err)
		}
	}
}

func (k Keeper) SendOrderToMarket(ctx sdk.Context, order market.Order) error {
	return k.marketKeeper.NewOrderSingle(ctx, order)
}

func (k Keeper) GetBestPrice(ctx sdk.Context, src, dst string) *sdk.Dec {
	return k.marketKeeper.GetBestPrice(ctx, src, dst)
}

func (k Keeper) GetStakingTokenDenom(ctx sdk.Context) string {
	return k.stakingKeeper.BondDenom(ctx)
}

func (k Keeper) UpdateBuybackMarket(ctx sdk.Context) bool {
	var (
		blockTime      = ctx.BlockTime()
		lastUpdate     = k.GetLastUpdated(ctx)
		updateInterval = k.GetUpdateInterval(ctx)
	)

	if blockTime.Sub(lastUpdate) < updateInterval {
		return false
	}

	newState, err := ptypes.TimestampProto(blockTime)
	if err != nil {
		panic(err)
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(newState)
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

func (k Keeper) GetLastUpdated(ctx sdk.Context) time.Time {
	var lastUpdate time.Time

	store := ctx.KVStore(k.storeKey)
	if bz := store.Get(types.GetLastUpdatedKey()); bz != nil {
		var state ptypes.Timestamp
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &state)
		var err error
		lastUpdate, err = ptypes.TimestampFromProto(&state)
		if err != nil {
			panic(err)
		}
	}

	return lastUpdate
}

func (k Keeper) GetUpdateInterval(ctx sdk.Context) time.Duration {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetUpdateIntervalKey())

	var updateInterval ptypes.Duration
	k.cdc.MustUnmarshalBinaryBare(bz, &updateInterval)
	ui, err := ptypes.DurationFromProto(&updateInterval)
	if err != nil {
		panic(err.Error())
	}
	return ui
}

func (k Keeper) SetUpdateInterval(ctx sdk.Context, newVal time.Duration) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshalBinaryBare(ptypes.DurationProto(newVal))
	store.Set(types.GetUpdateIntervalKey(), bz)
}
