// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"fmt"
	"math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authe "github.com/cosmos/cosmos-sdk/x/auth/exported"

	"github.com/e-money/em-ledger/x/market/types"
)

type Keeper struct {
	key         sdk.StoreKey
	cdc         *codec.Codec
	instruments types.Instruments
	ak          types.AccountKeeper
	bk          types.BankKeeper

	accountOrders types.Orders
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, authKeeper types.AccountKeeper, bankKeeper types.BankKeeper) *Keeper {
	k := &Keeper{
		cdc: cdc,
		key: key,
		ak:  authKeeper,
		bk:  bankKeeper,

		accountOrders: types.NewOrders(),
	}

	authKeeper.AddAccountListener(k.accountChanged)
	return k
}

func (k *Keeper) NewOrderSingle(ctx sdk.Context, aggressiveOrder *types.Order) sdk.Result {
	if aggressiveOrder.Source.Denom == aggressiveOrder.Destination.Denom {
		return types.ErrInvalidInstrument(aggressiveOrder.Source.Denom, aggressiveOrder.Destination.Denom).Result()
	}

	sourceAccount := k.ak.GetAccount(ctx, aggressiveOrder.Owner)
	if sourceAccount == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", aggressiveOrder.Owner.String())).Result()
	}

	// Verify account balance
	if _, anyNegative := sourceAccount.GetCoins().SafeSub(sdk.NewCoins(aggressiveOrder.Source)); anyNegative {
		return types.ErrAccountBalanceInsufficient(aggressiveOrder.Owner, aggressiveOrder.Source, sourceAccount.GetCoins().AmountOf(aggressiveOrder.Source.Denom)).Result()
	}

	// Verify uniqueness of client order id among active orders
	if k.accountOrders.ContainsClientOrderId(aggressiveOrder.Owner, aggressiveOrder.ClientOrderID) {
		return types.ErrNonUniqueClientOrderId(aggressiveOrder.Owner, aggressiveOrder.ClientOrderID).Result()
	}

	aggressiveOrder.ID = k.getNextOrderNumber(ctx)

	for _, i := range k.instruments {
		if i.Source == aggressiveOrder.Destination.Denom && i.Destination == aggressiveOrder.Source.Denom {
			for {
				if i.Orders.Empty() {
					break
				}

				passiveOrder := i.Orders.LeftKey().(*types.Order)
				if aggressiveOrder.Price().GT(passiveOrder.InvertedPrice()) {
					// Spread has not been crossed. Candidate should be added to order book.
					break
				}

				// Price is divided evenly between bid and offer. Price improvement is shared equally.
				//matchingPrice := aggressiveOrder.Destination.Amount.Add(passiveOrder.Source.Amount).ToDec().Quo(aggressiveOrder.Source.Amount.Add(passiveOrder.Destination.Amount).ToDec())

				// Divide evenly between two prices:
				//matchingPrice := passiveOrder.InvertedPrice().Add(aggressiveOrder.Price()).Quo(sdk.NewDec(2))

				// Use the passive order's price in the market.
				matchingPrice := passiveOrder.InvertedPrice()

				//The number of tokens from the aggressive orders sell side that the passive order will fill with the price that has been reached
				aggressiveSourceMatched := passiveOrder.SourceRemaining.ToDec().QuoRoundUp(matchingPrice).TruncateInt()

				if aggressiveOrder.SourceRemaining.LT(aggressiveSourceMatched) {
					aggressiveSourceMatched = aggressiveOrder.SourceRemaining
				}

				aggressiveDestinationMatched := aggressiveSourceMatched.ToDec().Mul(matchingPrice).Ceil().TruncateInt()

				// Adjust orders
				passiveOrder.SourceFilled = passiveOrder.SourceFilled.Add(aggressiveDestinationMatched)
				passiveOrder.SourceRemaining = passiveOrder.SourceRemaining.Sub(aggressiveDestinationMatched)
				aggressiveOrder.SourceFilled = aggressiveOrder.SourceFilled.Add(aggressiveSourceMatched)
				aggressiveOrder.SourceRemaining = aggressiveOrder.SourceRemaining.Sub(aggressiveSourceMatched)

				err := k.transferTradedAmounts(ctx, aggressiveDestinationMatched, aggressiveSourceMatched, passiveOrder, aggressiveOrder)
				if err != nil {
					return err.Result()
				}

				// Invariant check
				if aggressiveOrder.SourceRemaining.LT(sdk.ZeroInt()) || passiveOrder.SourceRemaining.LT(sdk.ZeroInt()) {
					msg := fmt.Sprintf("Remaining field is less than zero. order: %v candidate: %v", aggressiveOrder, passiveOrder)
					panic(msg)
				}

				if passiveOrder.SourceRemaining.IsZero() {
					// Order has been filled. Remove it from queue.
					k.deleteOrder(ctx, passiveOrder)
				} else {
					k.setOrder(ctx, passiveOrder)
				}

				if aggressiveOrder.SourceRemaining.IsZero() {
					// Order has been filled.
					break
				}
			}
		}
	}

	if !aggressiveOrder.SourceRemaining.IsZero() {
		// Order was not fully matched. Add to book.
		k.instruments.InsertOrder(aggressiveOrder)
		k.accountOrders.AddOrder(aggressiveOrder)
		k.setOrder(ctx, aggressiveOrder)
		// NOTE This should be the only place that an order is added to the book!
		// NOTE If this ceases to be true, move logic to func that cleans up all datastructures.
	}

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func (k *Keeper) initializeFromStore(ctx sdk.Context, key sdk.StoreKey) {
	store := ctx.KVStore(key)
	it := store.Iterator(types.GetOrderKey(0), types.GetOrderKey(math.MaxUint64))
	for ; it.Valid(); it.Next() {
		o := types.Order{}
		err := k.cdc.UnmarshalBinaryBare(it.Value(), &o)
		if err != nil {
			panic(err)
		}

		k.instruments.InsertOrder(&o)
		k.accountOrders.AddOrder(&o)
	}
}

