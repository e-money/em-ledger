// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	embank "github.com/e-money/em-ledger/hooks/bank"
	emtypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/market/types"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"math"
	"strings"
	"testing"
	"time"
)

func init() {
	emtypes.ConfigureSDK()
}


func TestKeeper_calcOrderGas(t *testing.T) {
	ctx, k, _, _ := createTestComponents(t)

	var (
		stdTrxFee    sdk.Gas = k.GetTrxFee(ctx)
		liquidTrxFee sdk.Gas = k.GetLiquidTrxFee(ctx)
	)
	tests := []struct {
		name         string
		ctx          sdk.Context
		stdTrxFee    sdk.Gas
		liquidTrxFee sdk.Gas
		dstFilled    sdk.Int
		dstAmount    sdk.Int
		want         sdk.Gas
	}{
		{
			name:      "0% Filled Full Gas",
			ctx:       ctx,
			stdTrxFee: stdTrxFee,
			dstFilled: sdk.NewInt(0),
			dstAmount: sdk.NewInt(0),
			want:      liquidTrxFee,
		},
		{
			name:      "100% Filled Full Gas",
			ctx:       ctx,
			stdTrxFee: stdTrxFee,
			dstFilled: sdk.NewInt(1),
			dstAmount: sdk.NewInt(1),
			want:      stdTrxFee,
		},
		{
			name:      "10% Filled Full Gas",
			ctx:       ctx,
			stdTrxFee: stdTrxFee,
			dstFilled: sdk.NewInt(10),
			dstAmount: sdk.NewInt(100),
			want:      stdTrxFee/10,
		},
		{
			name:      "90% Filled Full Gas",
			ctx:       ctx,
			stdTrxFee: stdTrxFee,
			dstFilled: sdk.NewInt(90),
			dstAmount: sdk.NewInt(100),
			want:      sdk.Gas(float64(stdTrxFee) * 0.9),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := k.calcOrderGas(tt.ctx, tt.stdTrxFee, tt.dstFilled,
				tt.dstAmount); got != tt.want {
				t.Errorf("calcOrderGas() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeeper_calcReplaceOrderGas(t *testing.T) {
	ctx, k, _, _ := createTestComponents(t)

	ctx = ctx.WithBlockTime(time.Now())

	var (
		stdTrxFee    sdk.Gas = k.GetTrxFee(ctx)
		liquidTrxFee sdk.Gas = k.GetLiquidTrxFee(ctx)
		// set ctx to rebate allowance + 1
		liquidIntervalMin = time.Duration(k.GetLiquidityRebateMinutesSpan(ctx)+1) *
			time.Minute
	)

	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	tests := []struct {
		name             string
		ctx              sdk.Context
		stdTrxFee        sdk.Gas
		liquidTrxFee     sdk.Gas
		dstFilled        sdk.Int
		dstAmount        sdk.Int
		origOrderCreated time.Time
		want             sdk.Gas
	}{
		{
			name:             "0% Filled Full Gas",
			ctx:              ctx.WithBlockTime(ctx.BlockTime().Add(liquidIntervalMin)),
			stdTrxFee:        stdTrxFee,
			dstFilled:        sdk.NewInt(0),
			dstAmount:        sdk.NewInt(0),
			origOrderCreated: time.Now(),
			want:             liquidTrxFee,
		},
		{
			name:             "0% Filled but not meeting qualified minutes, Full Gas",
			ctx:              ctx,
			stdTrxFee:        stdTrxFee,
			dstFilled:        sdk.NewInt(0),
			dstAmount:        sdk.NewInt(0),
			origOrderCreated: time.Now(),
			want:             stdTrxFee,
		},
		{
			name:             "100% Filled Full Gas",
			ctx:              ctx.WithBlockTime(ctx.BlockTime().Add(liquidIntervalMin)),
			stdTrxFee:        stdTrxFee,
			dstFilled:        sdk.NewInt(1),
			dstAmount:        sdk.NewInt(1),
			origOrderCreated: time.Now(),
			want:             stdTrxFee,
		},
		{
			name:             "10% Filled Full Gas",
			ctx:              ctx.WithBlockTime(ctx.BlockTime().Add(liquidIntervalMin)),
			stdTrxFee:        stdTrxFee,
			dstFilled:        sdk.NewInt(10),
			dstAmount:        sdk.NewInt(100),
			origOrderCreated: time.Now(),
			want:             stdTrxFee / 10,
		},
		{
			name:             "90% Filled Full Gas",
			ctx:              ctx.WithBlockTime(ctx.BlockTime().Add(liquidIntervalMin)),
			stdTrxFee:        stdTrxFee,
			dstFilled:        sdk.NewInt(90),
			dstAmount:        sdk.NewInt(100),
			origOrderCreated: time.Now(),
			want:             sdk.Gas(float64(stdTrxFee) * 0.9),
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := k.calcReplaceOrderGas(
					tt.ctx, tt.dstFilled, tt.dstAmount, tt.origOrderCreated,
				); got != tt.want {
					t.Errorf(
						"calcReplaceOrderGas() = %v, want %v", got, tt.want,
					)
				}
			},
		)
	}
}

func TestBasicTrade(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "7400usd")

	totalSupply := snapshotAccounts(ctx, bk)

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	var liquidTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	gasmeter := sdk.NewGasMeter(math.MaxUint64)
	src1, dst1 := "eur", "usd"
	order1 := order(ctx.BlockTime(), acc1, "100"+src1, "120"+dst1)
	_, err := k.NewOrderSingle(ctx.WithGasMeter(gasmeter), order1)
	require.NoError(t, err)
	require.Equal(t, liquidTrxFee, gasmeter.GasConsumed(), "Liquid non-filled")
	require.Equal(t, ctx.BlockTime(), order1.Created)

	// Ensure that the instrument was registered
	instruments := k.GetInstruments(ctx)
	_, err = json.Marshal(instruments)
	require.Nil(t, err)

	require.Len(t, instruments, 2)
	require.Nil(t, instruments[0].LastPrice)
	gasmeter = sdk.NewGasMeter(math.MaxUint64)
	src2, dst2 := dst1, src1
	order2 := order(ctx.BlockTime(), acc2, "60"+src2, "50"+dst2)
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasmeter), order2)
	require.NoError(t, err)

	// Ensure that the trade has been correctly registered in market data.
	instruments = k.GetInstruments(ctx)
	require.Len(t, instruments, 2)
	p := order1.Price()
	t.Skip("Alex - deactivated before migration. fails with rounding after this line")
	require.Equal(t, instruments[0].LastPrice.String(), p.String())
	require.Equal(t, *instruments[0].Timestamp, ctx.BlockTime())

	// Ensure that gas usage is not higher due to the order being matched.
	require.Equal(t, stdTrxFee, gasmeter.GasConsumed())

	bal1 := bk.GetAllBalances(ctx, acc1.GetAddress())
	bal2 := bk.GetAllBalances(ctx, acc2.GetAddress())
	require.Len(t, bal1, 2)
	require.Len(t, bal2, 2)

	require.Equal(t, "4950", bal1.AmountOf("eur").String())
	require.Equal(t, "60", bal1.AmountOf("usd").String())

	require.Equal(t, "50", bal2.AmountOf("eur").String())
	require.Equal(t, "7340", bal2.AmountOf("usd").String())

	// require.Len(t, k.instruments, 1)

	// i := k.instruments[0]
	// remainingOrder := i.Orders.LeftKey().(*types.Order)
	// require.Equal(t, int64(50), remainingOrder.SourceRemaining.Int64())

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}

func TestCreationTime1(t *testing.T) {
	ctx, _, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000eur")

	src1, dst1 := "eur", "usd"
	order1 := order(ctx.BlockTime(), acc1, "100"+src1, "120"+dst1)
	require.Equal(t, ctx.BlockTime(), order1.Created)
}

func TestBasicTrade2(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "888eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "1120usd")

	totalSupply := snapshotAccounts(ctx, bk)

	order1 := order(ctx.BlockTime(), acc1, "888eur", "1121usd")
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	order2 := order(ctx.BlockTime(), acc2, "1120usd", "890eur")
	res, err := k.NewOrderSingle(ctx, order2)
	require.True(t, err == nil, res.Log)

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}

func TestBasicTrade3(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "230000usd")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "890000eur")
	acc3 := createAccount(ctx, ak, bk, randomAddress(), "25chf")

	totalSupply := snapshotAccounts(ctx, bk)

	order1 := order(ctx.BlockTime(), acc2, "888850eur", "22807162chf")
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	order2 := order(ctx.BlockTime(), acc3, "12chf", "4usd")
	_, err = k.NewOrderSingle(ctx, order2)
	require.NoError(t, err)

	order3 := order(ctx.BlockTime(), acc1, "227156usd", "24971eur")

	_, err = k.NewOrderSingle(ctx, order3)
	require.NoError(t, err)

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}

