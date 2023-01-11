package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyLastAppliedTime   = []byte("LastAppliedTime")
	KeyLastAppliedHeight = []byte("LastAppliedHeight")
	KeyInflationAssets   = []byte("InflationAssets")
)

type InflationAssets = []InflationAsset

func (is InflationState) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyLastAppliedTime, &is.LastAppliedTime, nil),
		paramtypes.NewParamSetPair(KeyLastAppliedHeight, &is.LastAppliedHeight, nil),
		paramtypes.NewParamSetPair(KeyInflationAssets, &is.InflationAssets, nil),
	}
}

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&InflationState{})
}

func NewInflationState(now time.Time, assets ...string) InflationState {
	if len(assets)%2 != 0 {
		panic("Unable to parse asset parameters")
	}

	result := make(InflationAssets, 0)
	for i := 0; i < len(assets); i += 2 {
		inflation, err := sdk.NewDecFromStr(assets[i+1])
		if err != nil {
			panic(err)
		}

		result = append(result, InflationAsset{
			Denom:     assets[i],
			Inflation: inflation,
			Accum:     sdk.NewDec(0),
		})
	}

	return InflationState{
		InflationAssets:   result,
		LastAppliedTime:   now.UTC(),
		LastAppliedHeight: sdk.ZeroInt(),
	}
}

func DefaultInflationState() InflationState {
	// only called once when generating genesis
	return NewInflationState(time.Now())
}

// validate params
func ValidateInflationState(is InflationState) error {
	// Check for duplicates
	{
		duplicateDenoms := make(map[string]interface{})
		for _, asset := range is.InflationAssets {
			duplicateDenoms[strings.ToLower(asset.Denom)] = true
		}

		if len(duplicateDenoms) != len(is.InflationAssets) {
			return fmt.Errorf("inflation parameters contain duplicate denominations")
		}
	}

	// Check for negative inflation
	{
		for _, asset := range is.InflationAssets {
			if asset.Inflation.IsNegative() {
				return fmt.Errorf("inflation parameters contain an asset with negative interest: %v", asset.Denom)
			}
		}
	}

	return nil
}

func (is InflationState) String() string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Last inflation: %v\n", is.LastAppliedTime))
	result.WriteString("Inflation state:\n")
	for _, asset := range is.InflationAssets {
		result.WriteString(fmt.Sprintf("\tDenom: %v\t\t\tInflation: %v\t\tAccum: %v\n", asset.Denom, asset.Inflation, asset.Accum))
	}

	return result.String()
}

func (is *InflationState) FindByDenom(denom string) *InflationAsset {
	for i, a := range is.InflationAssets {
		if a.Denom == denom {
			return &is.InflationAssets[i]
		}
	}
	return nil
}
