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
	LastAccrual     time.Time       `json:"last_accrual" yaml:"last_accrual"`
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
		LastAccrual:     time.Now().UTC(),
	}
}

func DefaultInflationState() InflationState {
	return NewInflationState("caps", "0.01", "kredits", "0.05")
}

// validate params
func ValidateInflationState(is InflationState) error {
	// TODO No duplicate denoms

	//if params.MintDenom == "" {
	//	return fmt.Errorf("mint parameter MintDenom can't be an empty string")
	//}

	return nil
}

func (is InflationState) String() string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Last Accrual time: %v\n", is.LastAccrual))
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
