// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"math"
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authe "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/bank"

	emtypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/market/types"
)

const (
	// Gas prices must be predictable, and not depend on the number of passive orders matched.
	gasPriceNewOrder           = uint64(25000)
	gasPriceCancelReplaceOrder = uint64(25000)
	gasPriceCancelOrder        = uint64(12500)
)

type Keeper struct {
	key         sdk.StoreKey
	cdc         *codec.Codec
	instruments types.Instruments
	ak          types.AccountKeeper
	bk          types.BankKeeper
	sk          types.SupplyKeeper
	authorityk  types.RestrictedKeeper

	accountOrders types.Orders
	appstateInit  *sync.Once

	restrictedDenoms emtypes.RestrictedDenoms
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, authKeeper types.AccountKeeper, bankKeeper types.BankKeeper, supplyKeeper types.SupplyKeeper, authorityKeeper types.RestrictedKeeper) *Keeper {
	k := &Keeper{
		cdc: cdc,
		key: key,
		ak:  authKeeper,
		bk:  bankKeeper,
		sk:  supplyKeeper,

		accountOrders: types.NewOrders(),
		appstateInit:  new(sync.Once),

		authorityk: authorityKeeper,
	}

	authKeeper.AddAccountListener(k.accountChanged)
	return k
}

func (k *Keeper) createExecutionPlan(SourceDenom string, DestinationDenom string) types.ExecutionPlan {
	result := types.ExecutionPlan{
		Price: sdk.NewDec(math.MaxInt64),
	}

	// Find the best direct or synthetic price for a given source/destination denom
	for _, firstInstrument := range k.instruments {
		if firstInstrument.Source == SourceDenom {
			if firstInstrument.Orders.Empty() {
				continue
			}

			firstPassiveOrder := firstInstrument.Orders.LeftKey().(*types.Order)

			// Check direct price
			if firstInstrument.Destination == DestinationDenom {
				// Direct price is better than current plan

				planPrice := sdk.OneDec().Quo(firstPassiveOrder.Price())
				planPrice = planPrice.Add(sdk.NewDecWithPrec(1, sdk.Precision)) // Add floating point epsilon
				if planPrice.LT(result.Price) {
					result = types.ExecutionPlan{
						Price:      planPrice,
						FirstOrder: firstPassiveOrder,
					}
				}
			}

			// Check synthetic price by going through two orders:
			// (SourceDenom, X) -> (X, DestinationDenom)
			for _, secondInstrument := range k.instruments {
				if secondInstrument.Source != firstInstrument.Destination {
					continue
				}

				if secondInstrument.Destination != DestinationDenom {
					continue
				}

				if secondInstrument.Orders.Empty() {
					continue
				}

				secondPassiveOrder := secondInstrument.Orders.LeftKey().(*types.Order)

				planPrice := sdk.OneDec().Quo(firstPassiveOrder.Price().Mul(secondPassiveOrder.Price()))
				planPrice = planPrice.Add(sdk.NewDecWithPrec(1, sdk.Precision)) // Add floating point epsilon

				if planPrice.LT(result.Price) {
					result = types.ExecutionPlan{
						Price:       planPrice,
						FirstOrder:  firstPassiveOrder,
						SecondOrder: secondPassiveOrder,
					}
				}
			}
		}
	}

	return result
}

