package ante

import (
	"fmt"
	sdkante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/e-money/em-ledger/x/buyback"
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

	if addr := dfd.ak.GetModuleAddress(buyback.AccountName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", buyback.AccountName))
	}

	feePayer := feeTx.FeePayer()
	feePayerAcc := dfd.ak.GetAccount(ctx, feePayer)

	if feePayerAcc == nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", feePayer)
	}

	// deduct the fees
	if !feeTx.GetFee().IsZero() {
		err = deductFees(dfd.bankKeeper, dfd.stakingKeeper, ctx, feePayerAcc, feeTx.GetFee())
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// deductFees deducts fees from the given account.
func deductFees(bankKeeper types.BankKeeper, stakingKeeper StakingKeeper, ctx sdk.Context, acc types.AccountI, fees sdk.Coins) error {
	if !fees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	bondDenom := stakingKeeper.BondDenom(ctx)

	stakingTokenFee := sdk.NewCoins()
	stableCoinFee := sdk.NewCoins()

	// Separate fee into staking token and stablecoins.
	for _, coin := range fees {
		if coin.Denom == bondDenom {
			stakingTokenFee = stakingTokenFee.Add(coin)
		} else {
			stableCoinFee = stableCoinFee.Add(coin)
		}
	}

	if !stakingTokenFee.IsZero() {
		err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, stakingTokenFee)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
		}
	}

	if !stableCoinFee.IsZero() {
		err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), buyback.AccountName, stableCoinFee)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
		}
	}

	return nil
}