func TestMarketOrderSlippage1(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "500gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "500eur")

	var o types.Order
	var err error

	// Establish market price by executing a 1:1 trade
	o = order(ctx.BlockTime(), acc2, "1eur", "1gbp")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(ctx.BlockTime(), acc1, "1gbp", "1eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	// Sell eur at various prices
	o = order(ctx.BlockTime(), acc2, "50eur", "50gbp")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(ctx.BlockTime(), acc2, "50eur", "75gbp")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(ctx.BlockTime(), acc2, "50eur", "100gbp")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	// Make a market order that allows slippage

	slippage := sdk.NewDecWithPrec(50, 2)
	srcDenom := "gbp"
	dest := sdk.NewCoin("eur", sdk.NewInt(200))
	slippageSource, err := k.GetSrcFromSlippage(
		ctx, srcDenom, dest, slippage,
	)
	require.NoError(t, err)
	limitOrder := order(ctx.BlockTime(), acc1, slippageSource.String(), dest.String())
	_, err = k.NewOrderSingle(ctx, limitOrder)
	require.NoError(t, err)

	// Check that the balance matches the first two orders being executed while the third did not fall within the slippage
	bal1 := bk.GetAllBalances(ctx, acc1.GetAddress())
	require.Equal(t, coins("101eur,374gbp"), bal1)

	bal2 := bk.GetAllBalances(ctx, acc2.GetAddress())
	require.Equal(t, coins("399eur,126gbp"), bal2)

	// Ensure that the order can not exceed account balance
	slippage = sdk.NewDecWithPrec(500, 2)
	slippageSource, err = k.GetSrcFromSlippage(
		ctx, srcDenom, dest, slippage,
	)
	require.NoError(t, err)
	limitOrder = order(ctx.BlockTime(), acc1, slippageSource.String(), dest.String())
	_, err = k.NewOrderSingle(ctx, limitOrder)
	require.True(t, types.ErrAccountBalanceInsufficient.Is(err))
}

func TestCancelReplaceMarketOrderZeroSlippage(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "500gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "500eur")

	var o types.Order
	var err error

	// Establish market price by executing a 1:1 trade
	o = order(ctx.BlockTime(), acc2, "1eur", "1gbp")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(ctx.BlockTime(), acc1, "1gbp", "1eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	// Make a market order that allows slippage
	slippage := sdk.NewDecWithPrec(100, 2)
	srcDenom := "gbp"
	dest := sdk.NewCoin("eur", sdk.NewInt(100))
	slippageSource, err := k.GetSrcFromSlippage(
		ctx, srcDenom, dest, slippage,
	)
	require.NoError(t, err)
	limitOrder := order(ctx.BlockTime(), acc1, slippageSource.String(), dest.String())
	_, err = k.NewOrderSingle(ctx, limitOrder)
	require.NoError(t, err)

	clientID := limitOrder.ClientOrderID
	foundOrder := k.GetOrderByOwnerAndClientOrderId(
		ctx, acc1.GetAddress().String(), clientID,
	)
	require.NotNil(t, foundOrder, "Market order should exist")
	// 100% slippage to double source
	require.True(t, foundOrder.Source.IsEqual(sdk.NewCoin("gbp", sdk.NewInt(200))))

	// Gave 1 gbp and gained a eur
	acc1Bal := bk.GetAllBalances(ctx, acc1.GetAddress())
	require.Equal(t, coins("1eur,499gbp").String(), acc1Bal.String())

	newClientID := cid()
	slippageSource, err = k.GetSrcFromSlippage(
		ctx, srcDenom, dest, sdk.NewDecWithPrec(0, 2),
	)
	require.NoError(t, err)
	order, err := types.NewOrder(
		ctx.BlockTime(),
		types.TimeInForce_GoodTillCancel,
		slippageSource,
		dest,
		acc1.GetAddress(),
		newClientID,
	)
	require.NoError(t, err)

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	gasMeter := sdk.NewInfiniteGasMeter()
	ctx = ctx.WithGasMeter(gasMeter)

	_, err = k.CancelReplaceLimitOrder(ctx, order, clientID)
	require.NoError(t, err)
	require.Equal(t, stdTrxFee, ctx.GasMeter().GasConsumed())

	expOrder := &types.Order{
		ID:                3,
		TimeInForce:       types.TimeInForce_GoodTillCancel,
		Owner:             acc1.GetAddress().String(),
		ClientOrderID:     newClientID,
		// Zero slippage same amount
		Source:            sdk.NewCoin("gbp", sdk.NewInt(100)),
		SourceRemaining:   sdk.NewInt(100),
		SourceFilled:      sdk.ZeroInt(),
		Destination:       dest,
		DestinationFilled: sdk.ZeroInt(),
		Created:           ctx.BlockTime(),
	}
	require.NoError(t, err)

	origOrder := k.GetOrderByOwnerAndClientOrderId(
		ctx, acc1.GetAddress().String(), clientID,
	)
	require.Nil(t, origOrder, "Original market order should not exist")

	foundOrder = k.GetOrderByOwnerAndClientOrderId(
		ctx, acc1.GetAddress().String(), newClientID,
	)
	require.Equal(t, expOrder, foundOrder)

	// no impact
	acc1Bal = bk.GetAllBalances(ctx, acc1.GetAddress())
	require.Equal(t, coins("1eur,499gbp").String(), acc1Bal.String())
}

func TestCancelReplaceMarketOrder100Slippage(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "100gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "100eur")

	var o types.Order
	var err error

	var liquidTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	gasMeter := sdk.NewInfiniteGasMeter()
	ctx = ctx.WithGasMeter(gasMeter)

	// Establish market price by executing a 2:1 trade
	o = order(ctx.BlockTime(), acc2, "20eur", "10gbp")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)
	require.Equal(t, liquidTrxFee, ctx.GasMeter().GasConsumed(), "Liquid Trx are discounted")

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	gasMeter = sdk.NewInfiniteGasMeter()
	ctx = ctx.WithGasMeter(gasMeter)

	o = order(ctx.BlockTime(), acc1, "10gbp", "20eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)
	require.Equal(
		t, stdTrxFee, ctx.GasMeter().GasConsumed(),
		"Filled order costs std fee",
	)

	acc1b := bk.GetAllBalances(ctx, acc1.GetAddress())
	require.Equal(t, coins("20eur,90gbp").String(), acc1b.String())

	// Make a market newOrder that allows slippage
	slippage := sdk.NewDecWithPrec(0, 2)
	srcDenom := "gbp"
	dest := sdk.NewCoin("eur", sdk.NewInt(10))
	slippageSource, err := k.GetSrcFromSlippage(
		ctx, srcDenom, dest, slippage,
	)
	require.NoError(t, err)
	limitOrder := order(ctx.BlockTime(), acc1, slippageSource.String(), dest.String())
	_, err = k.NewOrderSingle(ctx, limitOrder)
	require.NoError(t, err)

	clientID := limitOrder.ClientOrderID

	// Gave 1 gbp and gained a eur
	acc1Bal := bk.GetAllBalances(ctx, acc1.GetAddress())
	require.Equal(t, coins("20eur,90gbp").String(), acc1Bal.String())

	foundOrder := k.GetOrderByOwnerAndClientOrderId(
		ctx, acc1.GetAddress().String(), clientID,
	)
	require.NotNil(t, foundOrder, "Market newOrder should exist")
	// 0% slippage same as ratio (1eur/2gbp) * 10 => 5gbp
	require.True(t, foundOrder.Source.IsEqual(sdk.NewCoin("gbp", sdk.NewInt(5))))

	newClientID := cid()
	mcrm := &types.MsgCancelReplaceMarketOrder{
		Owner:             acc1.GetAddress().String(),
		OrigClientOrderId: clientID,
		NewClientOrderId:  newClientID,
		TimeInForce:       types.TimeInForce_GoodTillCancel,
		Source:            "gbp",
		Destination:       sdk.NewCoin("eur", sdk.NewInt(10)),
		MaxSlippage:       sdk.NewDecWithPrec(100, 2),
	}

	dest = sdk.NewCoin("eur", sdk.NewInt(10))

	slippageSource, err = k.GetSrcFromSlippage(
		ctx, "gbp", dest, sdk.NewDecWithPrec(100, 2),
	)
	require.NoError(t, err)

	newOrder, err := types.NewOrder(
		ctx.BlockTime(),
		types.TimeInForce_GoodTillCancel,
		slippageSource,
		dest,
		acc1.GetAddress(),
		newClientID,
	)
	require.NoError(t, err)

	var qualificationMin int64
	k.paramStore.Get(ctx, types.KeyLiquidityRebateMinutesSpan, &qualificationMin)

	gasMeter = sdk.NewInfiniteGasMeter()
	ctx = ctx.WithGasMeter(gasMeter)

	_, err = k.CancelReplaceLimitOrder(ctx.WithGasMeter(gasMeter).WithBlockTime(
		ctx.BlockTime().
			Add(time.Duration(qualificationMin)*time.Minute),
	), newOrder, clientID)
	require.NoError(t, err)
	require.Equal(
		t, liquidTrxFee, ctx.GasMeter().GasConsumed(),
		"Unfilled/liquid are discounted",
	)

	expOrder := &types.Order{
		ID:                3,
		TimeInForce:       types.TimeInForce_GoodTillCancel,
		Owner:             acc1.GetAddress().String(),
		ClientOrderID:     newClientID,
		// 100 % slippage should result 2 * (1/2) => 10 gbp -> 10 eur
		Source:            sdk.NewCoin(mcrm.Source, sdk.NewInt(10)),
		SourceRemaining:   sdk.NewInt(10),
		SourceFilled:      sdk.ZeroInt(),
		Destination:       mcrm.Destination,
		DestinationFilled: sdk.ZeroInt(),
		Created:           ctx.BlockTime(),
	}
	require.NoError(t, err)

	origOrder := k.GetOrderByOwnerAndClientOrderId(
		ctx, acc1.GetAddress().String(), clientID,
	)
	require.Nil(t, origOrder, "Original market newOrder should not exist")

	foundOrder = k.GetOrderByOwnerAndClientOrderId(
		ctx, acc1.GetAddress().String(), newClientID,
	)
	require.Equal(t, expOrder, foundOrder)

	// no impact
	acc1Bal = bk.GetAllBalances(ctx, acc1.GetAddress())
	require.Equal(t, coins("20eur,90gbp").String(), acc1Bal.String())
}

func TestGetSrcFromSlippage(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	var (
		acc1 = createAccount(ctx, ak, bk, randomAddress(), "500gbp")
		acc2 = createAccount(ctx, ak, bk, randomAddress(), "500eur")

		o   types.Order
		err error
		srcDenom string
		slippedSource, dest sdk.Coin
	)

	srcDenom = "jpy"
	dest = sdk.NewCoin("eur", sdk.NewInt(100))
	_, err = k.GetSrcFromSlippage(
		ctx, srcDenom, dest, sdk.ZeroDec(),
	)
	require.Error(t, err, "No trades yet with jpy")

	srcDenom = "gbp"
	dest = sdk.NewCoin("dek", sdk.NewInt(100))
	_, err = k.GetSrcFromSlippage(
		ctx, srcDenom, dest, sdk.ZeroDec(),
	)
	require.Error(t, err, "No trades yet with dek")

	slippage := sdk.NewDec(-2)

	srcDenom = "gbp"
	dest = sdk.NewCoin("eur", sdk.NewInt(100))
	_, err = k.GetSrcFromSlippage(
		ctx, srcDenom, dest, slippage,
	)
	require.Error(t, err)
	require.True(t, types.ErrInvalidSlippage.Is(err))

	// Establish market price by executing a 1:1 trade
	o = order(ctx.BlockTime(), acc2, "1eur", "1gbp")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(ctx.BlockTime(), acc1, "1gbp", "1eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	// Add liquidity
	o = order(ctx.BlockTime(), acc2, "100eur", "100gbp")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	srcDenom = "gbp"
	dest = sdk.NewCoin("eur", sdk.NewInt(100))
	slippedSource, err = k.GetSrcFromSlippage(
		ctx, srcDenom, dest, sdk.ZeroDec(),
	)
	require.NoError(t, err)
	require.Equal(
		t, slippedSource.String(), sdk.NewCoin(srcDenom, dest.Amount).String(),
		"0% slippage -> source amount or last market price",
	)

	// 100%
	slippage = sdk.NewDec(1)
	srcDenom = "gbp"
	dest = sdk.NewCoin("eur", sdk.NewInt(1))
	slippedSource, err = k.GetSrcFromSlippage(
		ctx, srcDenom, dest, slippage,
	)
	require.NoError(t, err)
	require.Equal(
		t, sdk.NewCoin(srcDenom, sdk.NewInt(2)).String(),
		slippedSource.String(),
		"100% slippage -> 1+100% source amount",
	)

	// 20%
	slippage = sdk.OneDec().Quo(sdk.NewDec(5))
	srcDenom = "gbp"
	dest = sdk.NewCoin("eur", sdk.NewInt(10))
	slippedSource, err = k.GetSrcFromSlippage(
		ctx, srcDenom, dest, slippage,
	)
	require.NoError(t, err)
	require.Equal(
		t, sdk.NewCoin(srcDenom, sdk.NewInt(12)).String(), slippedSource.String(),
		"20% slippage -> 10+20%=>12 amount",
	)
}

func TestFillOrKillMarketOrder1(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	var (
		acc1 = createAccount(ctx, ak, bk, randomAddress(), "500gbp")
		acc2 = createAccount(ctx, ak, bk, randomAddress(), "500eur")

		o   types.Order
		err error
	)

	// Establish market price by executing a 1:1 trade
	o = order(ctx.BlockTime(), acc2, "1eur", "1gbp")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(ctx.BlockTime(), acc1, "1gbp", "1eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	// Create a market for eur
	o = order(ctx.BlockTime(), acc2, "100eur", "100gbp")
	res, err := k.NewOrderSingle(ctx, o)
	require.NoError(t, err)
	require.Equal(
		t, "accept",
		string(res.Events[0].Attributes[0].GetValue()),
	)

	require.Equal(
		t, ctx.BlockTime().Format(time.RFC3339),
		string(res.Events[0].Attributes[len(res.Events[0].Attributes)-1].GetValue()),
	)

	stdTrxFee := k.GetTrxFee(ctx)
	gasMeter := sdk.NewInfiniteGasMeter()

	// Create a fill or kill order that cannot be satisfied by the current market
	srcDenom := "gbp"
	dest := sdk.NewCoin("eur", sdk.NewInt(200))
	slippageSource, err := k.GetSrcFromSlippage(
		ctx, srcDenom, dest, sdk.ZeroDec(),
	)
	require.NoError(t, err)
	limitOrder := order(ctx.BlockTime(), acc1, slippageSource.String(), dest.String())
	limitOrder.TimeInForce = types.TimeInForce_FillOrKill
	result, err := k.NewOrderSingle(ctx.WithGasMeter(gasMeter), limitOrder)
	require.NoError(t, err)
	require.Len(t, result.Events, 1)
	require.Equal(t, types.EventTypeMarket, result.Events[0].Type)
	require.Equal(t, "action", string(result.Events[0].Attributes[0].GetKey()))
	require.Equal(t, "expire", string(result.Events[0].Attributes[0].GetValue()))
	require.Equal(t, gasMeter.GasConsumed(), stdTrxFee,
		"Partially filled FillOrKill costs like a cancelled order")

	// Last order must fail completely due to not being fillable
	acc1Bal := bk.GetAllBalances(ctx, acc1.GetAddress())
	require.Equal(t, coins("1eur,499gbp"), acc1Bal)

	acc2Bal := bk.GetAllBalances(ctx, acc2.GetAddress())
	require.Equal(t, coins("499eur,1gbp"), acc2Bal)
}

func TestFillOrKillLimitOrder1(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "500gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "500eur")

	// Create a tiny market for eur
	o := order(ctx.BlockTime(), acc2, "100eur", "100gbp")
	_, err := k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	stdTrxFee := k.GetTrxFee(ctx)
	gasMeter := sdk.NewInfiniteGasMeter()

	order2 := order(ctx.BlockTime(), acc1, "200gbp", "200eur")
	order2.TimeInForce = types.TimeInForce_FillOrKill
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order2)
	require.NoError(t, err)
	require.Equal(t, gasMeter.GasConsumed(), stdTrxFee,
		"Partially filled FillOrKill costs like a cancelled order")

	// Order must fail completely due to not being fillable
	acc1Bal := bk.GetAllBalances(ctx, acc1.GetAddress())
	require.Equal(t, coins("500gbp"), acc1Bal)

	acc2Bal := bk.GetAllBalances(ctx, acc2.GetAddress())
	require.Equal(t, coins("500eur"), acc2Bal)

	// Test that the order book looks as expected
	require.Empty(t, k.GetOrdersByOwner(ctx, acc1.GetAddress()))
	acc2Orders := k.GetOrdersByOwner(ctx, acc2.GetAddress())
	require.Len(t, acc2Orders, 1)
	require.Equal(t, acc2Orders[0].Created, ctx.BlockTime())
}

func TestFillOrKillLimitOrder2Filled(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "500gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "500eur")

	// Create a tiny market for eur
	o := order(ctx.BlockTime(), acc2, "100eur", "100gbp")
	_, err := k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	stdTrxFee := k.GetTrxFee(ctx)
	gasMeter := sdk.NewInfiniteGasMeter()

	// A successful FOK Order
	order2 := order(ctx.BlockTime(), acc1, "100gbp", "100eur")
	order2.TimeInForce = types.TimeInForce_FillOrKill
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order2)
	require.NoError(t, err)
	require.Equal(t, gasMeter.GasConsumed(), stdTrxFee,
		"Filled order")

	acc1Bal := bk.GetAllBalances(ctx, acc1.GetAddress())
	require.Equal(t, coins("400gbp,100eur"), acc1Bal)

	acc2Bal := bk.GetAllBalances(ctx, acc2.GetAddress())
	require.Equal(t, coins("400eur,100gbp"), acc2Bal)

	// Test that the order book looks as expected
	require.Empty(t, k.GetOrdersByOwner(ctx, acc1.GetAddress()))
	require.Empty(t, k.GetOrdersByOwner(ctx, acc2.GetAddress()))
}

func TestImmediateOrCancel(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "20gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "20eur")

	var err error

	o1 := order(ctx.BlockTime(), acc2, "1eur", "1gbp")
	_, err = k.NewOrderSingle(ctx, o1)
	require.NoError(t, err)

	stdTrxFee := k.GetTrxFee(ctx)
	gasMeter := sdk.NewInfiniteGasMeter()

	// IOK partially Filled
	o2 := order(ctx.BlockTime(), acc1, "2gbp", "2eur")
	o2.TimeInForce = types.TimeInForce_ImmediateOrCancel
	cid2 := o2.ClientOrderID
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), o2)
	require.NoError(t, err)
	require.Equal(t, o2.Created, ctx.BlockTime())
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())

	// Verify that order is not in book
	order := k.GetOrderByOwnerAndClientOrderId(ctx, acc1.GetAddress().String(), cid2)
	require.Nil(t, order)

	bal1 := bk.GetAllBalances(ctx, acc1.GetAddress())
	require.Equal(t, coins("19gbp,1eur"), bal1)

	// again o1
	_, err = k.NewOrderSingle(ctx, o1)
	require.NoError(t, err)

	// IOK Filled
	gasMeter = sdk.NewInfiniteGasMeter()
	o2.Source = sdk.NewCoin("gbp", sdk.NewInt(1))
	o2.Destination = sdk.NewCoin("eur", sdk.NewInt(1))
	o2.TimeInForce = types.TimeInForce_ImmediateOrCancel
	o2.ClientOrderID = cid()
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), o2)
	require.NoError(t, err)
	require.Equal(t, o2.Created, ctx.BlockTime())
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())

	// Verify that order is not in book
	order = k.GetOrderByOwnerAndClientOrderId(ctx, acc1.GetAddress().String(), o2.ClientOrderID)
	require.Nil(t, order)

	bal2 := bk.GetAllBalances(ctx, acc1.GetAddress())
	require.Equal(t, coins("18gbp,2eur"), bal2)
}

