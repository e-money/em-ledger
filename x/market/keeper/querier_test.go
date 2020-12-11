// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"strings"
	"testing"
	"time"

	json2 "encoding/json"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tidwall/gjson"
)

func TestQryGetAllInstrumentsWithNonZeroBestPrices(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000eur,2500chf,400ngm")
	acc2 := createAccount(ctx, ak, "acc2", "1000usd")

	// generate passive order
	o := order(acc1, "100eur", "120usd")
	_, err := k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	// generate passive order
	o = order(acc1, "72eur", "1213jpy")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	// generate passive order of half balances
	o = order(acc1, "72chf", "312usd")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	// Execute an order
	// fulfilled
	o = order(acc2, "156usd", "36chf")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	{
		bz, err := queryInstruments(ctx, k)
		require.NoError(t, err)
		json := gjson.ParseBytes(bz)
		instr := json.Get("instruments")
		require.True(t, instr.IsArray())
		var instrLst []QueryInstrumentsResponse
		err = json2.Unmarshal([]byte(instr.String()), &instrLst)
		require.Nil(t, err, "Unmarshal from instruments response")

		bestPriced := 0
		for _, instrResp := range instrLst {
			// for the 3 passive orders above
			if (instrResp.Source == "jpy" && instrResp.Destination == "eur") ||
				(instrResp.Source == "usd" && instrResp.Destination == "chf") ||
				(instrResp.Source == "usd" && instrResp.Destination == "eur") {
				require.False(t, instrResp.BestPrice.IsZero())
				bestPriced++
			} else {
				require.Nil(t, instrResp.BestPrice)
			}
		}
		require.Equal(t, bestPriced, 3, "3 passive orders")

		// 30 because of chf, eur, gbp, jpy, ngm, usd
		require.Len(t, instr.Array(), 30)
	}
}

func TestQryGetAllInstrumentsWithNilBestPrices(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "10000eur")
	acc2 := createAccount(ctx, ak, "acc2", "7400usd")
	acc3 := createAccount(ctx, ak, "acc3", "2200chf")

	// generate passive order
	_, err := k.NewOrderSingle(ctx, order(acc1, "10000eur", "11000usd"))
	require.NoError(t, err)

	_, err = k.NewOrderSingle(ctx, order(acc1, "10000eur", "1400chf"))
	require.NoError(t, err)

	res, err := k.NewOrderSingle(ctx, order(acc2, "7400usd", "5000eur"))
	require.True(t, err == nil, res.Log)

	res, err = k.NewOrderSingle(ctx, order(acc3, "2200chf", "5000eur"))
	require.True(t, err == nil, res.Log)

	// All acc1's EUR are sold by now. No orders should be on books
	orders := k.GetOrdersByOwner(ctx, acc1.GetAddress())
	require.Len(t, orders, 0)

	allInstruments := k.GetAllInstruments(ctx)
	// 30 because of chf, eur, gbp, jpy, ngm, usd
	require.Len(t, allInstruments, 30)

	{
		bz, err := queryInstruments(ctx, k)
		require.NoError(t, err)
		json := gjson.ParseBytes(bz)
		instr := json.Get("instruments")
		require.True(t, instr.IsArray())
		var instrLst []QueryInstrumentsResponse
		err = json2.Unmarshal([]byte(instr.String()), &instrLst)
		require.Nil(t, err, "Unmarshal from instruments response")

		transactedInstruments := "chfusd"
		for _, instrResp := range instrLst {
			if (instrResp.Source == "eur" || instrResp.Destination == "eur") &&
				(strings.Contains(transactedInstruments, instrResp.Source) || strings.Contains(transactedInstruments, instrResp.Destination)) {
				require.NotNil(t, instrResp.LastPrice)
				if instrResp.BestPrice != nil {
					require.False(t, instrResp.BestPrice.IsZero())
				}
			}
			// No unfulfilled orders
			require.Nil(t, instrResp.BestPrice)
		}

		// 30 because of chf, eur, gbp, jpy, ngm, usd
		require.Len(t, instr.Array(), 30)
	}
}

func TestQuerier1(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000eur,2500chf")
	acc2 := createAccount(ctx, ak, "acc2", "1000usd")

	o := order(acc1, "100eur", "120usd")
	_, err := k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(acc1, "72eur", "1213jpy")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	o = order(acc1, "72chf", "312usd")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	// Execute an order
	o = order(acc2, "156usd", "36chf")
	_, err = k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	{
		bz, err := queryInstruments(ctx, k)
		require.NoError(t, err)
		json := gjson.ParseBytes(bz)
		instr := json.Get("instruments")
		require.True(t, instr.IsArray())
		// An instrument is registered for both order directions
		// 30 because of chf, eur, gbp, jpy, ngm, usd
		require.Len(t, instr.Array(), 30)

		// Check that timestamps are included on instruments where trades have occurred
		tradedTimestamps := json.Get("instruments.#.last_traded")
		require.Len(t, tradedTimestamps.Array(), 2)

		// Verify that timestamps match the blocktime.
		require.Equal(t, tradedTimestamps.Array()[0].Str, ctx.BlockTime().Format(time.RFC3339Nano))
		require.Equal(t, tradedTimestamps.Array()[1].Str, ctx.BlockTime().Format(time.RFC3339Nano))
	}
	{
		bz, err := queryInstrument(ctx, k, []string{"eur", "usd"}, abci.RequestQuery{})
		require.NoError(t, err)

		json := gjson.ParseBytes(bz)
		require.Equal(t, "eur", json.Get("source").Str)
		require.Equal(t, "usd", json.Get("destination").Str)

		orders := json.Get("orders")
		require.True(t, orders.IsArray())
		require.Len(t, orders.Array(), 1)
	}
}