func (k *Keeper) NewOrderSingle(ctx sdk.Context, aggressiveOrder types.Order) (*sdk.Result, error) {
	// Use a fixed gas amount
	ctx.GasMeter().ConsumeGas(gasPriceNewOrder, "NewOrderSingle")
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	if err := aggressiveOrder.IsValid(); err != nil {
		return nil, err
	}

	if aggressiveOrder.IsFilled() {
		return nil, sdkerrors.Wrapf(types.ErrInvalidPrice, "Order price is invalid: %s -> %s", aggressiveOrder.Source, aggressiveOrder.Destination)
	}

	sourceAccount := k.ak.GetAccount(ctx, aggressiveOrder.Owner)
	if sourceAccount == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", aggressiveOrder.Owner.String())
		//return sdk.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", aggressiveOrder.Owner.String())).Result()
	}

	// Verify account balance
	if _, anyNegative := sourceAccount.SpendableCoins(ctx.BlockTime()).SafeSub(sdk.NewCoins(aggressiveOrder.Source)); anyNegative {
		return nil, sdkerrors.Wrapf(
			types.ErrAccountBalanceInsufficient,
			"Account %v has insufficient balance to execute trade: %v < %v",
			aggressiveOrder.Owner,
			sourceAccount.SpendableCoins(ctx.BlockTime()),
			aggressiveOrder.Source,
		)
	}

	// Ensure that the market is not showing "phantom liquidity" by rejecting multiple orders in an instrument based on the same balance.
	totalSourceDemand := k.accountOrders.GetAccountSourceDemand(aggressiveOrder.Owner, aggressiveOrder.Source.Denom, aggressiveOrder.Destination.Denom)
	totalSourceDemand = totalSourceDemand.Add(aggressiveOrder.Source)
	if _, anyNegative := sourceAccount.SpendableCoins(ctx.BlockTime()).SafeSub(sdk.NewCoins(totalSourceDemand)); anyNegative {
		// TODO Improve message
		return nil, sdkerrors.Wrapf(types.ErrAccountBalanceInsufficientForInstrument, "")
	}

	// Verify uniqueness of client order id among active orders
	if k.accountOrders.ContainsClientOrderId(aggressiveOrder.Owner, aggressiveOrder.ClientOrderID) {
		return nil, sdkerrors.Wrap(types.ErrNonUniqueClientOrderId, aggressiveOrder.ClientOrderID)
	}

	// Verify that the destination asset actually exists on chain before creating an instrument
	if !k.assetExists(ctx, aggressiveOrder.Destination) {
		return nil, sdkerrors.Wrap(types.ErrUnknownAsset, aggressiveOrder.Destination.Denom)
	}

	aggressiveOrder.ID = k.getNextOrderNumber(ctx)

	types.EmitNewOrderEvent(ctx, aggressiveOrder)

	for {
		plan := k.createExecutionPlan(aggressiveOrder.Destination.Denom, aggressiveOrder.Source.Denom)
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

		for _, passiveOrder := range []*types.Order{plan.SecondOrder, plan.FirstOrder} {
			if passiveOrder == nil {
				continue
			}

			// Use the passive order's price in the market.
			stepSourceFilled := stepDestinationFilled.Quo(passiveOrder.Price())

			// Update the aggressive order during the plan's final step.
			if passiveOrder.Destination.Denom == aggressiveOrder.Source.Denom {
				aggressiveOrder.SourceRemaining = aggressiveOrder.SourceRemaining.Sub(stepDestinationFilled.RoundInt())
				aggressiveOrder.SourceFilled = aggressiveOrder.SourceFilled.Add(stepDestinationFilled.RoundInt())

				// Invariant check
				if aggressiveOrder.SourceRemaining.LT(sdk.ZeroInt()) {
					panic(fmt.Sprintf("Aggressive order's SourceRemaining field is less than zero. order: %v", aggressiveOrder))
				}
			}

			if passiveOrder.Source.Denom == aggressiveOrder.Destination.Denom {
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

			if passiveOrder.IsFilled() {
				types.EmitFilledEvent(ctx, *passiveOrder)
				// Order has been filled. Remove it from queue.
				k.deleteOrder(ctx, passiveOrder)
			} else {
				types.EmitPartiallyFilledEvent(ctx, *passiveOrder)
				k.setOrder(ctx, passiveOrder)
			}

			stepDestinationFilled = stepSourceFilled
		}

		if aggressiveOrder.IsFilled() {
			types.EmitFilledEvent(ctx, aggressiveOrder)
			// Order has been filled.
			break
		}

		types.EmitPartiallyFilledEvent(ctx, aggressiveOrder)
	}

	// Order was not fully matched. Add to book unless restricted.
	if !aggressiveOrder.IsFilled() {
		// Check whether this denomination is restricted and thus cannot create passive orders
		addToBook := true
		if denom, found := k.restrictedDenoms.Find(aggressiveOrder.Source.Denom); found {
			addToBook = denom.IsAnyAllowed(aggressiveOrder.Owner)
		}

		if denom, found := k.restrictedDenoms.Find(aggressiveOrder.Destination.Denom); addToBook && found {
			addToBook = denom.IsAnyAllowed(aggressiveOrder.Owner)
		}

		if addToBook {
			op := &aggressiveOrder
			k.instruments.InsertOrder(op)
			k.accountOrders.AddOrder(op)
			k.setOrder(ctx, op)
			// NOTE This should be the only place that an order is added to the book!
			// NOTE If this ceases to be true, move logic to func that cleans up all datastructures.
		}
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func (k *Keeper) initializeFromStore(ctx sdk.Context) {
	k.appstateInit.Do(func() {
		// Load the restricted denominations from the authority module
		k.restrictedDenoms = k.authorityk.GetRestrictedDenoms(ctx)

		// Load the last known market state from app state.
		store := ctx.KVStore(k.key)
		it := store.Iterator(types.GetOrderKey(0), types.GetOrderKey(math.MaxUint64))
		for ; it.Valid(); it.Next() {
			o := &types.Order{}
			err := k.cdc.UnmarshalBinaryBare(it.Value(), o)
			if err != nil {
				panic(err)
			}

			k.instruments.InsertOrder(o)
			k.accountOrders.AddOrder(o)
		}
	})
}

// Check whether an asset even exists on the chain at the moment.
func (k Keeper) assetExists(ctx sdk.Context, asset sdk.Coin) bool {
	total := k.sk.GetSupply(ctx).GetTotal()
	return total.AmountOf(asset.Denom).GT(sdk.ZeroInt())
}

func (k *Keeper) CancelReplaceOrder(ctx sdk.Context, newOrder types.Order, origClientOrderId string) (*sdk.Result, error) {
	// Use a fixed gas amount
	ctx.GasMeter().ConsumeGas(gasPriceCancelReplaceOrder, "CancelReplaceOrder")
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	origOrder := k.accountOrders.GetOrder(newOrder.Owner, origClientOrderId)
	if origOrder == nil {
		return nil, sdkerrors.Wrap(types.ErrClientOrderIdNotFound, origClientOrderId)
	}

	// Verify that instrument is the same.
	if origOrder.Source.Denom != newOrder.Source.Denom || origOrder.Destination.Denom != newOrder.Destination.Denom {
		return nil, sdkerrors.Wrap(types.ErrOrderInstrumentChanged, "")
	}

	// Has the previous order already achieved the goal on the source side?
	if origOrder.SourceFilled.GTE(newOrder.Source.Amount) {
		return nil, sdkerrors.Wrap(types.ErrNoSourceRemaining, "")
	}

	resCancel, err := k.CancelOrder(ctx, newOrder.Owner, origClientOrderId)
	if err != nil {
		return nil, err
	}

	// Adjust remaining according to how much of the replaced order was filled:
	newOrder.SourceFilled = origOrder.SourceFilled
	newOrder.SourceRemaining = newOrder.Source.Amount.Sub(newOrder.SourceFilled)
	newOrder.DestinationFilled = origOrder.DestinationFilled

	resAdd, err := k.NewOrderSingle(ctx, newOrder)
	if err != nil {
		return nil, err
	}

	evts := append(ctx.EventManager().Events(), resCancel.Events...)
	evts = append(evts, resAdd.Events...)
	return &sdk.Result{Events: evts}, nil
}

func (k *Keeper) CancelOrder(ctx sdk.Context, owner sdk.AccAddress, clientOrderId string) (*sdk.Result, error) {
	// Use a fixed gas amount
	ctx.GasMeter().ConsumeGas(gasPriceCancelOrder, "CancelOrder")
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	orders := k.accountOrders.GetAllOrders(owner)

	var order *types.Order
	i, _ := orders.Find(func(index int, v interface{}) bool {
		order = v.(*types.Order)
		return order.ClientOrderID == clientOrderId
	})

	if i == -1 {
		return nil, sdkerrors.Wrap(types.ErrClientOrderIdNotFound, clientOrderId)
	}

	k.deleteOrder(ctx, order)
	types.EmitCancelEvent(ctx, *order)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// Update any orders that can no longer be filled with the account's balance.
func (k *Keeper) accountChanged(ctx sdk.Context, acc authe.Account) {
	orders := k.accountOrders.GetAllOrders(acc.GetAddress())

	orders.Each(func(_ int, v interface{}) {
		order := v.(*types.Order)
		denomBalance := acc.SpendableCoins(ctx.BlockTime()).AmountOf(order.Source.Denom)

		order.SourceRemaining = order.Source.Amount.Sub(order.SourceFilled)
		order.SourceRemaining = sdk.MinInt(order.SourceRemaining, denomBalance)

		if order.SourceRemaining.IsZero() {
			k.deleteOrder(ctx, order)
		}
	})
}

func (k Keeper) setOrder(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(k.key)
	bz := k.cdc.MustMarshalBinaryBare(order)
	store.Set(types.GetOrderKey(order.ID), bz)
}

func (k *Keeper) deleteOrder(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(k.key)
	store.Delete(types.GetOrderKey(order.ID))

	k.accountOrders.RemoveOrder(order)

	instrument := k.instruments.GetInstrument(order.Source.Denom, order.Destination.Denom)
	if instrument == nil {
		return
	}

	instrument.Orders.Remove(order)
	if instrument.Orders.Empty() {
		k.instruments.RemoveInstrument(*instrument)
	}
}

func (k Keeper) GetOrdersByOwner(owner sdk.AccAddress) []types.Order {
	orders := k.accountOrders.GetAllOrders(owner)
	res := make([]types.Order, orders.Size())

	it := orders.Iterator()
	for it.Next() {
		o := it.Value().(*types.Order)
		res[it.Index()] = *o // Copy all orders, so that the calling function can't modify state.
	}

	return res
}

func (k Keeper) transferTradedAmounts(ctx sdk.Context, sourceFilled, destinationFilled sdk.Coin, passiveAccountAddr, aggressiveAccountAddr sdk.AccAddress) error {
	inputs := []bank.Input{
		{aggressiveAccountAddr, sdk.NewCoins(sourceFilled)},
		{passiveAccountAddr, sdk.NewCoins(destinationFilled)},
	}

	outputs := []bank.Output{
		{aggressiveAccountAddr, sdk.NewCoins(destinationFilled)},
		{passiveAccountAddr, sdk.NewCoins(sourceFilled)},
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
		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &orderID)
		if err != nil {
			panic(err)
		}
	}

	bz = k.cdc.MustMarshalBinaryLengthPrefixed(orderID + 1)
	store.Set(types.GetOrderIDGeneratorKey(), bz)

	return orderID
}
