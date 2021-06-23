// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"fmt"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"math"
	"sync"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/market/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Gas prices must be predictable, and not depend on the number of passive orders matched.
	gasPriceNewOrder           = uint64(25000)
	gasPriceCancelReplaceOrder = uint64(25000)
	gasPriceCancelOrder        = uint64(12500)
)

var _ marketKeeper = &Keeper{}

type Keeper struct {
	key        sdk.StoreKey
	keyIndices sdk.StoreKey
	cdc        codec.BinaryMarshaler
	// instruments types.Instruments
	ak types.AccountKeeper
	bk types.BankKeeper

	// accountOrders types.Orders
	appstateInit *sync.Once
}

func NewKeeper(cdc codec.BinaryMarshaler, key sdk.StoreKey, keyIndices sdk.StoreKey, authKeeper types.AccountKeeper, bankKeeper types.BankKeeper) *Keeper {
	k := &Keeper{
		cdc:        cdc,
		key:        key,
		keyIndices: keyIndices,
		ak:         authKeeper,
		bk:         bankKeeper,

		appstateInit: new(sync.Once),
	}

	bankKeeper.AddBalanceListener(k.accountChanged)
	return k
}

func (k *Keeper) createExecutionPlan(ctx sdk.Context, SourceDenom, DestinationDenom string) types.ExecutionPlan {
	bestPlan := types.ExecutionPlan{
		Price: sdk.NewDec(math.MaxInt64),
	}

	instruments := k.GetInstruments(ctx)

	for _, firstInstrument := range instruments {
		//_, firstDenom := types.MustParseInstrumentKey(firstIt.Key())

		if firstInstrument.Source != SourceDenom {
			continue
		}

		// firstPassiveOrder := firstInstrument.Orders.LeftKey().(*types.Order)
		firstPassiveOrder := k.getBestOrder(ctx, SourceDenom, firstInstrument.Destination)
		if firstPassiveOrder == nil {
			continue
		}

		// Check direct price
		if firstInstrument.Destination == DestinationDenom {
			// Direct price is better than current plan

			planPrice := sdk.OneDec().Quo(firstPassiveOrder.Price())
			planPrice = planPrice.Add(sdk.NewDecWithPrec(1, sdk.Precision)) // Add floating point epsilon
			if planPrice.LT(bestPlan.Price) {
				bestPlan = types.ExecutionPlan{
					Price:      planPrice,
					FirstOrder: firstPassiveOrder,
				}
			}
		}

		// Check synthetic price by going through two orders:
		// (SourceDenom, X) -> (X, DestinationDenom)
		secondPassiveOrder := k.getBestOrder(ctx, firstInstrument.Destination, DestinationDenom)
		if secondPassiveOrder == nil {
			continue
		}

		planPrice := sdk.OneDec().Quo(firstPassiveOrder.Price().Mul(secondPassiveOrder.Price()))
		planPrice = planPrice.Add(sdk.NewDecWithPrec(1, sdk.Precision)) // Add floating point epsilon

		if planPrice.LT(bestPlan.Price) {
			bestPlan = types.ExecutionPlan{
				Price:       planPrice,
				FirstOrder:  firstPassiveOrder,
				SecondOrder: secondPassiveOrder,
			}
		}
	}

	return bestPlan
}

