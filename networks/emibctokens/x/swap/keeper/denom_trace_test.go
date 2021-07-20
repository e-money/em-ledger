package keeper

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/e-money/stargate/networks/emibctokens/x/swap/types"
)

func createNDenomTrace(keeper *Keeper, ctx sdk.Context, n int) []types.DenomTrace {
	items := make([]types.DenomTrace, n)
	for i := range items {
		items[i].Creator = "any"
		items[i].Index = fmt.Sprintf("%d", i)
		keeper.SetDenomTrace(ctx, items[i])
	}
	return items
}

func TestDenomTraceGet(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	items := createNDenomTrace(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetDenomTrace(ctx, item.Index)
		assert.True(t, found)
		assert.Equal(t, item, rst)
	}
}
func TestDenomTraceRemove(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	items := createNDenomTrace(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveDenomTrace(ctx, item.Index)
		_, found := keeper.GetDenomTrace(ctx, item.Index)
		assert.False(t, found)
	}
}

func TestDenomTraceGetAll(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	items := createNDenomTrace(keeper, ctx, 10)
	assert.Equal(t, items, keeper.GetAllDenomTrace(ctx))
}
