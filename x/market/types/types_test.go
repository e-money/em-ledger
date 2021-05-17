// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestSerialization(t *testing.T) {
	// Verify that non-public fields survive de-/serialization
	tm := time.Now()
	order1, _ := NewOrder(
		tm, TimeInForce_GoodTillCancel, coin("100eur"), coin("120usd"),
		[]byte("acc1"), "A", time.Time{},
	)
	order1.ID = 3123
	order1.SourceRemaining = sdk.NewInt(50)
	order1.SourceFilled = sdk.NewInt(10)
	order1.DestinationFilled = sdk.NewInt(50)
	order1.TimeInForce = TimeInForce_FillOrKill

	bz, err := ModuleCdc.MarshalBinaryBare(&order1)
	require.NoError(t, err)

	require.True(t, len(bz) > 0)

	order2 := Order{}
	err = ModuleCdc.UnmarshalBinaryBare(bz, &order2)
	require.NoError(t, err)

	// Some sanity checks to ensure we're not just comparing default values below.
	require.True(t, order2.Source.Amount.Int64() > 0)
	require.True(t, order2.Destination.Amount.Int64() > 0)
	require.True(t, order2.SourceRemaining.Int64() > 0)
	require.True(t, order2.DestinationFilled.Int64() > 0)
	require.True(t, order2.SourceFilled.Int64() > 0)
	require.True(t, order2.Created.Equal(tm))

	require.Equal(t, uint64(3123), order2.ID)
	require.Equal(t, order1.ID, order2.ID)
	require.Equal(t, order1.Source, order2.Source)
	require.Equal(t, order1.Destination, order2.Destination)
	require.Equal(t, sdk.NewInt(50), order2.SourceRemaining)
	require.Equal(t, order1.SourceRemaining, order2.SourceRemaining)
	require.Equal(t, order1.SourceFilled, order2.SourceFilled)
	require.Equal(t, sdk.NewInt(50), order2.DestinationFilled)
	require.Equal(t, order1.DestinationFilled, order2.DestinationFilled)

	require.Equal(t, order1.TimeInForce, order2.TimeInForce)
	require.True(t, order1.Created.Equal(order2.Created))
	require.Equal(t, order1.OrigOrderCreated, order2.OrigOrderCreated)
}

func TestInvalidOrder(t *testing.T) {
	// 0 amount source
	_, err := NewOrder(
		time.Now(), TimeInForce_GoodTillCancel, coin("0eur"), coin("120usd"),
		[]byte("acc"), "A", time.Time{},
	)
	require.Error(t, err)

	// 0 amount destination
	_, err = NewOrder(
		time.Now(), TimeInForce_GoodTillCancel, coin("120eur"), coin("0usd"),
		[]byte("acc"), "A", time.Time{},
	)
	require.Error(t, err)

	// Same denomination
	_, err = NewOrder(
		time.Now(), TimeInForce_GoodTillCancel, coin("1000eur"), coin("850eur"),
		[]byte("acc"), "A", time.Time{},
	)
	require.Error(t, err)

	c := sdk.Coin{
		Denom:  "eur",
		Amount: sdk.NewInt(-100),
	}

	// Negative source
	_, err = NewOrder(
		time.Now(), TimeInForce_GoodTillCancel, c, coin("120usd"),
		[]byte("acc"), "B", time.Time{},
	)
	require.Error(t, err)
}

func TestMarketDataSerialization1(t *testing.T) {
	md := MarketData{
		Source:      "EUR",
		Destination: "CHF",
		LastPrice:   nil,
		Timestamp:   nil,
	}

	bz, err := ModuleCdc.MarshalBinaryBare(&md)
	require.NoError(t, err)

	md2 := MarketData{}

	err = ModuleCdc.UnmarshalBinaryBare(bz, &md2)
	require.NoError(t, err)

	// Amino serialization converts nil time.Time pointers to Unix epoch.
	expected := time.Unix(0, 0).UTC()
	require.Equal(t, &expected, md2.Timestamp)
}

func TestMarketDataSerialization2(t *testing.T) {
	ts := time.Now()
	md := MarketData{
		Source:      "EUR",
		Destination: "CHF",
		LastPrice:   nil,
		Timestamp:   &ts,
	}

	bz, err := ModuleCdc.MarshalBinaryBare(&md)
	require.NoError(t, err)

	md2 := MarketData{}

	err = ModuleCdc.UnmarshalBinaryBare(bz, &md2)
	require.NoError(t, err)

	ts = ts.UTC()
	require.Equal(t, &ts, md2.Timestamp)
}

func TestTimeInForce(t *testing.T) {
	tif, err := TimeInForceFromString("FOK")
	require.NoError(t, err)
	require.Equal(t, TimeInForce_FillOrKill, tif)

	_, err = TimeInForceFromString("f0k")
	require.Error(t, err)
}

func coin(s string) sdk.Coin {
	coin, err := sdk.ParseCoinNormalized(s)
	if err != nil {
		panic(err)
	}
	return coin
}