// GetSrcFromSlippage expresses the maximum source amount to spend to buy the
// requested dst amount. Taking the corresponding source amount
// (dst amount/last market price) and adding the slippage percentage is the
// resulting value. The maxSlippage decimal expresses a percentage (1 is 100%).
func (k *Keeper) GetSrcFromSlippage(
	ctx sdk.Context, srcDenom string, dst sdk.Coin, maxSlippage sdk.Dec,
) (sdk.Coin, error) {
	// ValidateBasic() for the 2 Market messages has validated the src/dst coins
	if maxSlippage.LT(sdk.ZeroDec()) {
		return sdk.Coin{}, sdkerrors.Wrapf(
			types.ErrInvalidSlippage,
			"cannot specify negative slippage %s", maxSlippage.String(),
		)
	}

	// If the order allows for slippage, adjust the source amount accordingly.
	md := k.GetInstrument(ctx, srcDenom, dst.Denom)
	if md == nil || md.LastPrice == nil {
		return sdk.Coin{}, sdkerrors.Wrapf(
			types.ErrNoMarketDataAvailable, "%v/%v", srcDenom, dst.Denom,
		)
	}

	source := dst.Amount.ToDec().Quo(*md.LastPrice)
	source = source.Mul(sdk.NewDec(1).Add(maxSlippage))

	slippageSource := sdk.NewCoin(srcDenom, source.RoundInt())
	return slippageSource, nil
}

