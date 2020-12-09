// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tidwall/gjson"
)

func TestQryGetAllInstruments(t *testing.T) {
	ctx, k, ak, _, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000eur,2500chf,400ngm")
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