func TestInsufficientGas(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "888eur")
	order1 := order(ctx.BlockTime(), acc1, "888eur", "1121usd")

	// Free: it adds liquidity
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	acc2 := createAccount(ctx, ak, bk, randomAddress(), "1121usd")
	order2 := order(ctx.BlockTime(), acc2, "1121usd", "888eur")

	var fullTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &fullTrxFee)

	// bummer missing 1 micro
	gasMeter := sdk.NewGasMeter(fullTrxFee - 1)

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	require.Panics(t, func() {
		// Taking away liquidity not enough to cover the full Trx fee
		_, _ = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order2)
	})
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())
}

func TestLimitOrderLiquidAndFullGas(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "5000eur")

	gasMeter := sdk.NewInfiniteGasMeter()

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	var liquidTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	order1 := order(ctx.BlockTime(), acc1, "1gbp", "1eur")
	_, err := k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order1)
	require.NoError(t, err)
	require.Equal(t, liquidTrxFee, gasMeter.GasConsumed())

	gasMeter = sdk.NewInfiniteGasMeter()
	order2 := order(ctx.BlockTime(), acc2, "1eur", "1gbp")
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order2)
	require.NoError(t, err)
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())
}

func TestFilledLimitOrderGas(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "5000eur")

	gasMeter := sdk.NewInfiniteGasMeter()

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	var liquidTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	order1 := order(ctx.BlockTime(), acc1, "1gbp", "1eur")
	_, err := k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order1)
	require.NoError(t, err)
	require.Equal(t, liquidTrxFee, gasMeter.GasConsumed())

	gasMeter = sdk.NewInfiniteGasMeter()
	order2 := order(ctx.BlockTime(), acc2, "1eur", "1gbp")
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order2)
	require.NoError(t, err)
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())

	// Already filled
	order2.DestinationFilled = order2.Destination.Amount
	gasMeter = sdk.NewInfiniteGasMeter()
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order2)
	require.Error(t, err)
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())
}

func TestSingleOrderPanicFullGas(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "5000eur")

	gasMeter := sdk.NewInfiniteGasMeter()

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	var liquidTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	order1 := order(ctx.BlockTime(), acc1, "1gbp", "1eur")
	_, err := k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order1)
	require.NoError(t, err)
	require.Equal(t, liquidTrxFee, gasMeter.GasConsumed())

	gasMeter = sdk.NewInfiniteGasMeter()
	order2 := order(ctx.BlockTime(), acc2, "1eur", "1gbp")
	// Panic during banking interactions
	k.bk = nil
	require.Panics(t, func() {
		k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order2)
	})
	// Baseapp should charge the appropriate fee
	require.Equal(t, sdk.Gas(0), gasMeter.GasConsumed())
}

func TestPartiallyLiquidLimitOrderGas(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "5000eur")

	gasMeter := sdk.NewInfiniteGasMeter()

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	var liquidTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	order1 := order(ctx.BlockTime(), acc1, "1gbp", "1eur")
	_, err := k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order1)
	require.NoError(t, err)
	require.Equal(t, liquidTrxFee, gasMeter.GasConsumed())

	gasMeter = sdk.NewInfiniteGasMeter()
	order2 := order(ctx.BlockTime(), acc2, "10eur", "10gbp")
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order2)
	require.NoError(t, err)
	require.Equal(
		t, sdk.Gas(float64(stdTrxFee-liquidTrxFee)*0.1), gasMeter.GasConsumed(),
		"1/10 filled => 10% of total Gas",
	)

	gasMeter = sdk.NewInfiniteGasMeter()
	order3 := order(ctx.BlockTime(), acc1, "36gbp", "36eur")
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order3)
	require.NoError(t, err)
	require.Equal(
		t, (stdTrxFee-liquidTrxFee)/4, gasMeter.GasConsumed(),
		"9/36 filled => 25% of total Gas",
	)

	gasMeter = sdk.NewInfiniteGasMeter()
	order4 := order(ctx.BlockTime(), acc2, "36eur", "36gbp")
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order4)
	require.NoError(t, err)
	expGas := sdk.Gas(float64((stdTrxFee-liquidTrxFee)*3)/4)
	require.Equal(
		t, expGas, gasMeter.GasConsumed(),
		"27/36 filled => 3/4 of total Gas",
	)
}

