package mint

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)
import sdk "github.com/cosmos/cosmos-sdk/types"

func TestFullYear(t *testing.T) {
	accum := sdk.NewDec(0)
	minted := sdk.ZeroInt()
	supply := sdk.NewInt(2300000000)
	annualInterest := sdk.NewDecFromIntWithPrec(sdk.NewInt(1), 2)

	lastAccrual := time.Now()

	accum, minted = calculateAccrual(accum, supply, annualInterest, lastAccrual, lastAccrual.Add(365*24*time.Hour))

	assert.Equal(t, sdk.NewInt(23000000), minted)
	assert.True(t, sdk.NewDec(0).Equal(accum))
}

func TestOneYearPerHour(t *testing.T) {
	accum := sdk.NewDec(0)
	minted := sdk.ZeroInt()
	supply := sdk.NewInt(2300000000)
	annualInterest := sdk.NewDecFromIntWithPrec(sdk.NewInt(1), 2)

	lastAccrual := time.Now()

	totalMinted := sdk.ZeroInt()
	for i := 0; i < 365*24; i++ {
		accum, minted = calculateAccrual(accum, supply, annualInterest, lastAccrual, lastAccrual.Add(time.Hour))
		lastAccrual = lastAccrual.Add(time.Hour)
		totalMinted = totalMinted.Add(minted)
	}

	fmt.Println(totalMinted)
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

		accum, minted = calculateAccrual(accum, supply, annualInterest, lastAccrual, lastAccrual.Add(d))
		lastAccrual = lastAccrual.Add(d)
		totalMinted = totalMinted.Add(minted)

		if lastAccrual == endTime {
			break
		}
	}

	fmt.Println("Total minted: ", totalMinted)
	fmt.Println("Remaining accum: ", accum)
	fmt.Println("Total duration: ", totalDuration)
	fmt.Println("Block count", blockCount)
}
