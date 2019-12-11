// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/e-money/em-ledger/x/offer/types"
)

type Keeper struct {
	key         sdk.StoreKey
	cdc         *codec.Codec
	instruments types.Instruments
	ak          auth.AccountKeeper
	bk          bank.BaseKeeper

	accountOrders types.Orders
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, authKeeper auth.AccountKeeper, bankKeeper bank.BaseKeeper) Keeper {
	return Keeper{
		cdc: cdc,
		key: key,
		ak:  authKeeper,
		bk:  bankKeeper,

		accountOrders: types.NewOrders(),
	}
}

func (k *Keeper) NewOrderSingle(ctx sdk.Context, aggressiveOrder *types.Order) sdk.Result {
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
		return types.ErrNonUniqueClientOrderID(aggressiveOrder.Owner, aggressiveOrder.ClientOrderID).Result()
	}

	//if clientOrders := k.accountOrders[aggressiveOrder.Owner.String()]; clientOrders != nil {
	//	if clientOrders.Contains(aggressiveOrder) {
	//		return types.ErrNonUniqueClientOrderID(aggressiveOrder.Owner, aggressiveOrder.ClientOrderID).Result()
	//	}
	//}

	aggressiveOrder.ID = k.getNextOrderNumber(ctx)

	for _, i := range k.instruments {
		if i.Source == aggressiveOrder.Destination.Denom && i.Destination == aggressiveOrder.Source.Denom {
			for {
				if i.Orders.Empty() {
					break
				}

				passiveOrder := i.Orders.LeftKey().(*types.Order)
				if aggressiveOrder.Price() > passiveOrder.InvertedPrice() {
					// Spread has not been crossed. Candidate should be added to order book.
					break
				}

				// Price is divided evenly between bid and offer. Price improvement is shared equally.
				matchingPrice := aggressiveOrder.Destination.Amount.Add(passiveOrder.Source.Amount).ToDec().Quo(aggressiveOrder.Source.Amount.Add(passiveOrder.Destination.Amount).ToDec())

				// Price improvement is 100% given to the buyer.
				//matchingPrice := co.invertedPrice

				sourceMatched := passiveOrder.SourceRemaining.ToDec().QuoRoundUp(matchingPrice).TruncateInt()
				if aggressiveOrder.SourceRemaining.LT(sourceMatched) {
					sourceMatched = aggressiveOrder.SourceRemaining
				}

				destinationMatched := sourceMatched.ToDec().Mul(matchingPrice).Ceil().TruncateInt()

				err := k.transferTradedAmounts(ctx, destinationMatched, sourceMatched, passiveOrder, aggressiveOrder)
				if err != nil {
					return err.Result()
				}

				// Adjust orders
				passiveOrder.SourceRemaining = passiveOrder.SourceRemaining.Sub(destinationMatched)
				aggressiveOrder.SourceRemaining = aggressiveOrder.SourceRemaining.Sub(sourceMatched)

				// Invariant check
				if aggressiveOrder.SourceRemaining.LT(sdk.ZeroInt()) || passiveOrder.SourceRemaining.LT(sdk.ZeroInt()) {
					msg := fmt.Sprintf("Remaining field is less than zero. order: %v candidate: %v", aggressiveOrder, passiveOrder)
					panic(msg)
				}

				if passiveOrder.SourceRemaining.IsZero() {
					// Order has been filled. Remove it from queue.
					k.deleteOrder(passiveOrder)

					if i.Orders.Empty() {
						k.instruments.RemoveInstrument(i)
					}
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
		// NOTE This should be the only place that an order is added to the book!
		// NOTE If this ceases to be true, move logic to func that cleans up all datastructures.
	}

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func (k *Keeper) CancelReplaceOrder(ctx sdk.Context, newOrder types.Order, origClientOrderId string) sdk.Result {
	// TODO Verify that instrument is the same.
	//currentOrders := k.accountOrders[newOrder.Owner.String()]

	res := k.CancelOrder(ctx, newOrder.Owner, origClientOrderId)
	if !res.IsOK() {
		return res
	}

	// TODO Add new order
	// Adjust remaining according to how much of the replaced order was filled:
	// newOrder.remaining = newOrder.sourceAmount - (oldOrder.SourceAmount - oldOrder.Remaining)

	evts := append(ctx.EventManager().Events(), res.Events...)
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
		return types.ErrClientOrderIDNotFound(owner, clientOrderId).Result()
	}

	k.deleteOrder(order)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func (k *Keeper) deleteOrder(order *types.Order) {
	instrument := k.instruments.GetInstrument(order.Source.Denom, order.Destination.Denom)
	instrument.Orders.Remove(order)
	if instrument.Orders.Empty() {
		k.instruments.RemoveInstrument(*instrument)
	}

	k.accountOrders.RemoveOrder(order)
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
	k.bk.SendCoins(ctx, aggressiveOrder.Owner, passiveOrder.Owner, sdk.NewCoins(coinMatchedSrc))
	k.bk.SendCoins(ctx, passiveOrder.Owner, aggressiveOrder.Owner, sdk.NewCoins(coinMatchedDst))
	return nil
}

func (k Keeper) getNextOrderNumber(ctx sdk.Context) uint64 {
	var orderID uint64
	store := ctx.KVStore(k.key)
	bz := store.Get(types.GlobalOrderIDKey)
	if bz == nil {
		orderID = 0
	} else {
		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &orderID)
		if err != nil {
			panic(err)
		}
	}

	bz = k.cdc.MustMarshalBinaryLengthPrefixed(orderID + 1)
	store.Set(types.GlobalOrderIDKey, bz)

	return orderID
}