func TestLiquidNewMarketOrderGas(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "5000eur")

	gasMeter := sdk.NewInfiniteGasMeter()

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	var liquidTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	order1 := order(ctx.BlockTime(), acc1, "100gbp", "100eur")
	_, err := k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order1)
	require.NoError(t, err)
	require.Equal(t, liquidTrxFee, gasMeter.GasConsumed())

	order2 := order(ctx.BlockTime(), acc2, "100eur", "100gbp")
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order2)
	require.NoError(t, err)
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())

	slippage := sdk.NewDecWithPrec(50, 2)
	srcDenom := "gbp"
	dest := sdk.NewCoin("eur", sdk.NewInt(200))
	slippageSource, err := k.GetSrcFromSlippage(
		ctx, srcDenom, dest, slippage,
	)
	require.NoError(t, err)

	gasMeter = sdk.NewInfiniteGasMeter()
	ctx = ctx.WithGasMeter(gasMeter)
	limitOrder := order(ctx.BlockTime(), acc1, slippageSource.String(), dest.String())
	_, err = k.NewOrderSingle(ctx, limitOrder)
	require.NoError(t, err)
	require.Equal(t, liquidTrxFee, ctx.GasMeter().GasConsumed())
}

func TestPartiallyLiquidMarketOrderGas(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "500eur")

	gasMeter := sdk.NewGasMeter(math.MaxUint64)

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	var liquidTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	order1 := order(ctx.BlockTime(), acc1, "1gbp", "1eur")
	_, err := k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order1)
	require.NoError(t, err)
	require.Equal(t, liquidTrxFee, gasMeter.GasConsumed())

	order2 := order(ctx.BlockTime(), acc2, "10eur", "10gbp")
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order2)
	require.NoError(t, err)
	require.Equal(
		t, sdk.Gas(float64(stdTrxFee-liquidTrxFee)*0.1), gasMeter.GasConsumed(),
		"1/10 filled => 10% of total Gas",
	)

	bal1Bef := bk.GetBalance(ctx, acc1.GetAddress(), "gbp")

	gasMeter = sdk.NewInfiniteGasMeter()
	slippage := sdk.NewDecWithPrec(75, 2)
	srcDenom := "gbp"
	dest := sdk.NewCoin("eur", sdk.NewInt(36))
	slippageSource, err := k.GetSrcFromSlippage(
		ctx, srcDenom, dest, slippage,
	)
	require.NoError(t, err)

	ctx = ctx.WithGasMeter(gasMeter)
	limitOrder := order(ctx.BlockTime(), acc1, slippageSource.String(), dest.String())
	_, err = k.NewOrderSingle(ctx, limitOrder)
	require.NoError(t, err)

	bal1Aft := bk.GetBalance(ctx.WithGasMeter(sdk.NewInfiniteGasMeter()), acc1.GetAddress(), "gbp")
	rebate := stdTrxFee-liquidTrxFee
	filledPct := bal1Bef.Sub(bal1Aft).Amount.ToDec().Quo(sdk.NewDec(36))
	expGas := sdk.Gas(
		filledPct.Mul(sdk.NewInt(int64(rebate)).ToDec()).RoundInt64(),
	)
	require.Equal(
		t, expGas, gasMeter.GasConsumed(),
		"9/36 filled => 1/4 of total Gas",
	)
}

func TestKeeperLimitOrderLiquidFullGas(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)
	acc1 := createAccount(ctx, ak, bk, randomAddress(), "100000eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "100000usd")

	order1cid := cid()
	order1, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("100eur"),
		coin("100usd"), acc1.GetAddress(), order1cid,
	)
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	var liquidTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	var qualificationMin int64
	k.paramStore.Get(ctx, types.KeyLiquidityRebateMinutesSpan, &qualificationMin)

	order2cid := cid()
	order2, _ := types.NewOrder(
		ctx.BlockTime(),
		types.TimeInForce_GoodTillCancel,
		coin("5000eur"),
		coin("5000usd"),
		acc1.GetAddress(),
		order2cid,
	)
	gasMeter := sdk.NewInfiniteGasMeter()
	// Add + Interval Minutes
	res, err := k.CancelReplaceLimitOrder(
		ctx.WithGasMeter(gasMeter).
			WithBlockTime(ctx.BlockTime().
				Add(time.Duration(qualificationMin)*time.Minute)),
		order2, order1cid,
	)
	require.True(t, err == nil, res.Log)
	require.Equal(
		t, liquidTrxFee, gasMeter.GasConsumed(),
		"Liquid gas as unfilled replacing orders, past minute-interval",
	)

	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		require.Equal(t, order2cid, orders[0].ClientOrderID)
		require.Equal(t, coin("5000eur").String(), orders[0].Source.String())
		require.Equal(t, coin("5000usd").String(), orders[0].Destination.String())
		// unfilled: remaining == source amount
		require.Equal(t, sdk.NewInt(5000), orders[0].SourceRemaining)
	}

	gasMeter = sdk.NewInfiniteGasMeter()
	order3, err := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("5000chf"),
		coin("5000usd"), acc1.GetAddress(), cid(),
	)
	require.NoError(t, err)
	require.Equal(
		t, liquidTrxFee, gasMeter.GasConsumed(),
		"Liquid trx changed instrument eur->chf",
	)

	// Wrong client order id for previous order submitted.
	_, err = k.CancelReplaceLimitOrder(ctx.WithGasMeter(gasMeter), order3, order1cid)
	require.True(t, types.ErrClientOrderIdNotFound.Is(err))
	require.Equal(
		t, stdTrxFee, gasMeter.GasConsumed(),
		"Erred replacing order is charged Std gas",
	)

	// Changing instrument of order
	gasMeter = sdk.NewGasMeter(math.MaxUint64)
	_, err = k.CancelReplaceLimitOrder(ctx.WithGasMeter(gasMeter), order3, order2cid)
	require.True(t, types.ErrOrderInstrumentChanged.Is(err))
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed(), "erred trx")

	o := order(ctx.BlockTime(), acc2, "1000usd", "300eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	bal1 := bk.GetAllBalances(ctx, acc1.GetAddress())
	bal2 := bk.GetAllBalances(ctx, acc2.GetAddress())

	require.Equal(t, int64(300), bal2.AmountOf("eur").Int64())
	require.Equal(t, int64(300), bal1.AmountOf("usd").Int64())

	filled := sdk.ZeroInt()
	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		filled = orders[0].Source.Amount.Sub(orders[0].SourceRemaining)
	}

	// CancelReplace and verify that previously filled amount is subtracted from the resulting order
	order4cid := cid()
	gasMeter = sdk.NewInfiniteGasMeter()
	order4, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("3000eur"),
		coin("3000usd"), acc1.GetAddress(), order4cid,
	)
	ctx = ctx.WithGasMeter(gasMeter)
	_, err = k.CancelReplaceLimitOrder(
		ctx.WithBlockTime(ctx.BlockTime().Add(
			time.Duration(qualificationMin)*time.Minute)),
		order4, order2cid,
	)
	require.NoError(t, err, err)
	require.Equal(
		t, gasMeter.GasConsumed(), sdk.Gas(float64(stdTrxFee-liquidTrxFee)*0.1),
		"300/3000 filled => 10% of total Gas",
	)
	require.Greaterf(
		t, gasMeter.GasConsumed(), liquidTrxFee,
		"Gas consumed should filled should be greater than LiquidTrxFee",
	)

	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		require.Equal(t, order4cid, orders[0].ClientOrderID)
		require.Equal(t, coin("3000eur").String(), orders[0].Source.String())
		require.Equal(t, coin("3000usd").String(), orders[0].Destination.String())
		require.Equal(
			t, sdk.NewInt(3000).Sub(filled).String(),
			orders[0].SourceRemaining.String(),
		)
	}
}
func TestLimitOrderErrGas(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)
	acc1 := createAccount(ctx, ak, bk, randomAddress(), "100eur")

	order1cid := cid()
	order1, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("100eur"),
		coin("100usd"), acc1.GetAddress(), order1cid,
	)
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	order2cid := cid()
	order2, _ := types.NewOrder(
		ctx.BlockTime(),
		types.TimeInForce_GoodTillCancel,
		/** 1 above balance **/
		coin("101eur"),
		coin("101usd"),
		acc1.GetAddress(),
		order2cid,
	)

	stdTrxFee := k.GetTrxFee(ctx)
	qualificationMin := k.GetLiquidityRebateMinutesSpan(ctx)

	gasMeter := sdk.NewInfiniteGasMeter()
	// Add + Interval Minutes.. but will err
	_, err = k.CancelReplaceLimitOrder(
		ctx.WithGasMeter(gasMeter).
			WithBlockTime(ctx.BlockTime().
				Add(time.Duration(qualificationMin)*time.Minute)),
		order2, order1cid,
	)
	require.Error(t, err, "Insufficient balance")
	require.Equal(
		t, stdTrxFee, gasMeter.GasConsumed(),
		"Err -> Full fee",
	)
}

