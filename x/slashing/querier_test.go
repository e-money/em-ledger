package slashing

import (
	"github.com/tendermint/tendermint/libs/db"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

func TestNewQuerier(t *testing.T) {
	ctx, _, _, _, keeper, _ := createTestInput(t, keeperTestParams(), db.NewMemDB())
	querier := NewQuerier(keeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(ctx, []string{"parameters"}, query)
	require.NoError(t, err)
}

func TestQueryParams(t *testing.T) {
	cdc := codec.New()
	ctx, _, _, _, keeper, _ := createTestInput(t, keeperTestParams(), db.NewMemDB())

	var params Params

	res, errRes := queryParams(ctx, keeper)
	require.NoError(t, errRes)

	err := cdc.UnmarshalJSON(res, &params)
	require.NoError(t, err)
	require.Equal(t, keeper.GetParams(ctx), params)
}
