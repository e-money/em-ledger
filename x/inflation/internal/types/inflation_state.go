// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/x/params/subspace"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter store keys
var (
	KeyLastAppliedTime   = []byte("LastAppliedTime")
	KeyLastAppliedHeight = []byte("LastAppliedHeight")
	KeyInflationAssets   = []byte("InflationAssets")
)

type InflationAsset struct {
	Denom     string  `json:"denom" yaml:"denom"`
	Inflation sdk.Dec `json:"inflation" yaml:"inflation"`
	Accum     sdk.Dec `json:"accum" yaml:"accum"`
}

type InflationAssets = []InflationAsset

type InflationState struct {
	LastAppliedTime   time.Time       `json:"last_applied" yaml:"last_applied"`
	LastAppliedHeight sdk.Int         `json:"last_applied_height" yaml:"last_applied_height"`
	InflationAssets   InflationAssets `json:"assets" yaml:"assets"`
}

func (is InflationState) ParamSetPairs() subspace.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyLastAppliedTime, &is.LastAppliedTime, nil),
		params.NewParamSetPair(KeyLastAppliedHeight, &is.LastAppliedHeight, nil),
		params.NewParamSetPair(KeyInflationAssets, &is.InflationAssets, nil),
	}
}

func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&InflationState{})
}

func NewInflationState(assets ...string) InflationState {
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
		LastAppliedTime:   time.Now().UTC(),
		LastAppliedHeight: sdk.ZeroInt(),
	}
}

func DefaultInflationState() InflationState {
	return NewInflationState()
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