func (k *Keeper) NewOrderSingle(ctx sdk.Context, aggressiveOrder types.Order) (*sdk.Result, error) {
	// Use a fixed gas amount
	ctx.GasMeter().ConsumeGas(gasPriceNewOrder, "NewOrderSingle")
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	// Set this to true to roll back any state changes made by the aggressive order. Used for FillOrKill orders.
	KillOrder := false
	ctx, commitTrade := ctx.CacheContext()

	defer func() {
		if KillOrder {
			return
		}

		commitTrade()
	}()

	if err := aggressiveOrder.IsValid(); err != nil {
		return nil, err
	}

	if aggressiveOrder.IsFilled() {
		return nil, sdkerrors.Wrapf(types.ErrInvalidPrice, "Order price is invalid: %s -> %s", aggressiveOrder.Source, aggressiveOrder.Destination)
	}

	owner, err := sdk.AccAddressFromBech32(aggressiveOrder.Owner)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "owner")
	}
	spendableCoins := k.bk.SpendableCoins(ctx, owner)

	// Verify account balance
	if _, anyNegative := spendableCoins.SafeSub(sdk.NewCoins(aggressiveOrder.Source)); anyNegative {
		return nil, sdkerrors.Wrapf(
			types.ErrAccountBalanceInsufficient,
			"Account %v has insufficient balance to execute trade: %v < %v",
			owner,
			spendableCoins,
			aggressiveOrder.Source,
		)
	}

	accountOrders := k.GetOrdersByOwner(ctx, owner)

	// Ensure that the market is not showing "phantom liquidity" by rejecting multiple orders in an instrument based on the same balance.
	totalSourceDemand := getOrdersSourceDemand(accountOrders, aggressiveOrder.Source.Denom, aggressiveOrder.Destination.Denom)
	totalSourceDemand = totalSourceDemand.Add(aggressiveOrder.Source)
	if _, anyNegative := spendableCoins.SafeSub(sdk.NewCoins(totalSourceDemand)); anyNegative {
		// TODO Improve message
		return nil, sdkerrors.Wrapf(types.ErrAccountBalanceInsufficientForInstrument, "")
	}

	// Verify uniqueness of client order id among active orders
	if containsClientId(accountOrders, aggressiveOrder.ClientOrderID) {
		return nil, sdkerrors.Wrap(types.ErrNonUniqueClientOrderId, aggressiveOrder.ClientOrderID)
	}

	// Verify that the destination asset actually exists on chain before creating an instrument
	if !k.assetExists(ctx, aggressiveOrder.Destination) {
		return nil, sdkerrors.Wrap(types.ErrUnknownAsset, aggressiveOrder.Destination.Denom)
	}
	k.registerMarketData(ctx, aggressiveOrder.Source.Denom, aggressiveOrder.Destination.Denom)
	k.registerMarketData(ctx, aggressiveOrder.Destination.Denom, aggressiveOrder.Source.Denom)

	// Accept order
	aggressiveOrder.ID = k.getNextOrderNumber(ctx)
	types.EmitAcceptEvent(ctx, aggressiveOrder)

	for {
		plan := k.createExecutionPlan(ctx, aggressiveOrder.Destination.Denom, aggressiveOrder.Source.Denom)
		if plan.FirstOrder == nil {
			break
		}

		if aggressiveOrder.Price().GT(plan.Price) {
			// Spread has not been crossed. Aggressive order should be added to book.
			break
		}

		// All variables are named from the perspective of the passive order

		stepDestinationFilled := plan.DestinationCapacity()
		// Don't try to fill more than either the aggressive order capacity or the plan capacity (capacity of passive orders).
		stepDestinationFilled = sdk.MinDec(stepDestinationFilled, aggressiveOrder.SourceRemaining.ToDec())

		// Do not purchase more destination tokens than the order warrants
		aggressiveDestinationRemaining := aggressiveOrder.Destination.Amount.Sub(aggressiveOrder.DestinationFilled).ToDec().Quo(plan.Price)
		stepDestinationFilled = sdk.MinDec(stepDestinationFilled, aggressiveDestinationRemaining)

		// Track aggressive fill for event
		aggressiveSourceFilled := sdk.ZeroInt()
		aggressiveDestinationFilled := sdk.ZeroInt()

		for _, passiveOrder := range []*types.Order{plan.SecondOrder, plan.FirstOrder} {
			if passiveOrder == nil {
				continue
			}

			// Use the passive order's price in the market.
			stepSourceFilled := stepDestinationFilled.Quo(passiveOrder.Price())
			if stepSourceFilled.LT(sdk.NewDec(1)) {
				stepSourceFilled = sdk.NewDec(1)
			}

			// Update the aggressive order during the plan's final step.
			if passiveOrder.Destination.Denom == aggressiveOrder.Source.Denom {
				aggressiveSourceFilled = aggressiveSourceFilled.Add(stepDestinationFilled.RoundInt())
				aggressiveOrder.SourceRemaining = aggressiveOrder.SourceRemaining.Sub(stepDestinationFilled.RoundInt())
				aggressiveOrder.SourceFilled = aggressiveOrder.SourceFilled.Add(stepDestinationFilled.RoundInt())

				// Invariant check
				if aggressiveOrder.SourceRemaining.LT(sdk.ZeroInt()) {
					panic(fmt.Sprintf("Aggressive order's SourceRemaining field is less than zero. order: %v", aggressiveOrder))
				}
			}

			if passiveOrder.Source.Denom == aggressiveOrder.Destination.Denom {
				aggressiveDestinationFilled = aggressiveDestinationFilled.Add(stepSourceFilled.RoundInt())
				aggressiveOrder.DestinationFilled = aggressiveOrder.DestinationFilled.Add(stepSourceFilled.RoundInt())

				// Invariant check
				if aggressiveOrder.DestinationFilled.GT(aggressiveOrder.Destination.Amount) {
					panic(fmt.Sprintf("Aggressive order's DestinationFilled field is greater than Destination.Amount. order: %v", aggressiveOrder))
				}
			}

			passiveOrder.SourceRemaining = passiveOrder.SourceRemaining.Sub(stepSourceFilled.RoundInt())
			passiveOrder.SourceFilled = passiveOrder.SourceFilled.Add(stepSourceFilled.RoundInt())
			passiveOrder.DestinationFilled = passiveOrder.DestinationFilled.Add(stepDestinationFilled.RoundInt())

			// Invariant checks
			if passiveOrder.SourceRemaining.LT(sdk.ZeroInt()) {
				panic(fmt.Sprintf("Passive order's SourceRemaining field is less than zero. order: %v candidate: %v", aggressiveOrder, passiveOrder))
			}
			if passiveOrder.DestinationFilled.GT(passiveOrder.Destination.Amount) {
				panic(fmt.Sprintf("Passive order's DestinationFilled field is greater than Destination.Amount. order: %v", passiveOrder))
			}

			// Settle traded tokens
			nextDestinationFilledCoin := sdk.NewCoin(passiveOrder.Destination.Denom, stepDestinationFilled.RoundInt())
			nextSourceFilledCoin := sdk.NewCoin(passiveOrder.Source.Denom, stepSourceFilled.RoundInt())
			err := k.transferTradedAmounts(ctx, nextDestinationFilledCoin, nextSourceFilledCoin, passiveOrder.Owner, aggressiveOrder.Owner)
			if err != nil {
				panic(err)
			}

			types.EmitFillEvent(ctx, *passiveOrder, false, stepSourceFilled.RoundInt(), stepDestinationFilled.RoundInt())

			if passiveOrder.IsFilled() {
				k.deleteOrder(ctx, passiveOrder)
				types.EmitExpireEvent(ctx, *passiveOrder)
			} else {
				k.setOrder(ctx, passiveOrder)
			}

			// Register trades in market data
			k.setMarketData(ctx, passiveOrder.Source.Denom, passiveOrder.Destination.Denom, passiveOrder.Price())
			k.setMarketData(ctx, passiveOrder.Destination.Denom, passiveOrder.Source.Denom, sdk.NewDec(1).Quo(passiveOrder.Price()))

			stepDestinationFilled = stepSourceFilled
		}

		types.EmitFillEvent(ctx, aggressiveOrder, true, aggressiveSourceFilled, aggressiveDestinationFilled)

		// Register trades in market data
		k.setMarketData(ctx, aggressiveOrder.Source.Denom, aggressiveOrder.Destination.Denom, plan.Price)
		k.setMarketData(ctx, aggressiveOrder.Destination.Denom, aggressiveOrder.Source.Denom, sdk.NewDec(1).Quo(plan.Price))

		if aggressiveOrder.IsFilled() {
			break
		}
	}

	if aggressiveOrder.IsFilled() {
		types.EmitExpireEvent(ctx, aggressiveOrder)
	} else {
		addToBook := true

		switch aggressiveOrder.TimeInForce {
		case types.TimeInForce_ImmediateOrCancel:
			addToBook = false
			types.EmitExpireEvent(ctx, aggressiveOrder)
		case types.TimeInForce_FillOrKill:
			KillOrder = true
			addToBook = false
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			types.EmitExpireEvent(ctx, aggressiveOrder)
		}

		if addToBook {
			op := &aggressiveOrder
			k.setOrder(ctx, op)

			// NOTE This should be the only place that an order is added to the book!
			// NOTE If this ceases to be true, move logic to func that cleans up all datastructures.
		}
	}

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k *Keeper) initializeFromStore(ctx sdk.Context) {
	k.appstateInit.Do(func() {
		// TODO Reinstate this when the mem store arrives in v0.40 of the Cosmos SDK.
		//// Load the last known market state from app state.
		//store := ctx.KVStore(k.key)
		//idxStore := ctx.KVStore(k.keyIndices)
		//
		//ownersPrefix := types.GetOwnersPrefix()
		//it := store.Iterator(ownersPrefix, nil)
		//if it.Valid() {
		//	defer it.Close()
		//}
		//
		//for ; it.Valid(); it.Next() {
		//
		//	o := &types.Order{}
		//	err := k.cdc.UnmarshalBinaryBare(it.Value(), o)
		//	if err != nil {
		//		panic(err)
		//	}
		//
		//	key := types.GetPricingKey(o.Source.Denom, o.Destination.Denom, o.Price(), o.ID)
		//	idxStore.Set(key, util.Uint64ToBytes(o.ID))
		//}
	})
}