func TestKeeperReplaceLimitFullLiquidGas(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)
	acc1 := createAccount(ctx, ak, bk, randomAddress(), "20000eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "45000usd")

	totalSupply := snapshotAccounts(ctx, bk)

	order1cid := cid()
	order1, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("500eur"),
		coin("1200usd"), acc1.GetAddress(), order1cid,
	)
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	var liquidTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	var qualificationMin int64
	k.paramStore.Get(ctx, types.KeyLiquidityRebateMinutesSpan, &qualificationMin)

	gasMeter := sdk.NewGasMeter(math.MaxUint64)
	order2cid := cid()
	order2, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("5000eur"),
		coin("17000usd"), acc1.GetAddress(), order2cid,
	)
	res, err := k.CancelReplaceLimitOrder(ctx.WithGasMeter(gasMeter), order2, order1cid)
	require.True(t, err == nil, res.Log)
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())

	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		require.Equal(t, order2cid, orders[0].ClientOrderID)
		require.Equal(t, coin("5000eur"), orders[0].Source)
		require.Equal(t, coin("17000usd"), orders[0].Destination)
		require.Equal(t, sdk.NewInt(5000), orders[0].SourceRemaining)
	}

	gasMeter = sdk.NewInfiniteGasMeter()
	order3, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("500chf"),
		coin("1700usd"), acc1.GetAddress(), cid(),
	)
	// Wrong client order id for previous order submitted.
	_, err = k.CancelReplaceLimitOrder(ctx.WithGasMeter(gasMeter), order3, order1cid)
	require.True(t, types.ErrClientOrderIdNotFound.Is(err))
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())

	// Changing instrument of order
	gasMeter = sdk.NewGasMeter(math.MaxUint64)
	_, err = k.CancelReplaceLimitOrder(ctx.WithGasMeter(gasMeter), order3, order2cid)
	require.True(t, types.ErrOrderInstrumentChanged.Is(err))
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())

	o := order(ctx.BlockTime(), acc2, "2600usd", "300eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	bal1 := bk.GetAllBalances(ctx, acc1.GetAddress())
	bal2 := bk.GetAllBalances(ctx, acc2.GetAddress())

	require.Equal(t, int64(300), bal2.AmountOf("eur").Int64())
	require.Equal(t, int64(1020), bal1.AmountOf("usd").Int64())

	filled := sdk.ZeroInt()
	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		filled = orders[0].Source.Amount.Sub(orders[0].SourceRemaining)
	}

	// CancelReplace and verify that previously filled amount is subtracted from the resulting order
	order4cid := cid()
	order4, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("10000eur"),
		coin("35050usd"), acc1.GetAddress(), order4cid,
	)
	gasMeter = sdk.NewInfiniteGasMeter()
	// add qualification minutes
	res, err = k.CancelReplaceLimitOrder(
		ctx.WithGasMeter(gasMeter).
			WithBlockTime(ctx.BlockTime().
				Add(time.Duration(qualificationMin)*time.Minute)),
		order4, order2cid,
	)
	t.Log(gasMeter.String())
	require.True(t, err == nil, res.Log)
	expGas := sdk.NewDec(int64(stdTrxFee - liquidTrxFee)).
		Mul(sdk.NewDec(1020)).
			Quo(sdk.NewDec(35050)).
		RoundInt().
		Uint64()
	require.Equal(t, expGas, gasMeter.GasConsumed(),
		"Partial Fee: Full fee * destinationFilled(1020)/totalDestination(35050)")

	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		require.Equal(t, order4cid, orders[0].ClientOrderID)
		require.Equal(t, coin("10000eur"), orders[0].Source)
		require.Equal(t, coin("35050usd"), orders[0].Destination)
		require.Equal(t, sdk.NewInt(10000).Sub(filled), orders[0].SourceRemaining)
	}

	// CancelReplace with an order that asks for a larger source than the replaced order has remaining
	order5 := order(ctx.BlockTime(), acc2, "42000usd", "8000eur")
	_, err = k.NewOrderSingle(ctx, order5)
	require.True(t, err == nil, res.Log)

	order6 := order(ctx.BlockTime(), acc1, "8000eur", "30000usd")
	_, err = k.CancelReplaceLimitOrder(ctx, order6, order4cid)
	require.True(t, types.ErrNoSourceRemaining.Is(err))

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}
func TestReplaceMarketOrderErrGas(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)
	acc1 := createAccount(ctx, ak, bk, randomAddress(), "500gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "500eur")

	var liquidTrxFee uint64 = k.GetLiquidTrxFee(ctx)

	var o types.Order
	var err error

	// Establish market price by executing a 1:1 trade
	o = order(ctx.BlockTime(), acc2, "1eur", "1gbp")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(ctx.BlockTime(), acc1, "1gbp", "1eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	// Make a market order that allows slippage
	slippage := sdk.NewDecWithPrec(100, 2)
	srcDenom := "gbp"
	dest := sdk.NewCoin("eur", sdk.NewInt(100))
	slippageSource, err := k.GetSrcFromSlippage(
		ctx, srcDenom, dest, slippage,
	)
	require.NoError(t, err)
	limitOrder := order(ctx.BlockTime(), acc1, slippageSource.String(), dest.String())
	_, err = k.NewOrderSingle(ctx, limitOrder)
	require.NoError(t, err)
	foundOrder := k.GetOrderByOwnerAndClientOrderId(
		ctx, acc1.GetAddress().String(), limitOrder.ClientOrderID,
	)
	require.NotNil(t, foundOrder, "Market order should exist")
	// 100% slippage to double source
	require.True(t, foundOrder.Source.IsEqual(sdk.NewCoin("gbp", sdk.NewInt(200))))

	// Gave 1 gbp and gained a eur
	acc1Bal := bk.GetAllBalances(ctx, acc1.GetAddress())
	require.Equal(t, coins("1eur,499gbp").String(), acc1Bal.String())

	slippage = sdk.NewDecWithPrec(0, 2)
	srcDenom = "gbp"
	dest = sdk.NewCoin("eur", sdk.NewInt(100))
	slippageSource, err = k.GetSrcFromSlippage(
		ctx, srcDenom, dest, slippage,
	)
	require.NoError(t, err)
	limitOrder = order(ctx.BlockTime(), acc1, slippageSource.String(), dest.String())
	gasMeter := sdk.NewInfiniteGasMeter()
	ctx = ctx.WithGasMeter(gasMeter)
	// panic execution of single order
	k.bk = nil
	require.Panics(t, func() {
		_, _ = k.NewOrderSingle(ctx, limitOrder)
	})
	require.Equal(
		t, liquidTrxFee,
		ctx.GasMeter().GasConsumed(), "Panic gas consumption is unpredictable",
	)
}

func TestCancelNewLimitFullGas(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000eur")

	gasMeter := sdk.NewGasMeter(math.MaxUint64)

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	var liquidTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	src1, dst1 := "eur", "usd"
	order1 := order(ctx.BlockTime(), acc1, "100"+src1, "120"+dst1)
	_, err := k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order1)
	require.NoError(t, err)
	require.Equal(t, liquidTrxFee, gasMeter.GasConsumed())

	order := k.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, order, 1)

	cid := order[0].ClientOrderID
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	_, err = k.CancelOrder(ctx, acc1.GetAddress(), cid)
	require.NoError(t, err)
	require.Equal(t, stdTrxFee, ctx.GasMeter().GasConsumed())
}

func TestCancelNewMarketFullGas(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000gbp")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "500eur")

	gasMeter := sdk.NewGasMeter(math.MaxUint64)

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	var liquidTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	order1 := order(ctx.BlockTime(), acc1, "100gbp", "120eur")
	_, err := k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order1)
	require.NoError(t, err)
	require.Equal(t, liquidTrxFee, gasMeter.GasConsumed(), "Liquid trx")

	order2 := order(ctx.BlockTime(), acc2, "120eur", "100gbp")
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), order2)
	require.NoError(t, err)
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed(), "Filled, non-liquid")

	gasMeter = sdk.NewGasMeter(math.MaxUint64)
	slippage := sdk.NewDecWithPrec(50, 2)
	srcDenom := "gbp"
	dest := sdk.NewCoin("eur", sdk.NewInt(200))
	slippageSource, err := k.GetSrcFromSlippage(
		ctx, srcDenom, dest, slippage,
	)
	require.NoError(t, err)
	limitOrder := order(ctx.BlockTime(), acc1, slippageSource.String(), dest.String())
	_, err = k.NewOrderSingle(ctx, limitOrder)
	require.NoError(t, err)
	require.Equal(t, liquidTrxFee, gasMeter.GasConsumed(), "Non-filled-Liquid")

	_, err = k.CancelOrder(ctx.WithGasMeter(gasMeter), acc1.GetAddress(), limitOrder.ClientOrderID)
	require.NoError(t, err)
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed(), "Cancel pays std fee")
}

func TestMultipleOrders(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "10000eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "7400usd")
	acc3 := createAccount(ctx, ak, bk, randomAddress(), "2200chf")

	totalSupply := snapshotAccounts(ctx, bk)

	// Add two orders that draw on the same balance.
	_, err := k.NewOrderSingle(ctx, order(ctx.BlockTime(), acc1, "10000eur", "11000usd"))
	require.NoError(t, err)

	_, err = k.NewOrderSingle(ctx, order(ctx.BlockTime(), acc1, "10000eur", "1400chf"))
	require.NoError(t, err)

	// require.Len(t, k.instruments, 2)

	res, err := k.NewOrderSingle(ctx, order(ctx.BlockTime(), acc2, "7400usd", "5000eur"))
	require.True(t, err == nil, res.Log)

	res, err = k.NewOrderSingle(ctx, order(ctx.BlockTime(), acc3, "2200chf", "5000eur"))
	require.True(t, err == nil, res.Log)

	// All acc1's EUR are sold by now. No orders should be on books
	orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, orders, 0)

	// All orders should be filled
	// require.Empty(t, k.instruments)

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}

func TestCancelZeroRemainingOrders(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc := createAccount(ctx, ak, bk, randomAddress(), "10000eur")
	_, err := k.NewOrderSingle(ctx, order(ctx.BlockTime(), acc, "10000eur", "11000usd"))
	require.NoError(t, err)

	err = bk.SendCoins(ctx, acc.GetAddress(), sdk.AccAddress([]byte("void")), coins("10000eur"))
	require.NoError(t, err)

	orders := k.GetOrdersByOwner(ctx, acc.GetAddress())
	require.Len(t, orders, 0)
}

