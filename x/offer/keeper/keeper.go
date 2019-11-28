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
	sourceAccount := k.ak.GetAccount(ctx, order.SourceAccount)
	if sourceAccount == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", order.SourceAccount.String())).Result()
	}

	// Verify account balance
	if _, anyNegative := sourceAccount.GetCoins().SafeSub(sdk.NewCoins(order.Source)); anyNegative {
		return types.ErrAccountBalanceInsufficient(order.SourceAccount, order.Source, sourceAccount.GetCoins().AmountOf(order.Source.Denom)).Result()
	}

	order.ID = k.GetNextOrderNumber(ctx)

	for _, i := range k.instruments {
		if i.Source == order.Destination.Denom && i.Destination == order.Source.Denom {
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

				//fmt.Println("Candidate order:\n", co.String())
				//fmt.Println("Incoming order:\n", order.String())
				//fmt.Println()

				// Price is divided evenly between bid and offer. Price improvement is shared equally.
				matchingPrice := order.Destination.Amount.Add(co.Source.Amount).ToDec().Quo(order.Source.Amount.Add(co.Destination.Amount).ToDec())

				//matchingPrice := (float64(order.Destination.Amount.Int64()) + float64(co.Source.Amount.Int64())) / (float64(order.Source.Amount.Int64()) + float64(co.Destination.Amount.Int64()))

				// Price improvement is 100% given to the buyer.
				//matchingPrice := co.invertedPrice

				sourceMatched := co.Remaining.ToDec().QuoRoundUp(matchingPrice).TruncateInt()
				//sourceMatched := int64(math.Ceil(float64(co.Remaining.Int64()) / matchingPrice))
				if order.Remaining.LT(sourceMatched) {
					sourceMatched = order.Remaining
				}

				destinationMatched := sourceMatched.ToDec().Mul(matchingPrice).Ceil().TruncateInt()
				//destinationMatched := int64(math.Floor(float64(sourceMatched) * matchingPrice))

				err := k.transferTradedVolumes(ctx, destinationMatched, sourceMatched, co, order)
				if err != nil {
					return err.Result()
				}

				// Adjust orders
				co.Remaining = co.Remaining.Sub(destinationMatched)
				order.Remaining = order.Remaining.Sub(sourceMatched)

				// Invariant check
				if order.Remaining.LT(sdk.ZeroInt()) || co.Remaining.LT(sdk.ZeroInt()) {
					msg := fmt.Sprintf("Remaining field is less than zero. order: %v candidate: %v", order, co)
					panic(msg)
				}

				if co.Remaining.IsZero() {
					// Order has been filled. Remove it from queue.
					_, _ = i.Orders.Get(1)
				}

				if order.Remaining.IsZero() {
					// Order has been filled.
					break
				}
			}
		}
	}

	if !order.Remaining.IsZero() {
		k.instruments.InsertOrder(order)
	}

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func (k Keeper) transferTradedVolumes(ctx sdk.Context, destinationMatched, sourceMatched sdk.Int, co, order *types.Order) sdk.Error {
	var (
		candidateAccount = k.ak.GetAccount(ctx, co.SourceAccount)
		orderAccount     = k.ak.GetAccount(ctx, order.SourceAccount)
	)

	// Verify that the passive order still holds the balance
	coinMatchedDst := sdk.NewCoin(co.Source.Denom, destinationMatched)
	if _, anyNegative := candidateAccount.GetCoins().SafeSub(sdk.NewCoins(coinMatchedDst)); anyNegative {
		return types.ErrAccountBalanceInsufficient(candidateAccount.GetAddress(), coinMatchedDst, candidateAccount.GetCoins().AmountOf(coinMatchedDst.Denom))
	}

	// Verify that the aggressive order still holds the balance
	coinMatchedSrc := sdk.NewCoin(order.Source.Denom, sourceMatched)
	if _, anyNegative := orderAccount.GetCoins().SafeSub(sdk.NewCoins(coinMatchedSrc)); anyNegative {
		return types.ErrAccountBalanceInsufficient(orderAccount.GetAddress(), coinMatchedSrc, orderAccount.GetCoins().AmountOf(coinMatchedSrc.Denom))
	}

	// Balances appear sufficient. Do the transfers
	k.bk.SendCoins(ctx, order.SourceAccount, co.SourceAccount, sdk.NewCoins(coinMatchedSrc))
	k.bk.SendCoins(ctx, co.SourceAccount, order.SourceAccount, sdk.NewCoins(coinMatchedDst))
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