// Check whether an asset even exists on the chain at the moment.
func (k Keeper) assetExists(ctx sdk.Context, asset sdk.Coin) bool {
	total := k.bk.GetSupply(ctx).GetTotal()
	return total.AmountOf(asset.Denom).GT(sdk.ZeroInt())
}

func (k *Keeper) CancelReplaceLimitOrder(ctx sdk.Context, newOrder types.Order, origClientOrderId string) (*sdk.Result, error) {
	// Use a fixed gas amount
	ctx.GasMeter().ConsumeGas(gasPriceCancelReplaceOrder, "CancelReplaceOrder")
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	origOrder := k.GetOrderByOwnerAndClientOrderId(ctx, newOrder.Owner, origClientOrderId)

	if origOrder == nil {
		return nil, sdkerrors.Wrap(types.ErrClientOrderIdNotFound, origClientOrderId)
	}

	// Verify that instrument is the same.
	if origOrder.Source.Denom != newOrder.Source.Denom || origOrder.Destination.Denom != newOrder.Destination.Denom {
		return nil, sdkerrors.Wrap(
			types.ErrOrderInstrumentChanged, fmt.Sprintf(
				"source %s != %s Or dest %s != %s", origOrder.Source, newOrder.Source,
				origOrder.Destination.Denom, newOrder.Destination.Denom,
			),
		)
	}

	// Has the previous order already achieved the goal on the source side?
	if origOrder.SourceFilled.GTE(newOrder.Source.Amount) {
		return nil, sdkerrors.Wrap(types.ErrNoSourceRemaining, "")
	}

	k.deleteOrder(ctx, origOrder)
	types.EmitExpireEvent(ctx, *origOrder)

	// Adjust remaining according to how much of the replaced order was filled:
	newOrder.SourceFilled = origOrder.SourceFilled
	newOrder.SourceRemaining = newOrder.Source.Amount.Sub(newOrder.SourceFilled)
	newOrder.DestinationFilled = origOrder.DestinationFilled

	newOrder.TimeInForce = origOrder.TimeInForce

	resAdd, err := k.NewOrderSingle(ctx, newOrder)
	if err != nil {
		return nil, err
	}

	evts := append(ctx.EventManager().ABCIEvents(), resAdd.Events...)
	return &sdk.Result{Events: evts}, nil
}

