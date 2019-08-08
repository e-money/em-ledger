// +build ignore

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"
)

type AssetState struct {
	LastAccrual time.Time
	Accum       sdk.Dec
}

// Minter represents the minting state.
type Minter struct {
	AssetsInflationState map[string]AssetState

	//Inflation        sdk.Dec `json:"inflation" yaml:"inflation"`                 // current annual inflation rate
	//AnnualProvisions sdk.Dec `json:"annual_provisions" yaml:"annual_provisions"` // current annual expected provisions
	//LastAccrual      time.Time `json:"last_accrual_time" yaml:"last_accrual_time"`
}

// NewMinter returns a new Minter object with the given inflation and annual
// provisions values.
func NewMinter(inflation, annualProvisions sdk.Dec) Minter {
	return Minter{}
}

// InitialMinter returns an initial Minter object with a given inflation value.
func InitialMinter(inflation sdk.Dec) Minter {
	return NewMinter(
		inflation,
		sdk.NewDec(0),
	)
}

// DefaultInitialMinter returns a default initial Minter object for a new chain
// which uses an inflation rate of 13%.
func DefaultInitialMinter() Minter {
	return InitialMinter(
		sdk.NewDecWithPrec(13, 2),
	)
}

// validate minter
func ValidateMinter(minter Minter) error {
	//if minter.Inflation.LT(sdk.ZeroDec()) {
	//	return fmt.Errorf("mint parameter Inflation should be positive, is %s",
	//		minter.Inflation.String())
	//}
	return nil
}
