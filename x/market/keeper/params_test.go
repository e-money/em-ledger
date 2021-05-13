package keeper

import (
	"github.com/e-money/em-ledger/x/market/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParams(t *testing.T) {
	ctx, k, _, _ := createTestComponents(t)
	expParams := types.DefaultTxParams()

	params := k.GetParams(ctx)
	require.Equal(t, expParams, params)

	expParams.TrxFee = 100_000
	expParams.LiquidTrxFee = 250
	expParams.LiquidityRebateMinutesSpan = 10

	k.SetParams(ctx, expParams)
	params = k.GetParams(ctx)

	require.Equal(t, expParams, params)
}