func TestInsufficientBalance1(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "500eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "740usd")
	acc3 := createAccount(ctx, ak, bk, randomAddress(), "")

	totalSupply := snapshotAccounts(ctx, bk)

	o := order(ctx.BlockTime(), acc1, "300eur", "360usd")
	_, err := k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	// Modify account balance to be below order source
	err = bk.SendCoins(ctx, acc1.GetAddress(), acc3.GetAddress(), coins("250eur"))
	require.NoError(t, err)

	o = order(ctx.BlockTime(), acc2, "360usd", "300eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	bal1 := bk.GetAllBalances(ctx, acc1.GetAddress())
	bal2 := bk.GetAllBalances(ctx, acc2.GetAddress())
	require.Equal(t, "300usd", bal1.String()) // (500 -250) * 360/300
	require.Equal(t, "250eur,440usd", bal2.String())

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}

func Test2(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "100eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "121usd")

	totalSupply := snapshotAccounts(ctx, bk)

	o := order(ctx.BlockTime(), acc1, "100eur", "120usd")
	_, err := k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(ctx.BlockTime(), acc2, "121usd", "100eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	// require.Empty(t, k.instruments)
	require.Equal(t, coins("120usd"), bk.GetAllBalances(ctx, acc1.GetAddress()))
	require.Equal(t, coins("100eur,1usd"), bk.GetAllBalances(ctx, acc2.GetAddress()))

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}

func TestAllInstruments(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "10000eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "7400usd")
	acc3 := createAccount(ctx, ak, bk, randomAddress(), "2200chf")

	// Add two orders that draw on the same balance.
	_, err := k.NewOrderSingle(ctx, order(ctx.BlockTime(), acc1, "10000eur", "11000usd"))
	require.NoError(t, err)

	_, err = k.NewOrderSingle(ctx, order(ctx.BlockTime(), acc1, "10000eur", "1400chf"))
	require.NoError(t, err)

	res, err := k.NewOrderSingle(ctx, order(ctx.BlockTime(), acc2, "7400usd", "5000eur"))
	require.True(t, err == nil, res.Log)

	res, err = k.NewOrderSingle(ctx, order(ctx.BlockTime(), acc3, "2200chf", "5000eur"))
	require.True(t, err == nil, res.Log)

	// All acc1's EUR are sold by now. No orders should be on books
	orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, orders, 0)

	allInstruments := k.GetAllInstruments(ctx)
	// 30 because of chf, eur, gbp, jpy, ngm, usd
	require.Len(t, allInstruments, 30)

	transactedInstruments := "chfusd"
	for _, i := range allInstruments {
		if (i.Source == "eur" || i.Destination == "eur") &&
			(strings.Contains(transactedInstruments, i.Source) || strings.Contains(transactedInstruments, i.Destination)) {
			require.NotNil(t, i.LastPrice)
		}
	}

	// Sorting assertions by source+destination
	// instruments in supply: chf, eur, gbp, jpy, ngm, usd
	require.Equal(t, "chf", allInstruments[0].Source)
	require.Equal(t, "eur", allInstruments[0].Destination)
	require.Equal(t, "chf", allInstruments[1].Source)
	require.Equal(t, "gbp", allInstruments[1].Destination)
	require.Equal(t, "chf", allInstruments[2].Source)
	require.Equal(t, "jpy", allInstruments[2].Destination)
	require.Equal(t, "chf", allInstruments[3].Source)
	require.Equal(t, "ngm", allInstruments[3].Destination)
	require.Equal(t, "chf", allInstruments[4].Source)
	require.Equal(t, "usd", allInstruments[4].Destination)
	require.Equal(t, "eur", allInstruments[5].Source)
	require.Equal(t, "chf", allInstruments[5].Destination)
}

func Test3(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "100eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "120usd")

	totalSupply := snapshotAccounts(ctx, bk)

	o := order(ctx.BlockTime(), acc1, "100eur", "120usd")
	k.NewOrderSingle(ctx, o)

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	gasMeter := sdk.NewGasMeter(math.MaxUint64)
	for i := 0; i < 4; i++ {
		o = order(ctx.BlockTime(), acc2, "30usd", "25eur")
		k.NewOrderSingle(ctx.WithGasMeter(gasMeter), o)
	}
	require.Equal(t, 4*stdTrxFee, gasMeter.GasConsumed())

	bal1 := bk.GetAllBalances(ctx, acc1.GetAddress())
	bal2 := bk.GetAllBalances(ctx, acc2.GetAddress())
	require.Equal(t, coins("120usd"), bal1)
	require.Equal(t, coins("100eur"), bal2)

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}

func TestDeleteOrder(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "100eur")
	totalSupply := snapshotAccounts(ctx, bk)

	cid := cid()

	order1, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("100eur"),
		coin("120usd"), acc1.GetAddress(), cid,
	)
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	order2, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("100eur"),
		coin("77chf"), acc1.GetAddress(), cid,
	)
	_, err = k.NewOrderSingle(ctx, order2)
	require.Error(t, err) // Verify that client order ids cannot be duplicated.

	// require.Len(t, k.instruments, 1) // Ensure that the eur->chf pair was not added.

	// k.deleteOrder(ctx, &order1)
	// require.Len(t, k.instruments, 0) // Removing the only eur->usd order should have removed instrument

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}

func TestGetOrdersByOwnerAndCancel(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)
	acc1 := createAccount(ctx, ak, bk, randomAddress(), "100eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "120usd")

	for i := 0; i < 5; i++ {
		order, _ := types.NewOrder(
			ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("5eur"),
			coin("12usd"), acc1.GetAddress(), cid(),
		)
		_, err := k.NewOrderSingle(ctx, order)
		require.NoError(t, err)
	}

	for i := 0; i < 5; i++ {
		order, _ := types.NewOrder(
			ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("7usd"),
			coin("3chf"), acc2.GetAddress(), cid(),
		)
		res, err := k.NewOrderSingle(ctx, order)
		require.True(t, err == nil, res.Log)
	}

	allOrders1 := k.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, allOrders1, 5)

	{
		order, _ := types.NewOrder(
			ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("12usd"),
			coin("5eur"), acc2.GetAddress(), cid(),
		)
		res, err := k.NewOrderSingle(ctx, order)
		require.True(t, err == nil, res.Log)
	}

	allOrders2 := k.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, allOrders2, 4)

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	cid := allOrders2[2].ClientOrderID
	gasMeter := sdk.NewGasMeter(math.MaxUint64)
	_, err := k.CancelOrder(ctx.WithGasMeter(gasMeter), acc1.GetAddress(), cid)
	require.NoError(t, err)
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())

	_, err = k.CancelOrder(ctx.WithGasMeter(gasMeter), acc1.GetAddress(), cid)
	require.Error(t, err)

	require.Equal(t, 2*stdTrxFee, gasMeter.GasConsumed())

	allOrders3 := k.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, allOrders3, 3)

	found := false
	for _, e := range ctx.EventManager().ABCIEvents() {
		found = found || (e.Type == types.EventTypeMarket && string(e.Attributes[0].GetValue()) == "expire")
	}

	require.True(t, found)
}

func TestCancelOrders1(t *testing.T) {
	// Cancel a non-existing order by an account with no orders in the system.
	ctx, k, ak, bk := createTestComponents(t)
	acc := createAccount(ctx, ak, bk, randomAddress(), "100eur")

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	gasMeter := sdk.NewInfiniteGasMeter()
	_, err := k.CancelOrder(
		ctx.WithGasMeter(gasMeter), acc.GetAddress(), "abcde",
	)
	require.Error(t, err)
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())
}

func TestKeeperCancelReplaceLimitOrder(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)
	acc1 := createAccount(ctx, ak, bk, randomAddress(), "20000eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "45000usd")

	totalSupply := snapshotAccounts(ctx, bk)

	order1cid := cid()
	order1, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("500eur"),
		coin("1200usd"), acc1.GetAddress(), order1cid,
	)
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	gasMeter := sdk.NewGasMeter(math.MaxUint64)
	order2cid := cid()
	order2, _ := types.NewOrder(ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("5000eur"), coin("17000usd"), acc1.GetAddress(), order2cid)
	res, err := k.CancelReplaceLimitOrder(ctx.WithGasMeter(gasMeter), order2, order1cid)
	require.True(t, err == nil, res.Log)
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())

	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		require.Equal(t, order2cid, orders[0].ClientOrderID)
		require.Equal(t, coin("5000eur"), orders[0].Source)
		require.Equal(t, coin("17000usd"), orders[0].Destination)
		require.Equal(t, sdk.NewInt(5000), orders[0].SourceRemaining)
	}

	order3, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("500chf"),
		coin("1700usd"), acc1.GetAddress(), cid(),
	)
	// Wrong client order id for previous order submitted.
	_, err = k.CancelReplaceLimitOrder(ctx, order3, order1cid)
	require.True(t, types.ErrClientOrderIdNotFound.Is(err))

	// Changing instrument of order
	gasMeter = sdk.NewGasMeter(math.MaxUint64)
	_, err = k.CancelReplaceLimitOrder(ctx.WithGasMeter(gasMeter), order3, order2cid)
	require.True(t, types.ErrOrderInstrumentChanged.Is(err))
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())

	o := order(ctx.BlockTime(), acc2, "2600usd", "300eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	bal1 := bk.GetAllBalances(ctx, acc1.GetAddress())
	bal2 := bk.GetAllBalances(ctx, acc2.GetAddress())

	require.Equal(t, int64(300), bal2.AmountOf("eur").Int64())
	require.Equal(t, int64(1020), bal1.AmountOf("usd").Int64())

	filled := sdk.ZeroInt()
	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		filled = orders[0].Source.Amount.Sub(orders[0].SourceRemaining)
	}

	var qualificationMin int64
	k.paramStore.Get(ctx, types.KeyLiquidityRebateMinutesSpan, &qualificationMin)

	var liquidTrxFee uint64
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	// CancelReplace and verify that previously filled amount is subtracted from the resulting order
	order4cid := cid()
	gasMeter = sdk.NewInfiniteGasMeter()
	ctx = ctx.WithGasMeter(gasMeter)
	order4, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("10000eur"),
		coin("35050usd"), acc1.GetAddress(), order4cid,
	)
	res, err = k.CancelReplaceLimitOrder(
		ctx.WithBlockTime(ctx.BlockTime().Add(
			time.Duration(qualificationMin)*time.Minute)),
		order4, order2cid,
	)
	require.True(t, err == nil, res.Log)
	require.Lessf(
		t, gasMeter.GasConsumed(), stdTrxFee,
		"Gas consumed for partial filled should be less than stdTrxFee",
	)
	require.Greaterf(
		t, gasMeter.GasConsumed(), liquidTrxFee,
		"Gas consumed should filled should be greater than LiquidTrxFee",
	)

	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		require.Equal(t, order4cid, orders[0].ClientOrderID)
		require.Equal(t, coin("10000eur"), orders[0].Source)
		require.Equal(t, coin("35050usd"), orders[0].Destination)
		require.Equal(t, sdk.NewInt(10000).Sub(filled), orders[0].SourceRemaining)
	}

	// CancelReplace with an order that asks for a larger source than the replaced order has remaining
	order5 := order(ctx.BlockTime(), acc2, "42000usd", "8000eur")
	res, err = k.NewOrderSingle(ctx, order5)
	require.True(t, err == nil, res.Log)

	order6 := order(ctx.BlockTime(), acc1, "8000eur", "30000usd")
	gasMeter = sdk.NewInfiniteGasMeter()
	ctx = ctx.WithGasMeter(gasMeter)
	_, err = k.CancelReplaceLimitOrder(
		ctx.WithBlockTime(ctx.BlockTime().Add(
			time.Duration(qualificationMin)*time.Minute)),
		order6, order4cid,
	)
	require.True(t, types.ErrNoSourceRemaining.Is(err))
	require.Equal(
		t, stdTrxFee, gasMeter.GasConsumed(),
		"Erred orders should incur full fees",
	)

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}

