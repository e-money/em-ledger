// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"fmt"
	"math/rand"
	"runtime/debug"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/e-money/em-ledger/x/market/types"
	"github.com/stretchr/testify/require"
)

const (
	maxPriceVariation = 20 // Create variations on price in the interval +-10% of the base price
)

// go test -v -timeout 24h -run TestFuzzingInfinite ./x/market/keeper/
//func TestFuzzingInfinite(t *testing.T) {
//	for {
//		TestFuzzing1(t)
//	}
//}

func TestFuzzing1(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Println("Using seed", seed)
	r := rand.New(rand.NewSource(seed))

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("seed", seed, "caused a panic:", r)
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			t.Fail()
		}
	}()

	ctx, k, ak, _, _ := createTestComponents(t)

	var (
		acc1 = createAccount(ctx, ak, "acc1", "1000000000eur")
		acc2 = createAccount(ctx, ak, "acc2", "1000000000usd")
		acc3 = createAccount(ctx, ak, "acc3", "1000000000chf")
	)

	totalSupply := snapshotAccounts(ctx, ak)

	basepriceEURUSD := sdk.ZeroDec()
	for basepriceEURUSD.IsZero() {
		basepriceEURUSD = sdk.NewDecWithPrec(r.Int63n(1000), 2)
	}

	basePriceUSDCHF := sdk.ZeroDec()
	for basePriceUSDCHF.IsZero() {
		basePriceUSDCHF = sdk.NewDecWithPrec(r.Int63n(1000), 2)
	}

	ONE := sdk.OneDec()
	testdata := []struct {
		src, dst string
		price    sdk.Dec
		seller   exported.Account
	}{
		{"eur", "usd", basepriceEURUSD, acc1},
		{"usd", "eur", ONE.Quo(basepriceEURUSD), acc2},
		{"usd", "chf", basePriceUSDCHF, acc2},
		{"chf", "usd", ONE.Quo(basePriceUSDCHF), acc3},
		{"eur", "chf", basepriceEURUSD.Mul(basePriceUSDCHF), acc1},
		{"chf", "eur", ONE.Quo(basepriceEURUSD.Mul(basePriceUSDCHF)), acc3},
	}

	allOrders := make([]types.Order, 0)

	for _, instr := range testdata {
		allOrders = append(allOrders, generateOrders(instr.src, instr.dst, instr.price, instr.seller, r)...)
	}

	r.Shuffle(len(allOrders), func(i, j int) {
		allOrders[i], allOrders[j] = allOrders[j], allOrders[i]
	})

	for _, order := range allOrders {
		_, err := k.NewOrderSingle(ctx, order)
		if order.IsFilled() {
			fmt.Println("Order is filled on creation. Ignoring.", order)
			continue
		}
		require.NoError(t, err)
	}

	//dumpEvents(ctx.EventManager().ABCIEvents())
	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, ak)).IsZero())
}

func generateOrders(srcDenom, dstDenom string, basePrice sdk.Dec, seller exported.Account, r *rand.Rand) (res []types.Order) {
	priceGen := priceGenerator(basePrice, r)

	for i := 0; i < 500; i++ {
		var (
			source      = sdk.NewCoin(srcDenom, sdk.NewInt(r.Int63n(1000000)+1)) // Sell up to 1 million.
			destination = sdk.NewCoin(dstDenom, source.Amount.ToDec().Mul(priceGen()).RoundInt())
		)

		if destination.IsZero() {
			// A low number of source tokens combined with a low price may create invalid orders. Skip these. (Seed 1580745227 if you want to see for yourself)
			continue
		}

		o := order(seller, source.String(), destination.String())

		switch r.Intn(3) {
		case 0:
			o.TimeInForce = types.TimeInForce_FillOrKill
		case 1:
			o.TimeInForce = types.TimeInForce_GoodTilCancel
		case 2:
			o.TimeInForce = types.TimeInForce_ImmediateOrCancel
		}

		res = append(res, o)
	}

	return res
}

func priceGenerator(baseprice sdk.Dec, r *rand.Rand) func() sdk.Dec {
	return func() sdk.Dec {
		modifier := sdk.NewDecWithPrec(r.Int63n(maxPriceVariation), 2)
		modifier = modifier.Sub(sdk.NewDecWithPrec(maxPriceVariation>>1, 2))
		modifier = sdk.OneDec().Add(modifier)
		return baseprice.Mul(modifier)
	}
}
