package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/offer/types"
	"math"
)

type Keeper struct {
	instruments types.Instruments
}

func NewKeeper() Keeper {
	return Keeper{}
}

func (k *Keeper) ProcessOrder(ctx sdk.Context, order *types.Order) sdk.Result {
	fmt.Println("\nProcessing order", order.ID)
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
			}
		}
	}

	if order.RemainingAmount > 0 {
		k.instruments.InsertOrder(order)

	}

	return sdk.Result{Events: ctx.EventManager().Events()}
}
