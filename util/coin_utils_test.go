package util

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestSplit1(t *testing.T) {
	coins := sdk.NewCoins(
		sdk.NewCoin("coin1", sdk.OneInt()),
		sdk.NewCoin("coin2", sdk.OneInt()),
		sdk.NewCoin("coin3", sdk.OneInt()),
		sdk.NewCoin("coin4", sdk.OneInt()),
	)

	splitCoins, remaining := SplitCoinsByDenom(coins, "coin2", "coin3", "coins5")
	require.Equal(t, sdk.OneInt(), splitCoins.AmountOf("coin2"))
	require.Equal(t, sdk.OneInt(), splitCoins.AmountOf("coin3"))

	require.Equal(t, sdk.OneInt(), remaining.AmountOf("coin1"))
	require.Equal(t, sdk.OneInt(), remaining.AmountOf("coin4"))

	require.Equal(t, sdk.ZeroInt(), splitCoins.AmountOf("coin5"))
	require.Equal(t, sdk.ZeroInt(), remaining.AmountOf("coin5"))
}

func TestSplit2(t *testing.T) {
	coins := sdk.NewCoins()

	splitCoins, remaining := SplitCoinsByDenom(coins, "coin1", "coin2")
	require.True(t, splitCoins.IsZero())
	require.True(t, remaining.IsZero())
}
