package keeper

import (
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"
)

func TestQuerier1(t *testing.T) {
	ctx, k, ak, _ := createTestComponents(t)

	acc1 := createAccount(ctx, ak, "acc1", "5000eur,2500chf")

	o := order(acc1, "100eur", "120usd")
	require.True(t, k.NewOrderSingle(ctx, o).IsOK())

	o = order(acc1, "72eur", "1213jpy")
	require.True(t, k.NewOrderSingle(ctx, o).IsOK())

	o = order(acc1, "72chf", "312usd")
	require.True(t, k.NewOrderSingle(ctx, o).IsOK())

	bz, err := queryInstruments(ctx, k)
	require.NoError(t, err)
	json := gjson.ParseBytes(bz)
	instr := json.Get("Instruments")
	require.True(t, instr.IsArray())
	require.Len(t, instr.Array(), 3)
}