func (k *Keeper) GetOrderByOwnerAndClientOrderId(ctx sdk.Context, owner, clientOrderId string) *types.Order {
	store := ctx.KVStore(k.key)

	key := types.GetOwnerKey(owner, clientOrderId)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	o := &types.Order{}
	err := k.cdc.UnmarshalBinaryBare(bz, o)
	if err != nil {
		panic(err)
	}

	return o
}

func (k *Keeper) CancelOrder(ctx sdk.Context, owner sdk.AccAddress, clientOrderId string) (*sdk.Result, error) {
	// Use a fixed gas amount
	ctx.GasMeter().ConsumeGas(gasPriceCancelOrder, "CancelOrder")
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	// orders := k.accountOrders.GetAllOrders(owner)
	order := k.GetOrderByOwnerAndClientOrderId(ctx, owner.String(), clientOrderId)

	if order == nil {
		return nil, sdkerrors.Wrap(types.ErrClientOrderIdNotFound, clientOrderId)
	}

	types.EmitExpireEvent(ctx, *order)
	k.deleteOrder(ctx, order)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

// Update any orders that can no longer be filled with the account's balance.
func (k *Keeper) accountChanged(ctx sdk.Context, accounts []sdk.AccAddress) {
	for _, acc := range accounts {
		orders := k.GetOrdersByOwner(ctx, acc)
		for _, order := range orders {
			spendableCoins := k.bk.SpendableCoins(ctx, acc)
			denomBalance := spendableCoins.AmountOf(order.Source.Denom)

			origSourceRemaining := order.SourceRemaining
			order.SourceRemaining = order.Source.Amount.Sub(order.SourceFilled)
			order.SourceRemaining = sdk.MinInt(order.SourceRemaining, denomBalance)

			if order.SourceRemaining.IsZero() {
				types.EmitExpireEvent(ctx, *order)
				k.deleteOrder(ctx, order)
			} else if !origSourceRemaining.Equal(order.SourceRemaining) {
				types.EmitUpdateEvent(ctx, *order)
				k.setOrder(ctx, order)
			}
		}
	}
}

func (k Keeper) setOrder(ctx sdk.Context, order *types.Order) {
	var (
		store    = ctx.KVStore(k.key)
		idxStore = ctx.KVStore(k.keyIndices)
	)

	orderbz := k.cdc.MustMarshalBinaryBare(order)

	ownerKey := types.GetOwnerKey(order.Owner, order.ClientOrderID)
	store.Set(ownerKey, orderbz)

	priorityKey := types.GetPriorityKey(order.Source.Denom, order.Destination.Denom, order.Price(), order.ID)
	idxStore.Set(priorityKey, orderbz)
}

func (k Keeper) GetInstrument(ctx sdk.Context, src, dst string) *types.MarketData {
	idxStore := ctx.KVStore(k.keyIndices)

	key := types.GetMarketDataKey(src, dst)
	bz := idxStore.Get(key)
	if bz == nil {
		return nil
	}

	md := new(types.MarketData)
	k.cdc.MustUnmarshalBinaryBare(bz, md)
	return md
}

// Get instruments based on current order book. Does not include synthetic instruments.
func (k Keeper) GetInstruments(ctx sdk.Context) (instrs []types.MarketData) {
	idxStore := ctx.KVStore(k.keyIndices)

	it := idxStore.Iterator(types.GetMarketDataPrefix(), sdk.PrefixEndBytes(types.GetMarketDataPrefix()))
	defer it.Close()

	for ; it.Valid(); it.Next() {
		md := types.MarketData{}
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &md)

		// Amino appears to serialize nil *time.Time entries as Unix epoch. Convert to nil.
		if md.Timestamp != nil && md.Timestamp.Equal(time.Unix(0, 0)) {
			md.Timestamp = nil
		}

		instrs = append(instrs, md)
	}

	return
}

