package keeper

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/e-money/stargate/networks/emibctokens/x/swap/types"
)

func createNIbcToken(keeper *Keeper, ctx sdk.Context, n int) []types.IbcToken {
	items := make([]types.IbcToken, n)
	for i := range items {
		items[i].Creator = "any"
		items[i].Index = fmt.Sprintf("%d", i)
		keeper.SetIbcToken(ctx, items[i])
	}
	return items
}

func TestIbcTokenGet(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	items := createNIbcToken(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetIbcToken(ctx, item.Index)
		assert.True(t, found)
		assert.Equal(t, item, rst)
	}
}
func TestIbcTokenRemove(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	items := createNIbcToken(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveIbcToken(ctx, item.Index)
		_, found := keeper.GetIbcToken(ctx, item.Index)
		assert.False(t, found)
	}
}

func TestIbcTokenGetAll(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	items := createNIbcToken(keeper, ctx, 10)
	assert.Equal(t, items, keeper.GetAllIbcToken(ctx))
}
