package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/emirpasic/gods/sets/treeset"
	"github.com/emirpasic/gods/utils"

	"github.com/e-money/em-ledger/x/offer/types"
)

type Keeper struct {
	key         sdk.StoreKey
	cdc         *codec.Codec
	instruments types.Instruments
	ak          auth.AccountKeeper
	bk          bank.BaseKeeper

	accountOrders map[string]*treeset.Set
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, authKeeper auth.AccountKeeper, bankKeeper bank.BaseKeeper) Keeper {
	return Keeper{
		cdc: cdc,
		key: key,
		ak:  authKeeper,
		bk:  bankKeeper,

		accountOrders: make(map[string]*treeset.Set),
	}
}

func (k *Keeper) ProcessOrder(ctx sdk.Context, aggressiveOrder *types.Order) sdk.Result {
	sourceAccount := k.ak.GetAccount(ctx, aggressiveOrder.SourceAccount)
	if sourceAccount == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", aggressiveOrder.SourceAccount.String())).Result()
	}

	// Verify account balance
	if _, anyNegative := sourceAccount.GetCoins().SafeSub(sdk.NewCoins(aggressiveOrder.Source)); anyNegative {
		return types.ErrNonUniqueClientOrderID(aggressiveOrder.Owner, aggressiveOrder.ClientOrderID).Result()
	}

	// Verify uniqueness of client order id among active orders
	if clientOrders := k.accountOrders[aggressiveOrder.Owner.String()]; clientOrders != nil {
		if clientOrders.Contains(aggressiveOrder) {
			// TODO Add error
			// TODO Add unit test
			return sdk.ErrUnknownAddress(fmt.Sprintf("Duplicate client order id")).Result()
		}
	}

	aggressiveOrder.ID = k.GetNextOrderNumber(ctx)

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
					//i.Orders.Remove(passiveOrder)
					k.DeleteOrder(passiveOrder)

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
		k.instruments.InsertOrder(aggressiveOrder)

		clientOrders := k.accountOrders[aggressiveOrder.Owner.String()]
		if clientOrders == nil {
			clientOrders = treeset.NewWith(OrderClientIdComparator)
			k.accountOrders[aggressiveOrder.Owner.String()] = clientOrders
		}

		clientOrders.Add(aggressiveOrder)
	}

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func (k Keeper) DeleteOrder(order *types.Order) {
	instrument := k.instruments.GetInstrument(order.Source.Denom, order.Destination.Denom)
	instrument.Orders.Remove(order)

	orders := k.accountOrders[order.Owner.String()]
	orders.Remove(order)
}

func (k Keeper) transferTradedAmounts(ctx sdk.Context, destinationMatched, sourceMatched sdk.Int, passiveOrder, aggressiveOrder *types.Order) sdk.Error {
	var (
		passiveAccount    = k.ak.GetAccount(ctx, passiveOrder.SourceAccount)
		aggressiveAccount = k.ak.GetAccount(ctx, aggressiveOrder.SourceAccount)
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
	k.bk.SendCoins(ctx, aggressiveOrder.SourceAccount, passiveOrder.SourceAccount, sdk.NewCoins(coinMatchedSrc))
	k.bk.SendCoins(ctx, passiveOrder.SourceAccount, aggressiveOrder.SourceAccount, sdk.NewCoins(coinMatchedDst))
	return nil
}

func (k Keeper) GetNextOrderNumber(ctx sdk.Context) uint64 {
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

func OrderClientIdComparator(a, b interface{}) int {
	aAsserted := a.(*types.Order)
	bAsserted := b.(*types.Order)

	return utils.StringComparator(aAsserted.ClientOrderID, bAsserted.ClientOrderID)
}
