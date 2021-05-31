package ante

import (
	"fmt"
	sdkante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// This is a fork of the DeductFeeDecorator from the SDK version v0.42.4
// It deposits the fees differently based on what denomination it is paid with.
// If the fee is paid in NGM, it is sent to the tradition fee pool and distributed as rewards
// If the fee is paid with a stablecoin balance, it is sent to the buyback module
// https://github.com/e-money/em-ledger/issues/41
type DeductFeeDecorator struct {
	ak            sdkante.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper StakingKeeper
}

func NewDeductFeeDecorator(ak sdkante.AccountKeeper, bk types.BankKeeper, sk StakingKeeper) DeductFeeDecorator {
	return DeductFeeDecorator{
		ak:            ak,
		bankKeeper:    bk,
		stakingKeeper: sk,
	}
}

func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if addr := dfd.ak.GetModuleAddress(types.FeeCollectorName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.FeeCollectorName))
	}

	feePayer := feeTx.FeePayer()
	feePayerAcc := dfd.ak.GetAccount(ctx, feePayer)

	if feePayerAcc == nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", feePayer)
	}

	// deduct the fees
	if !feeTx.GetFee().IsZero() {
		err = DeductFees(dfd.bankKeeper, ctx, feePayerAcc, feeTx.GetFee())
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// DeductFees deducts fees from the given account.
func DeductFees(bankKeeper types.BankKeeper, ctx sdk.Context, acc types.AccountI, fees sdk.Coins) error {
	if !fees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fees)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}
