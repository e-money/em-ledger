package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestLegacyQuerier(t *testing.T) {
	authority := mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
	gp, _ := sdk.ParseDecCoins("0.400000000000000000echf,0.400000000000000000eeur")

	ctx, k, _, _ := createTestComponents(t)

	k.BootstrapAuthority(ctx, authority)
	res, err := k.SetGasPrices(ctx, authority, gp)
	require.True(t, err == nil, res.Log)

	prices, err := queryGasPrices(ctx, k)
	require.NoError(t, err)
	json := gjson.ParseBytes(prices)
	require.True(t, json.Get("min_gas_prices").IsArray())
	require.Len(t, json.Get("min_gas_prices").Array(), 2)
}
