package ante

import (
	"fmt"

	sdkante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/e-money/em-ledger/x/buyback"
)

// DeductFeeDecorator deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted
// Fees are deposited differently based on what denomination it is paid with.
// If the fee is paid in NGM, it is sent to the tradition fee pool and distributed as rewards
// If the fee is paid with a stablecoin balance, it is sent to the buyback module
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
// https://github.com/e-money/em-ledger/issues/41
// This is a fork of from the SDK (version v0.42.4)
type DeductFeeDecorator struct {
	ak             sdkante.AccountKeeper
	bankKeeper     types.BankKeeper
	stakingKeeper  StakingKeeper
	feegrantKeeper FeegrantKeeper
}

func NewDeductFeeDecorator(ak sdkante.AccountKeeper, bk types.BankKeeper, sk StakingKeeper, fk FeegrantKeeper) DeductFeeDecorator {
	return DeductFeeDecorator{
		ak:             ak,
		bankKeeper:     bk,
		stakingKeeper:  sk,
		feegrantKeeper: fk,
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

	fee := feeTx.GetFee()
	feeGranter := feeTx.FeeGranter()
	feePayer := feeTx.FeePayer()
	feePayerAcc := dfd.ak.GetAccount(ctx, feePayer)

	deductFeesFrom := feePayer

	if feePayerAcc == nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", feePayer)
	}

	// if feegranter set deduct fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil {
		if dfd.feegrantKeeper == nil {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, tx.GetMsgs())

			if err != nil {
				return ctx, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := dfd.ak.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", deductFeesFrom)
	}

	// deduct the fees
	if !feeTx.GetFee().IsZero() {
		//err = deductFees(dfd.bankKeeper, dfd.stakingKeeper, ctx, feePayerAcc, feeTx.GetFee())
		err = DeductFees(dfd.bankKeeper, ctx, deductFeesFromAcc, feeTx.GetFee())
		if err != nil {
			return ctx, err
		}
	}

	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
	)}
	ctx.EventManager().EmitEvents(events)

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
