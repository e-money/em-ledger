// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOrders1(t *testing.T) {

	acc1 := sdk.AccAddress([]byte("acc1"))
	acc2 := sdk.AccAddress([]byte("acc2"))

	orders := NewOrders()
	order1 := NewOrder(coin("100eur"), coin("120usd"), acc1, "A")
	order1.ID = 1

	order2 := NewOrder(coin("100eur"), coin("100usd"), acc2, "A")
	order2.ID = 2

	orders.AddOrder(order1)
	orders.AddOrder(order2)

	require.True(t, orders.ContainsClientOrderId(acc1, "A"))
	require.True(t, orders.ContainsClientOrderId(acc2, "A"))

}

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

func TestOrderClientIdComparator(t *testing.T) {
	order1 := NewOrder(coin("100eur"), coin("120usd"), sdk.AccAddress([]byte("acc1")), "A")
	order1.ID = 1

	order2 := NewOrder(coin("100eur"), coin("100usd"), sdk.AccAddress([]byte("acc1")), "B")
	order2.ID = 2

	require.True(t, OrderClientIdComparator(order1, order2) < 0)
	require.True(t, OrderClientIdComparator(order2, order1) > 0)
	require.True(t, OrderClientIdComparator(order1, order1) == 0)
	require.True(t, OrderClientIdComparator(order2, order2) == 0)
}

func coin(s string) sdk.Coin {
	coin, err := sdk.ParseCoin(s)
	if err != nil {
		panic(err)
	}
	return coin
}
