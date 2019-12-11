package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestComparator(t *testing.T) {
	order1 := NewOrder(coin("100eur"), coin("120usd"), sdk.AccAddress([]byte("acc1")), "A")
	order1.ID = 1

	order2 := NewOrder(coin("100eur"), coin("100usd"), sdk.AccAddress([]byte("acc2")), "A")
	order2.ID = 2

	require.True(t, OrderPriorityComparator(order1, order2) > 0)
	require.True(t, OrderPriorityComparator(order2, order1) < 0)

	require.True(t, OrderPriorityComparator(order1, order1) == 0)
	require.True(t, OrderPriorityComparator(order2, order2) == 0)
}

func coin(s string) sdk.Coin {
	coin, err := sdk.ParseCoin(s)
	if err != nil {
		panic(err)
	}
	return coin
}
