package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/offer/types"
)

type Keeper struct {
	key         sdk.StoreKey
	cdc         *codec.Codec
	instruments types.Instruments
	ak          auth.AccountKeeper
	bk          bank.BaseKeeper
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, authKeeper auth.AccountKeeper, bankKeeper bank.BaseKeeper) Keeper {
	return Keeper{
		cdc: cdc,
		key: key,
		ak:  authKeeper,
		bk:  bankKeeper,
	}
}

func (k *Keeper) ProcessOrder(ctx sdk.Context, order *types.Order) sdk.Result {
	//acc := k.ak.GetAccount(ctx, order.SourceAccount)
	//// Verify account balance
	//acc.GetCoins().SafeSub()

	order.ID = k.GetNextOrderNumber(ctx)

	for _, i := range k.instruments {
		if i.Source == order.Destination && i.Destination == order.Source {
			for {
				if i.Orders.Len() == 0 {
					k.instruments.RemoveInstrument(i)
					break
				}

				co := i.Orders.Peek().(*types.Order)
				if order.Price() > co.InvertedPrice() {
					fmt.Printf("Price mismatch: %v > %v\n", order.Price(), co.InvertedPrice())
					break
				}

				// Price is divided evenly between bid and offer. Price improvement is shared equally.
				matchingPrice := (float64(order.DestinationAmount) + float64(co.SourceAmount)) / (float64(order.SourceAmount) + float64(co.DestinationAmount))

				// Price improvement is 100% given to the buyer.
				//matchingPrice := co.invertedPrice

				sourceMatched := uint(math.Ceil(float64(co.RemainingAmount) / matchingPrice))
				if order.RemainingAmount < sourceMatched {
					sourceMatched = order.RemainingAmount
				}

				destinationMatched := uint(math.Floor(float64(sourceMatched) * matchingPrice))

				co.RemainingAmount = co.RemainingAmount - destinationMatched
				order.RemainingAmount = order.RemainingAmount - sourceMatched

				// Invariant check
				if order.RemainingAmount < 0 || co.RemainingAmount < 0 {
					msg := fmt.Sprintf("Remaining field is less than zero. order: %v candidate: %v", order, co)
					panic(msg)
				}

				if co.RemainingAmount == 0 {
					// Order has been filled. Remove it from queue.
					_, _ = i.Orders.Get(1)
				}

				if order.RemainingAmount == 0 {
					// Order has been filled.
					break
				}
			}
		}
	}

	if order.RemainingAmount > 0 {
		k.instruments.InsertOrder(order)

	}

	return sdk.Result{Events: ctx.EventManager().Events()}
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
