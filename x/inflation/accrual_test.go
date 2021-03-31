// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package inflation

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestYearHourlyAccrual(t *testing.T) {
	accum := sdk.NewDec(0)
	minted := sdk.ZeroInt()
	supply := sdk.NewInt(2300000000)
	annualInflation := sdk.NewDecFromIntWithPrec(sdk.NewInt(1), 2)

	lastAccrual := time.Now()

	totalMinted := sdk.ZeroInt()
	for i := 0; i < 365*24; i++ {
		accum, minted = calculateInflation(accum, supply, annualInflation, lastAccrual, lastAccrual.Add(time.Hour))
		lastAccrual = lastAccrual.Add(time.Hour)
		totalMinted = totalMinted.Add(minted)
	}

	assert.Equal(t, sdk.NewInt(23000000), totalMinted, "minted %v", totalMinted)
	assert.True(t, sdk.NewDec(0).Equal(accum), "accum", accum.String())

}

func TestRandomBlockTimes(t *testing.T) {
	accum := sdk.NewDec(0)
	minted := sdk.ZeroInt()
	supply := sdk.NewInt(2300000000)
	annualInterest := sdk.NewDecFromIntWithPrec(sdk.NewInt(1), 2)

	startTime := time.Now()
	lastAccrual := startTime

	endTime := startTime.Add(365 * 24 * time.Hour)

	totalMinted := sdk.ZeroInt()
	totalDuration := time.Duration(0)
	blockCount := 0

	rand.Seed(1) // Reset rand

	for {
		blockCount++
		d := time.Duration(rand.Int63n(120)) * time.Second
		if lastAccrual.Add(d).After(endTime) {
			d = endTime.Sub(lastAccrual)
		}

		totalDuration = totalDuration + d

		accum, minted = calculateInflation(accum, supply, annualInterest, lastAccrual, lastAccrual.Add(d))
		lastAccrual = lastAccrual.Add(d)
		totalMinted = totalMinted.Add(minted)

		if lastAccrual == endTime {
			break
		}
	}

	// Sanity check test
	assert.Equal(t, 365*24*time.Hour, totalDuration)

	assert.Equal(t, sdk.NewInt(23000000), totalMinted, "minted %v", totalMinted)
	assert.True(t, sdk.NewDec(0).Equal(accum), "accum", accum.String())
}

// Simulate an entire years worth of compounded inflation when calculated each minute
func TestMultipleCoinsAccrual(t *testing.T) {
	currentTime := time.Now().UTC()
	state := NewInflationState(currentTime, "credit", "0.001", "buck", "0.03")

	supply := sdk.NewCoins(
		sdk.NewCoin("buck", sdk.NewInt(1000000000)),
		sdk.NewCoin("credit", sdk.NewInt(1000)),
	)

	state.LastAppliedTime = currentTime

	for i := 0; i < 365*24*60; i++ {
		currentTime = currentTime.Add(time.Minute)
		mintedCoins := applyInflation(&state, supply, currentTime)

		// Add the minted coins to the total supply
		supply = supply.Add(mintedCoins...)
	}

	assert.Equal(t, sdk.NewInt(1001), supply.AmountOf("credit"))
	assert.Equal(t, sdk.NewInt(1030454533), supply.AmountOf("buck"))
}