func (k *Keeper) CancelReplaceOrder(ctx sdk.Context, newOrder *types.Order, origClientOrderId string) sdk.Result {
	origOrder := k.accountOrders.GetOrder(newOrder.Owner, origClientOrderId)
	if origOrder == nil {
		return types.ErrClientOrderIdNotFound(newOrder.Owner, origClientOrderId).Result()
	}

	// Verify that instrument is the same.
	if origOrder.Source.Denom != newOrder.Source.Denom || origOrder.Destination.Denom != newOrder.Destination.Denom {
		return types.ErrOrderInstrumentChanged(
			origOrder.Source.Denom, origOrder.Destination.Denom,
			newOrder.Source.Denom, newOrder.Destination.Denom,
		).Result()
	}

	resCancel := k.CancelOrder(ctx, newOrder.Owner, origClientOrderId)
	if !resCancel.IsOK() {
		return resCancel
	}

	// Adjust remaining according to how much of the replaced order was filled:
	newOrder.SourceFilled = origOrder.SourceFilled
	newOrder.SourceRemaining = newOrder.Source.Amount.Sub(origOrder.SourceFilled)

	resAdd := k.NewOrderSingle(ctx, newOrder)
	if !resAdd.IsOK() {
		resAdd.Events = append(resAdd.Events, resCancel.Events...)
		return resAdd
	}

	evts := append(ctx.EventManager().Events(), resCancel.Events...)
	evts = append(evts, resAdd.Events...)
	return sdk.Result{Events: evts}
}

func (k *Keeper) CancelOrder(ctx sdk.Context, owner sdk.AccAddress, clientOrderId string) sdk.Result {
	orders := k.accountOrders.GetAllOrders(owner)

	var order *types.Order
	i, _ := orders.Find(func(index int, v interface{}) bool {
		order = v.(*types.Order)
		return order.ClientOrderID == clientOrderId
	})

	if i == -1 {
		return types.ErrClientOrderIdNotFound(owner, clientOrderId).Result()
	}

	k.deleteOrder(ctx, order)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

// Update any orders that can no longer be filled with the account's balance.
func (k *Keeper) accountChanged(ctx sdk.Context, acc authe.Account) {
	orders := k.accountOrders.GetAllOrders(acc.GetAddress())

	orders.Each(func(_ int, v interface{}) {
		order := v.(*types.Order)
		denomBalance := acc.GetCoins().AmountOf(order.Source.Denom)

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

func (k Keeper) transferTradedAmounts(ctx sdk.Context, destinationMatched, sourceMatched sdk.Int, passiveOrder, aggressiveOrder *types.Order) sdk.Error {
	var (
		passiveAccount    = k.ak.GetAccount(ctx, passiveOrder.Owner)
		aggressiveAccount = k.ak.GetAccount(ctx, aggressiveOrder.Owner)
	)

	// Verify that the passive order still holds the balance
	coinMatchedDst := sdk.NewCoin(passiveOrder.Source.Denom, destinationMatched)
	if _, anyNegative := passiveAccount.GetCoins().SafeSub(sdk.NewCoins(coinMatchedDst)); anyNegative {
		return types.ErrAccountBalanceInsufficient(passiveAccount.GetAddress(), coinMatchedDst, passiveAccount.GetCoins().AmountOf(coinMatchedDst.Denom))
	}

	// Verify that the aggressive order still holds the balance
	coinMatchedSrc := sdk.NewCoin(aggressiveOrder.Source.Denom, sourceMatched)
	if _, anyNegative := aggressiveAccount.GetCoins().SafeSub(sdk.NewCoins(coinMatchedSrc)); anyNegative {
		return types.ErrAccountBalanceInsufficient(aggressiveAccount.GetAddress(), coinMatchedSrc, aggressiveAccount.GetCoins().AmountOf(coinMatchedSrc.Denom))
	}

	// Balances appear sufficient. Do the transfers
	err := k.bk.SendCoins(ctx, aggressiveOrder.Owner, passiveOrder.Owner, sdk.NewCoins(coinMatchedSrc))
	if err != nil {
		return err
	}

	err = k.bk.SendCoins(ctx, passiveOrder.Owner, aggressiveOrder.Owner, sdk.NewCoins(coinMatchedDst))
	if err != nil {
		// TODO Reverse the successful send?
		return err
	}

	return nil
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
