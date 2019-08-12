package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter store keys
var (
	KeyMintDenom = []byte("MintDenom")
	KeyParams    = []byte("MintParameters")
)

// TODO Divide into two? One "base class" holding Denom and inflation for use in Genesis and a "subclass" with current state
type InflationAsset struct {
	Denom     string  `json:"denom" yaml:"denom"`
	Inflation sdk.Dec `json:"inflation" yaml:"inflation"`
	Accum     sdk.Dec `json:"accum" yaml:"accum"`
}

type InflationAssets = []InflationAsset

type InflationState struct {
	LastApplied     time.Time       `json:"last_applied" yaml:"last_applied"`
	InflationAssets InflationAssets `json:"assets" yaml:"assets"`
}

// ParamTable for minting module.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterType(KeyParams, InflationState{})
	//.RegisterParamSet(&Params{})
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
		InflationAssets: result,
		LastApplied:     time.Now().UTC(),
	}
}

func DefaultInflationState() InflationState {
	return NewInflationState("caps", "0.01", "kredits", "0.05")
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

	result.WriteString(fmt.Sprintf("Last inflation: %v\n", is.LastApplied))
	result.WriteString("Inflation state:\n")
	for _, asset := range is.InflationAssets {
		result.WriteString(fmt.Sprintf("\tDenom: %v\t\t\tInflation: %v\t\tAccum: %v\n", asset.Denom, asset.Inflation, asset.Accum))
	}
	return result.String()
}

// Implements params.ParamSet
func (is *InflationState) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		//{KeyMintDenom, &p.MintDenom},
	}
}