// GetAllInstruments gets all instruments based on current order book. It
// includes synthetic pairs, last order timestamp and calculates the best price
// for the pair.
func (k Keeper) GetAllInstruments(ctx sdk.Context) []*types.MarketData {
	coins := k.bk.GetSupply(ctx).GetTotal().Sort()
	n := len(coins)
	// n instruments producing n*(n-1) pairs below
	instrLst := make([]*types.MarketData, n*(n-1))
	idx := 0
	// produce cartesian products of denominations resulting in all
	// denominations paired with each other except for themselves
	for _, srcCoin := range coins {
		source := srcCoin.Denom
		for _, dstCoin := range coins {
			destination := dstCoin.Denom
			if source == destination {
				continue
			}

			instrLst[idx] = &types.MarketData{
				Source:      source,
				Destination: destination,
			}

			// fill in last order price, timestamp
			md := k.GetInstrument(ctx, source, destination)
			if md != nil && md.LastPrice != nil {
				instrLst[idx].LastPrice = md.LastPrice
				instrLst[idx].Timestamp = md.Timestamp
			}

			idx++
		}
	}

	return instrLst
}

func (k *Keeper) deleteOrder(ctx sdk.Context, order *types.Order) {
	var (
		store    = ctx.KVStore(k.key)
		idxStore = ctx.KVStore(k.keyIndices)
	)

	ownerKey := types.GetOwnerKey(order.Owner, order.ClientOrderID)
	store.Delete(ownerKey)

	priorityKey := types.GetPriorityKey(order.Source.Denom, order.Destination.Denom, order.Price(), order.ID)
	idxStore.Delete(priorityKey)
}

func (k Keeper) getBestOrder(ctx sdk.Context, src, dst string) *types.Order {
	idxStore := ctx.KVStore(k.keyIndices)
	key := types.GetPriorityKeyBySrcAndDst(src, dst)

	it := sdk.KVStorePrefixIterator(idxStore, key)
	defer it.Close()

	if it.Valid() {
		order := new(types.Order)
		k.cdc.MustUnmarshalBinaryBare(it.Value(), order)
		return order
	}

	return nil
}