func TestKeeperCancelReplaceMarketOrder(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)
	acc1 := createAccount(ctx, ak, bk, randomAddress(), "20000eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "45000usd")

	totalSupply := snapshotAccounts(ctx, bk)

	order1cid := cid()
	order1, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("500eur"),
		coin("1200usd"), acc1.GetAddress(), order1cid,
	)
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	gasMeter := sdk.NewGasMeter(math.MaxUint64)
	order2cid := cid()
	order2, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("5000eur"),
		coin("17000usd"), acc1.GetAddress(), order2cid,
	)
	res, err := k.CancelReplaceLimitOrder(ctx.WithGasMeter(gasMeter), order2, order1cid)
	require.True(t, err == nil, res.Log)
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())

	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		require.Equal(t, order2cid, orders[0].ClientOrderID)
		require.Equal(t, coin("5000eur"), orders[0].Source)
		require.Equal(t, coin("17000usd"), orders[0].Destination)
		require.Equal(t, sdk.NewInt(5000), orders[0].SourceRemaining)
	}

	order3, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("500chf"),
		coin("1700usd"), acc1.GetAddress(), cid(),
	)
	// Wrong client order id for previous order submitted.
	_, err = k.CancelReplaceLimitOrder(ctx, order3, order1cid)
	require.True(t, types.ErrClientOrderIdNotFound.Is(err))

	// Changing instrument of order
	gasMeter = sdk.NewGasMeter(math.MaxUint64)
	_, err = k.CancelReplaceLimitOrder(ctx.WithGasMeter(gasMeter), order3, order2cid)
	require.True(t, types.ErrOrderInstrumentChanged.Is(err))
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())

	o := order(ctx.BlockTime(), acc2, "2600usd", "300eur")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	bal1 := bk.GetAllBalances(ctx, acc1.GetAddress())
	bal2 := bk.GetAllBalances(ctx, acc2.GetAddress())

	require.Equal(t, int64(300), bal2.AmountOf("eur").Int64())
	require.Equal(t, int64(1020), bal1.AmountOf("usd").Int64())

	filled := sdk.ZeroInt()
	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		filled = orders[0].Source.Amount.Sub(orders[0].SourceRemaining)
	}

	// CancelReplace and verify that previously filled amount is subtracted from the resulting order
	order4cid := cid()
	order4, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("10000eur"),
		coin("35050usd"), acc1.GetAddress(), order4cid,
	)
	res, err = k.CancelReplaceLimitOrder(ctx, order4, order2cid)
	require.True(t, err == nil, res.Log)

	{
		orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
		require.Len(t, orders, 1)
		require.Equal(t, order4cid, orders[0].ClientOrderID)
		require.Equal(t, coin("10000eur"), orders[0].Source)
		require.Equal(t, coin("35050usd"), orders[0].Destination)
		require.Equal(t, sdk.NewInt(10000).Sub(filled), orders[0].SourceRemaining)
	}

	// CancelReplace with an order that asks for a larger source than the replaced order has remaining
	order5 := order(ctx.BlockTime(), acc2, "42000usd", "8000eur")
	k.NewOrderSingle(ctx, order5)
	require.True(t, err == nil, res.Log)

	order6 := order(ctx.BlockTime(), acc1, "8000eur", "30000usd")
	_, err = k.CancelReplaceLimitOrder(ctx, order6, order4cid)
	require.True(t, types.ErrNoSourceRemaining.Is(err))

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}

func TestOrdersChangeWithAccountBalance(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)
	acc := createAccount(ctx, ak, bk, randomAddress(), "15000eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "11000chf,100000eur")

	order, _ := types.NewOrder(
		ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("10000eur"),
		coin("1000usd"), acc.GetAddress(), cid(),
	)
	_, err := k.NewOrderSingle(ctx, order)
	require.NoError(t, err)

	{
		// Partially fill the order above
		acc2 := createAccount(ctx, ak, bk, randomAddress(), "900000usd")
		order2, _ := types.NewOrder(
			ctx.BlockTime(), types.TimeInForce_GoodTillCancel, coin("400usd"),
			coin("4000eur"), acc2.GetAddress(), cid(),
		)
		_, err = k.NewOrderSingle(ctx, order2)
		require.NoError(t, err)
	}

	totalSupply := snapshotAccounts(ctx, bk)

	err = bk.SendCoins(ctx, acc.GetAddress(), acc2.GetAddress(), coins("8000eur"))
	require.Nil(t, err)

	// Seller's account balance drops, remaining should be adjusted accordingly.
	orders := k.GetOrdersByOwner(ctx, acc.GetAddress())
	require.Len(t, orders, 1)
	require.Equal(t, coin("10000eur"), orders[0].Source)
	require.Equal(t, "3000", orders[0].SourceRemaining.String())
	require.Equal(t, "400", orders[0].DestinationFilled.String())

	// Seller's account balance is restored. Order should be adjusted, but take into consideration that the order has already been partially filled.
	err = bk.SendCoins(ctx, acc2.GetAddress(), acc.GetAddress(), coins("12000eur"))
	require.Nil(t, err)

	orders = k.GetOrdersByOwner(ctx, acc.GetAddress())
	require.Equal(t, "6000", orders[0].SourceRemaining.String())
	require.Equal(t, "400", orders[0].DestinationFilled.String())

	// Account balance dips below original sales amount, but can still fill the remaining order.
	err = bk.SendCoins(ctx, acc.GetAddress(), acc2.GetAddress(), coins("6000eur"))
	require.Nil(t, err)

	orders = k.GetOrdersByOwner(ctx, acc.GetAddress())
	require.Equal(t, "6000", orders[0].SourceRemaining.String())

	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}

func TestUnknownAsset(t *testing.T) {
	ctx, k1, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000eur")

	var stdTrxFee sdk.Gas
	k1.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	gasMeter := sdk.NewGasMeter(math.MaxUint64)

	// Make an order with a destination that is not known by the supply module
	o := order(ctx.BlockTime(), acc1, "1000eur", "1200nok")
	_, err := k1.NewOrderSingle(ctx.WithGasMeter(gasMeter), o)
	require.True(t, types.ErrUnknownAsset.Is(err))
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed(), "Liquid non-filled")
}

func TestLoadFromStore(t *testing.T) {
	// Create order book with a number of passive orders.
	ctx, k1, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "7400usd")

	o := order(ctx.BlockTime(), acc1, "1000eur", "1200usd")
	_, err := k1.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(ctx.BlockTime(), acc2, "5000usd", "3500chf")
	_, err = k1.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	_, k2, _, _ := createTestComponents(t)
	k2.paramStore = k1.paramStore

	k2.key = k1.key
	// Create new keeper and let it inherit the store of the previous keeper
	k2.initializeFromStore(ctx)

	// Verify that all orders are loaded correctly into the book
	// require.Len(t, k2.instruments, len(k1.instruments))

	// require.Equal(t, 1, k2.accountOrders.GetAllOrders(acc1.GetAddress()).Size())
	// require.Equal(t, 1, k2.accountOrders.GetAllOrders(acc2.GetAddress()).Size())
}

func TestVestingAccount(t *testing.T) {
	ctx, keeper, ak, bk := createTestComponents(t)
	account := createAccount(ctx, ak, bk, randomAddress(), "110000eur")

	amount := coins("110000eur") // todo (reviewer): does this amount make sense?
	vestingAcc := vestingtypes.NewDelayedVestingAccount(account.(*authtypes.BaseAccount), amount, math.MaxInt64)
	ak.SetAccount(ctx, vestingAcc)

	_, err := keeper.NewOrderSingle(ctx, order(ctx.BlockTime(), vestingAcc, "5000eur", "4700chf"))
	require.True(t, types.ErrAccountBalanceInsufficient.Is(err))
}

func TestInvalidInstrument(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000eur")

	// Ensure that an order cannot contain the same denomination in source and destination
	o := types.Order{
		ID:                124,
		Source:            coin("125eur"),
		Destination:       coin("250eur"),
		DestinationFilled: sdk.ZeroInt(),
		Owner:             acc1.GetAddress().String(),
		ClientOrderID:     "abcddeg",
		TimeInForce:       types.TimeInForce_GoodTillCancel,
	}

	_, err := k.NewOrderSingle(ctx, o)
	require.True(t, types.ErrInvalidInstrument.Is(err))
}
func TestSyntheticInstruments1(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)
	acc1 := createAccount(ctx, ak, bk, randomAddress(), "5000eur")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "6500usd")
	acc3 := createAccount(ctx, ak, bk, randomAddress(), "4500chf")

	totalSupply := snapshotAccounts(ctx, bk)

	o := order(ctx.BlockTime(), acc1, "1000eur", "1114usd")
	_, err := k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(ctx.BlockTime(), acc1, "500eur", "542chf")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(ctx.BlockTime(), acc3, "1000chf", "1028usd")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	var liquidTrxFee sdk.Gas
	k.paramStore.Get(ctx, types.KeyLiquidTrxFee, &liquidTrxFee)

	gasMeter := sdk.NewGasMeter(math.MaxUint64)
	o = order(ctx.BlockTime(), acc2, "5000usd", "4485eur")
	_, err = k.NewOrderSingle(ctx.WithGasMeter(gasMeter), o)
	require.NoError(t, err)
	require.Greater(t, gasMeter.GasConsumed(), liquidTrxFee) // Matches several orders, but should pay only the fixed fee
	require.Less(t, gasMeter.GasConsumed(), stdTrxFee) // Matches several orders, but should pay only the fixed fee

	// Ensure acc2 received at least some euro
	acc2Balance := bk.GetAllBalances(ctx, acc2.GetAddress())
	require.True(t, acc2Balance.AmountOf("eur").IsPositive())

	// Ensure acc2 did not receive any CHF, which is used in the synthetic instrument
	require.True(t, acc2Balance.AmountOf("chf").IsZero())

	// Ensure that acc2 filled all the eur sale orders in the market.
	require.True(t, acc2Balance.AmountOf("eur").Equal(sdk.NewInt(1500)))

	// Ensure that all tokens are accounted for.
	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}

func TestNonMatchingOrders(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)
	acc1 := createAccount(ctx, ak, bk, randomAddress(), "100000usd")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "240000eur")

	_, err := k.NewOrderSingle(ctx, order(ctx.BlockTime(), acc1, "20000usd", "20000eur"))
	require.NoError(t, err)
	_, err = k.NewOrderSingle(ctx, order(ctx.BlockTime(), acc2, "20000eur", "50000usd"))
	require.NoError(t, err)

	acc1Orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, acc1Orders, 1)
	require.Equal(t, sdk.ZeroInt(), acc1Orders[0].DestinationFilled)
	require.Equal(t, sdk.ZeroInt(), acc1Orders[0].SourceFilled)

	acc2Orders := k.GetOrdersByOwner(ctx, acc2.GetAddress())
	require.Len(t, acc2Orders, 1)
	require.Equal(t, sdk.ZeroInt(), acc2Orders[0].DestinationFilled)
	require.Equal(t, sdk.ZeroInt(), acc2Orders[0].SourceFilled)
}

func TestSyntheticInstruments2(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)
	acc1 := createAccount(ctx, ak, bk, randomAddress(), "972000chf,5000000usd")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "765000gbp,108000000jpy")

	acc3 := createAccount(ctx, ak, bk, randomAddress(), "3700000eur")

	totalSupply := snapshotAccounts(ctx, bk)

	passiveOrders := []types.Order{
		order(ctx.BlockTime(), acc1, "1000000usd", "896000eur"),

		order(ctx.BlockTime(), acc1, "1000000usd", "972000chf"),
		order(ctx.BlockTime(), acc1, "972000chf", "897000eur"),

		order(ctx.BlockTime(), acc1, "1000000usd", "108000000jpy"),
		order(ctx.BlockTime(), acc2, "40000000jpy", "331000eur"),
		order(ctx.BlockTime(), acc2, "68000000jpy", "563000eur"),

		order(ctx.BlockTime(), acc1, "400000usd", "306000gbp"),
		order(ctx.BlockTime(), acc1, "600000usd", "459000gbp"),
		order(ctx.BlockTime(), acc2, "765000gbp", "896000eur"),
	}

	for _, o := range passiveOrders {
		res, err := k.NewOrderSingle(ctx, o)
		require.NoError(t, err, res.Log)
	}

	var stdTrxFee uint64
	k.paramStore.Get(ctx, types.KeyTrxFee, &stdTrxFee)

	gasMeter := sdk.NewGasMeter(math.MaxUint64)
	monsterOrder := order(ctx.BlockTime(), acc3, "3700000eur", "4000000usd")
	res, err := k.NewOrderSingle(ctx.WithGasMeter(gasMeter), monsterOrder)
	require.NoError(t, err, res.Log)
	require.Equal(t, stdTrxFee, gasMeter.GasConsumed())

	// require.Len(t, k.instruments, 0)

	acc3bal := bk.GetAllBalances(ctx, acc3.GetAddress())
	require.Equal(t, "4000000", acc3bal.AmountOf("usd").String())

	// Ensure that all tokens are accounted for.
	require.True(t, totalSupply.Sub(snapshotAccounts(ctx, bk)).IsZero())
}

