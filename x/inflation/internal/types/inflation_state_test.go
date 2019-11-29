// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"github.com/stretchr/testify/assert"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNewParams1(t *testing.T) {
	is := NewInflationState("caps", "0.04", "kredits", "0.0")
	assert.NoError(t, ValidateInflationState(is))
	assert.Len(t, is.InflationAssets, 2)

	assert.Equal(t, sdk.NewDecWithPrec(4, 2), is.InflationAssets[0].Inflation)
	assert.Equal(t, sdk.NewDec(0), is.InflationAssets[1].Inflation)
}

func TestValidation(t *testing.T) {
	inflationStates := [...]InflationState{
		NewInflationState("caps", "-0.04"),
		NewInflationState("caps", "0.04", "CAPS", "0.10"),
	}

	for _, is := range inflationStates {
		err := ValidateInflationState(is)
		assert.Error(t, err)
	}
}

func TestFindAndChangeAssetByDenom(t *testing.T) {
	is := NewInflationState("caps", "0.04", "kredits", "0.0")

	kroner := is.FindByDenom("kroner")
	assert.Nil(t, kroner)

	caps := is.FindByDenom("caps")
	assert.NotNil(t, caps)
	caps.Inflation, _ = sdk.NewDecFromStr("0.25")

	assert.Equal(t, sdk.NewDecWithPrec(25, 2), is.FindByDenom("caps").Inflation)

}
