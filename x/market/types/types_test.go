// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSerialization(t *testing.T) {
	// Verify that non-public fields survive de-/serialization
	order1, _ := NewOrder(coin("100eur"), coin("120usd"), sdk.AccAddress([]byte("acc1")), "A")
	order1.ID = 3123
	order1.SourceRemaining = sdk.NewInt(50)
	order1.SourceFilled = sdk.NewInt(50)

	bz, err := ModuleCdc.MarshalBinaryBare(order1)
	require.NoError(t, err)

	require.True(t, len(bz) > 0)

	order2 := Order{}
	err = ModuleCdc.UnmarshalBinaryBare(bz, &order2)
	require.NoError(t, err)

	// Some sanity checks to ensure we're not just comparing default values below.
	require.True(t, order2.Source.Amount.Int64() > 0)
	require.True(t, order2.Destination.Amount.Int64() > 0)
	require.True(t, order2.SourceRemaining.Int64() > 0)
	require.True(t, order2.SourceFilled.Int64() > 0)
	require.True(t, order2.price.GT(sdk.ZeroDec()))
	require.True(t, order2.invertedPrice.GT(sdk.ZeroDec()))

	require.Equal(t, uint64(3123), order2.ID)
	require.Equal(t, order1.ID, order2.ID)
	require.Equal(t, order1.Source, order2.Source)
	require.Equal(t, order1.Destination, order2.Destination)
	require.Equal(t, sdk.NewInt(50), order2.SourceRemaining)
	require.Equal(t, order1.SourceRemaining, order2.SourceRemaining)
	require.Equal(t, sdk.NewInt(50), order2.SourceFilled)
	require.Equal(t, order1.SourceFilled, order2.SourceFilled)
	require.Equal(t, order1.price, order2.price)
	require.Equal(t, order1.invertedPrice, order2.invertedPrice)
}

func TestOrders1(t *testing.T) {
	acc1 := sdk.AccAddress([]byte("acc1"))
	acc2 := sdk.AccAddress([]byte("acc2"))

	orders := NewOrders()
	order1, _ := NewOrder(coin("100eur"), coin("120usd"), acc1, "A")
	order1.ID = 1

	order2, _ := NewOrder(coin("250usd"), coin("180chf"), acc2, "A")
	order2.ID = 2

	orders.AddOrder(&order1)
	orders.AddOrder(&order2)

	require.True(t, orders.ContainsClientOrderId(acc1, "A"))
	require.True(t, orders.ContainsClientOrderId(acc2, "A"))

	require.NotNil(t, orders.GetOrder(acc1, "A"))
	require.NotNil(t, orders.GetOrder(acc2, "A"))
	require.Nil(t, orders.GetOrder(acc1, "B"))

}

func TestInvalidOrder(t *testing.T) {
	// 0 amount source
	_, err := NewOrder(coin("0eur"), coin("120usd"), []byte("acc"), "A")
	require.Error(t, err)

	// Same denomination
	_, err = NewOrder(coin("1000eur"), coin("850eur"), []byte("acc"), "A")
	require.Error(t, err)

	c := sdk.Coin{
		Denom:  "eur",
		Amount: sdk.NewInt(-100),
	}

	// Negative source
	_, err = NewOrder(c, coin("120usd"), []byte("acc"), "B")
	require.Error(t, err)
}

func TestComparator(t *testing.T) {
	order1, _ := NewOrder(coin("100eur"), coin("120usd"), []byte("acc1"), "A")
	order1.ID = 1

	order2, _ := NewOrder(coin("100eur"), coin("100usd"), []byte("acc2"), "A")
	order2.ID = 2

	require.True(t, OrderPriorityComparator(&order1, &order2) > 0)
	require.True(t, OrderPriorityComparator(&order2, &order1) < 0)

	require.True(t, OrderPriorityComparator(&order1, &order1) == 0)
	require.True(t, OrderPriorityComparator(&order2, &order2) == 0)
}

func TestOrderClientIdComparator(t *testing.T) {
	order1, _ := NewOrder(coin("100eur"), coin("120usd"), []byte("acc1"), "A")
	order1.ID = 1

	order2, _ := NewOrder(coin("100eur"), coin("100usd"), []byte("acc1"), "B")
	order2.ID = 2

	require.True(t, OrderClientIdComparator(&order1, &order2) < 0)
	require.True(t, OrderClientIdComparator(&order2, &order1) > 0)
	require.True(t, OrderClientIdComparator(&order1, &order1) == 0)
	require.True(t, OrderClientIdComparator(&order2, &order2) == 0)
}

func coin(s string) sdk.Coin {
	coin, err := sdk.ParseCoin(s)
	if err != nil {
		panic(err)
	}
	return coin
}