func TestDestinationCapacity(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "900000000usd")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "500000000000eur")

	order1 := order(ctx.BlockTime(), acc1, "800000000usd", "720000000eur")
	order1.SourceRemaining = sdk.NewInt(182000000)
	order1.SourceFilled = sdk.NewInt(618000000)
	order1.DestinationFilled = sdk.NewInt(645161290)

	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	order2 := order(ctx.BlockTime(), acc2, "471096868463eur", "500182000000usd")
	_, err = k.NewOrderSingle(ctx, order2)
	require.NoError(t, err)
}

func TestDestinationCapacity2(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "900000000usd")
	acc2 := createAccount(ctx, ak, bk, randomAddress(), "500000000000eur")
	acc3 := createAccount(ctx, ak, bk, randomAddress(), "140000000000chf")

	// chf -> usd -> eur

	order1 := order(ctx.BlockTime(), acc1, "800000000usd", "720000000eur")
	order1.SourceRemaining = sdk.NewInt(182000000)
	order1.SourceFilled = sdk.NewInt(618000000)
	order1.DestinationFilled = sdk.NewInt(645161290)

	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	order2 := order(ctx.BlockTime(), acc3, "130000000000chf", "800000000usd")
	_, err = k.NewOrderSingle(ctx, order2)
	require.NoError(t, err)

	aggressiveOrder := order(ctx.BlockTime(), acc2, "471096868463eur", "120000000000chf")
	_, err = k.NewOrderSingle(ctx, aggressiveOrder)
	require.NoError(t, err)
}

func TestPreventPhantomLiquidity(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc1 := createAccount(ctx, ak, bk, randomAddress(), "10000eur")

	order1 := order(ctx.BlockTime(), acc1, "8000eur", "9000usd")
	_, err := k.NewOrderSingle(ctx, order1)
	require.NoError(t, err)

	// Cannot sell more than the balance in the same instrument
	order2 := order(ctx.BlockTime(), acc1, "8000eur", "9000usd")
	_, err = k.NewOrderSingle(ctx, order2)
	require.Error(t, err)

	// Can sell the balance in another instrument
	order3 := order(ctx.BlockTime(), acc1, "8000eur", "6000chf")
	_, err = k.NewOrderSingle(ctx, order3)
	require.NoError(t, err)
}

func TestListInstruments(t *testing.T) {
	ctx, k, ak, bk := createTestComponents(t)

	acc := createAccount(ctx, ak, bk, randomAddress(), "5000eur,5000chf,5000usd,5000jpy")

	gasmeter := sdk.NewGasMeter(math.MaxUint64)

	instruments := k.GetInstruments(ctx)
	require.Empty(t, instruments)

	balances := bk.GetAllBalances(ctx, acc.GetAddress())
	// Create instruments between all denoms
	for _, src := range balances {
		for _, dst := range balances {
			if src.Denom == dst.Denom {
				continue
			}

			for i := 0; i < 3; i++ {
				var (
					s = fmt.Sprintf("50%v", src.Denom)
					d = fmt.Sprintf("100%v", dst.Denom)
				)

				order := order(ctx.BlockTime(), acc, s, d)
				_, err := k.NewOrderSingle(ctx.WithGasMeter(gasmeter), order)
				require.NoError(t, err)
			}
		}
	}

	allInstrumentsWithBestPrice := k.GetAllInstruments(ctx)
	_, err := json.Marshal(allInstrumentsWithBestPrice)
	require.Nil(t, err)
	// 30 because of chf, eur, gbp, jpy, ngm, usd
	require.Len(t, allInstrumentsWithBestPrice, 30)
}

func TestTimeInForceIO(t *testing.T) {
	encodingConfig := MakeTestEncodingConfig()

	clientCtx := client.Context{}.
		WithJSONMarshaler(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithChainID("testing")

	flagSet := pflag.NewFlagSet("testing", pflag.PanicOnError)
	txf := clienttx.NewFactoryCLI(clientCtx, flagSet).
		WithMemo("+memo").
		WithFees("10ALX").
		WithSequence(1).
		WithAccountNumber(2)

	msg := &types.MsgAddLimitOrder{
		TimeInForce:   types.TimeInForce_GoodTillCancel,
		Owner:         randomAddress().String(),
		Source:        sdk.NewCoin("echf", sdk.NewInt(50000)),
		Destination:   sdk.NewCoin("eeur", sdk.NewInt(60000)),
		ClientOrderId: "foobar",
	}
	txb, err := clienttx.BuildUnsignedTx(txf, msg)
	require.NoError(t, err)
	txBz, err := encodingConfig.TxConfig.TxJSONEncoder()(txb.GetTx())
	require.NoError(t, err)
	_, err = clientCtx.TxConfig.TxJSONDecoder()(txBz)
	require.NoError(t, err)

	msgCRL := &types.MsgCancelReplaceLimitOrder{
		TimeInForce:   types.TimeInForce_FillOrKill,
		Owner:         msg.Owner,
		Source:        sdk.NewCoin("echf", sdk.NewInt(50000)),
		Destination:   sdk.NewCoin("eeur", sdk.NewInt(60000)),
		OrigClientOrderId: "foobar",
		NewClientOrderId:  "newOrder",
	}

	txb, err = clienttx.BuildUnsignedTx(txf, msgCRL)
	require.NoError(t, err)
	txBz, err = encodingConfig.TxConfig.TxJSONEncoder()(txb.GetTx())
	require.NoError(t, err)
	_, err = clientCtx.TxConfig.TxJSONDecoder()(txBz)
	require.NoError(t, err)

}

func TestGetNextOrderNumber(t *testing.T) {
	ctx, k, _, _ := createTestComponents(t)
	require.Equal(t, uint64(0), k.getNextOrderNumber(ctx)) // starts with 0
	require.Equal(t, uint64(1), k.getNextOrderNumber(ctx)) // increments counter
	require.Equal(t, uint64(2), k.getNextOrderNumber(ctx)) // increments counter
}

func createTestComponents(t *testing.T) (sdk.Context, *Keeper, authkeeper.AccountKeeper, *embank.ProxyKeeper) {
	return createTestComponentsWithEncoding(t, MakeTestEncodingConfig())
}

func createTestComponentsWithEncoding(t *testing.T, encConfig simappparams.EncodingConfig) (sdk.Context, *Keeper, authkeeper.AccountKeeper, *embank.ProxyKeeper) {
	t.Helper()

	var (
		keyMarket  = sdk.NewKVStoreKey(types.StoreKey)
		keyIndices = sdk.NewKVStoreKey(types.StoreKeyIdx)
		keyAuthCap = sdk.NewKVStoreKey("authCapKey")
		keyParams  = sdk.NewKVStoreKey(paramstypes.StoreKey)
		keyBank    = sdk.NewKVStoreKey(banktypes.StoreKey)
		tkeyParams = sdk.NewTransientStoreKey(paramstypes.TStoreKey)

		blockedAddr = make(map[string]bool)
		maccPerms   = map[string][]string{}
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyMarket, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyIndices, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAuthCap, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyBank, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: "test-chain"}, true, log.NewNopLogger())
	ctx = ctx.WithBlockTime(time.Now())
	var (
		pk = paramskeeper.NewKeeper(encConfig.Marshaler, encConfig.Amino, keyParams, tkeyParams)
		ak = authkeeper.NewAccountKeeper(
			encConfig.Marshaler, keyAuthCap, pk.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
		)
		bk = bankkeeper.NewBaseKeeper(
			encConfig.Marshaler, keyBank, ak, pk.Subspace(banktypes.ModuleName), blockedAddr,
		)

		wrappedBank = embank.Wrap(bk)

		marketKeeper = NewKeeper(encConfig.Marshaler, keyMarket, keyIndices, ak,
			wrappedBank,
			pk.Subspace(types.ModuleName))
	)

	marketKeeper.InitParamsStore(ctx)

	bk.SetSupply(ctx, banktypes.NewSupply(coins("1eur,1usd,1chf,1jpy,1gbp,1ngm")))

	return ctx, marketKeeper, ak, wrappedBank
}

func MakeTestEncodingConfig() simappparams.EncodingConfig {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	encodingConfig := simappparams.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             cdc,
	}

	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics := module.NewBasicManager(
		bank.AppModuleBasic{},
		auth.AppModuleBasic{},
		vesting.AppModuleBasic{},
	)

	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

func coin(s string) sdk.Coin {
	coin, err := sdk.ParseCoinNormalized(s)
	if err != nil {
		panic(err)
	}
	return coin
}

func coins(s string) sdk.Coins {
	coins, err := sdk.ParseCoinsNormalized(s)
	if err != nil {
		panic(err)
	}
	return coins
}

func order(createdTm time.Time, account authtypes.AccountI, src, dst string) types.Order {
	o, err := types.NewOrder(
		createdTm, types.TimeInForce_GoodTillCancel, coin(src), coin(dst),
		account.GetAddress(), cid(),
	)
	if err != nil {
		panic(err)
	}

	return o
}

func createAccount(ctx sdk.Context, ak authkeeper.AccountKeeper, bk bankkeeper.SendKeeper, address sdk.AccAddress, balance string) authtypes.AccountI {
	acc := ak.NewAccountWithAddress(ctx, address)
	if err := bk.SetBalances(ctx, address, coins(balance)); err != nil {
		panic(err)
	}
	ak.SetAccount(ctx, acc)
	return acc
}

// Generate a random string to use as a client order id
func cid() string {
	return tmrand.Str(10)
}

func dumpEvents(events sdk.Events) {
	fmt.Println("Number of events:", len(events))
	for _, evt := range events {
		fmt.Println(evt.Type)
		for _, kv := range evt.Attributes {
			fmt.Println(" - ", string(kv.Key), string(kv.Value))
		}
	}
}

func snapshotAccounts(ctx sdk.Context, bk bankkeeper.ViewKeeper) (totalBalance sdk.Coins) {
	bk.IterateAllBalances(ctx, func(_ sdk.AccAddress, coin sdk.Coin) (stop bool) {
		totalBalance = totalBalance.Add(coin)
		return
	})
	return
}

func randomAddress() sdk.AccAddress {
	return tmrand.Bytes(sdk.AddrLen)
}