// GetBestPrice returns the best priced passive order for source and
// destination instruments. Returns nil when executePlan cannot find a best
// plan.
func (k Keeper) GetBestPrice(ctx sdk.Context, source, destination string) *sdk.Dec {
	var bestPrice *sdk.Dec

	bestPlan := k.createExecutionPlan(ctx, destination, source)
	if !bestPlan.DestinationCapacity().IsZero() {
		bestPrice = &bestPlan.Price
	}

	return bestPrice
}

func (k Keeper) GetOrdersByOwner(ctx sdk.Context, owner sdk.AccAddress) (res []*types.Order) {
	store := ctx.KVStore(k.key)

	key := types.GetOwnerKey(owner.String(), "")
	it := sdk.KVStorePrefixIterator(store, key)
	defer it.Close()

	for ; it.Valid(); it.Next() {
		o := &types.Order{}
		err := k.cdc.UnmarshalBinaryBare(it.Value(), o)
		if err != nil {
			panic(err)
		}

		res = append(res, o)
	}

	return
}

func containsClientId(orders []*types.Order, clientOrderId string) bool {
	// TODO Orders are already ordered by ClientOrderId. Consider using a binary search.
	for _, order := range orders {
		if order.ClientOrderID == clientOrderId {
			return true
		}
	}
	return false
}

func getOrdersSourceDemand(orders []*types.Order, src, dst string) sdk.Coin {
	sumSourceRemaining := sdk.ZeroInt()
	for _, order := range orders {
		if order.Source.Denom != src || order.Destination.Denom != dst {
			continue
		}

		sumSourceRemaining = sumSourceRemaining.Add(order.SourceRemaining)
	}

	return sdk.NewCoin(src, sumSourceRemaining)
}

func (k Keeper) transferTradedAmounts(ctx sdk.Context, sourceFilled, destinationFilled sdk.Coin, passiveAccountAddr, aggressiveAccountAddr string) error {
	inputs := []banktypes.Input{
		{Address: aggressiveAccountAddr, Coins: sdk.NewCoins(sourceFilled)},
		{Address: passiveAccountAddr, Coins: sdk.NewCoins(destinationFilled)},
	}

	outputs := []banktypes.Output{
		{Address: aggressiveAccountAddr, Coins: sdk.NewCoins(destinationFilled)},
		{Address: passiveAccountAddr, Coins: sdk.NewCoins(sourceFilled)},
	}

	return k.bk.InputOutputCoins(ctx, inputs, outputs)
}

func (k Keeper) getNextOrderNumber(ctx sdk.Context) uint64 {
	var orderID uint64
	store := ctx.KVStore(k.key)
	bz := store.Get(types.GetOrderIDGeneratorKey())
	if bz == nil {
		orderID = 0
	} else {
		orderID = sdk.BigEndianToUint64(bz)
	}

	bz = sdk.Uint64ToBigEndian(orderID + 1)
	store.Set(types.GetOrderIDGeneratorKey(), bz)
	return orderID
}

func (k Keeper) registerMarketData(ctx sdk.Context, src, dst string) {
	idxStore := ctx.KVStore(k.keyIndices)

	key := types.GetMarketDataKey(src, dst)

	if idxStore.Has(key) {
		return
	}

	md := types.MarketData{
		Source:      src,
		Destination: dst,
	}

	bz := k.cdc.MustMarshalBinaryBare(&md)
	idxStore.Set(key, bz)
}

// Register successful trade execution
func (k Keeper) setMarketData(ctx sdk.Context, src, dst string, price sdk.Dec) {
	idxStore := ctx.KVStore(k.keyIndices)
	timestamp := ctx.BlockTime()

	md := types.MarketData{Source: src, Destination: dst, LastPrice: &price, Timestamp: &timestamp}
	key := types.GetMarketDataKey(src, dst)

	bz := k.cdc.MustMarshalBinaryBare(&md)
	idxStore.Set(key, bz)
}
