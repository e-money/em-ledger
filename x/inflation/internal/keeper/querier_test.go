package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"emoney/x/inflation/internal/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestNewQuerier(t *testing.T) {
	input := newTestInput(t)
	querier := NewQuerier(input.mintKeeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(input.ctx, []string{types.QueryInflation}, query)
	require.NoError(t, err)

	_, err = querier(input.ctx, []string{"foo"}, query)
	require.Error(t, err)
}

func TestQueryInflation(t *testing.T) {
	input := newTestInput(t)

	var inflation types.InflationState

	res, sdkErr := queryInflation(input.ctx, input.mintKeeper)
	require.NoError(t, sdkErr)

	err := input.cdc.UnmarshalJSON(res, &inflation)

	require.NoError(t, err)

	require.True(t, len(inflation.InflationAssets) > 0)
}
