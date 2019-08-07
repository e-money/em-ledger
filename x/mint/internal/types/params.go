package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter store keys
var (
	KeyMintDenom = []byte("MintDenom")
	KeyParams    = []byte("MintParameters")
)

type InflationAsset struct {
	Denom     string  `json:"denom" yaml:"denom"`
	Inflation sdk.Dec `json:"inflation" yaml:"inflation"`
}

type InflationAssets = []InflationAsset

// mint parameters
type Params struct {
	MintDenom       string          `json:"mint_denom" yaml:"mint_denom"` // type of coin to mint
	InflationAssets InflationAssets `json:"assets" yaml:"assets"`
}

// ParamTable for minting module.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{}).RegisterType(KeyParams, Params{})
}

func NewParams(mintDenom string, assets ...string) Params {
	if len(assets)%2 != 0 {
		panic("Unable to parse asset parameters")
	}

	result := make(InflationAssets, 0)
	for i := 0; i < len(assets); i += 2 {
		inflation, err := sdk.NewDecFromStr(assets[i+1])
		if err != nil {
			panic(err)
		}

		result = append(result, InflationAsset{assets[i], inflation})
	}

	return Params{
		MintDenom:       mintDenom,
		InflationAssets: result,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return NewParams("", "caps", "0.01", "kredits", "0.05")
}

// validate params
func ValidateParams(params Params) error {
	// TODO No duplicate denoms

	if params.MintDenom == "" {
		return fmt.Errorf("mint parameter MintDenom can't be an empty string")
	}
	return nil
}

func (p Params) String() string {
	var result strings.Builder

	result.WriteString("Minting params:\n")
	for _, asset := range p.InflationAssets {
		result.WriteString(fmt.Sprintf("	 Denom: %v	 	 Inflation: %v\n", asset.Denom, asset.Inflation))
	}
	return result.String()
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyMintDenom, &p.MintDenom},
	}
}
