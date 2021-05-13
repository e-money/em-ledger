package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/pkg/errors"
)

const (
	defaultTrxFee = 25000
	defaultLiquidMinutesSpan = 5
)

var (
	// DefaultLiquidTrxFee zero by default
	DefaultLiquidTrxFee = sdk.ZeroInt()

	// KeyTrxFee is store's key for TrxFee Param
	KeyTrxFee = []byte("TrxFee")
	// KeyLiquidTrxFee is store's key for the LiquidTrxFee
	KeyLiquidTrxFee = []byte("LiquidTrxFee")
	// KeyLiquidityRebateMinutesSpan is store's key for the
	// LiquidityRebateMinutesSpan
	KeyLiquidityRebateMinutesSpan = []byte("LiquidityRebateMinutesSpan")
)

// ParamKeyTable for bank module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&TxParams{})
}

// ParamSetPairs implements params.ParamSet
func (p *TxParams) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyTrxFee, &p.TrxFee, validateIsUInt),
		paramtypes.NewParamSetPair(KeyLiquidTrxFee, &p.LiquidTrxFee, validateIsUInt),
		paramtypes.NewParamSetPair(
			KeyLiquidityRebateMinutesSpan, &p.LiquidityRebateMinutesSpan,
			validateTimeSpan,
		),
	}
}

// NewTxParams creates a new parameter configuration for the bank module
func NewTxParams(trxFee, liquidTrxFee uint64, liquidityRebateMinutes int64) TxParams {
	return TxParams{
		TrxFee:                     trxFee,
		LiquidTrxFee:               liquidTrxFee,
		LiquidityRebateMinutesSpan: liquidityRebateMinutes,
	}
}

// DefaultTxParams are the default Trx market parameters.
func DefaultTxParams() TxParams {
	return TxParams{
		TrxFee:                     defaultTrxFee,
		LiquidTrxFee:               0,
		LiquidityRebateMinutesSpan: defaultLiquidMinutesSpan,
	}
}

// Validate all Tx Market Params parameters
func (p TxParams) Validate() error {
	if err := validateIsNonZeroUInt(p.TrxFee); err != nil {
		return err
	}

	if err := validateIsUInt(p.LiquidTrxFee); err != nil {
		return err
	}

	return validateTimeSpan(p.LiquidityRebateMinutesSpan)
}

func validateTimeSpan(i interface{}) error {
	m, ok := i.(int64)

	if !ok {
		return fmt.Errorf("invalid minutes parameter type: %T", i)
	}

	if m < 0 {
		return fmt.Errorf("minutes parameter cannot be < 0: %d", m)
	}

	return nil
}

func validateIsUInt(u interface{}) error {
	_, ok := u.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", u)
	}

	return nil
}

func validateIsNonZeroUInt(u interface{}) error {
	v, ok := u.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", u)
	}
	if v == 0 {
		return errors.New("cannot be 0")
	}

	return nil
}
