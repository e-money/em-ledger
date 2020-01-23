package keeper

import (
	"fmt"
	"math/rand"
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

func TestFuzzing1(t *testing.T) {
	seed := time.Now().Unix()
	fmt.Println("Using seed", seed)
	r := rand.New(rand.NewSource(seed))

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
		res := k.NewOrderSingle(ctx, order)
		require.True(t, res.IsOK())
	}

	//dumpEvents(ctx.EventManager().Events())
	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, ak)).IsZero())
}

func generateOrders(srcDenom, dstDenom string, basePrice sdk.Dec, seller exported.Account, r *rand.Rand) (res []types.Order) {
	priceGen := priceGenerator(basePrice, r)

	for i := 0; i < 500; i++ {
		var (
			source      = sdk.NewCoin(srcDenom, sdk.NewInt(r.Int63n(1000000)+1)) // Sell up to 1 million.
			destination = sdk.NewCoin(dstDenom, source.Amount.ToDec().Mul(priceGen()).RoundInt())
		)

		res = append(res, order(seller, source.String(), destination.String()))
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
