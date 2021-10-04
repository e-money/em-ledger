package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
// Forked from sdk v0.42.4
func NewAnteHandler(
	ak sdkante.AccountKeeper, bankKeeper types.BankKeeper, stakingKeeper StakingKeeper,
	signModeHandler signing.SignModeHandler,
) sdk.AnteHandler {
	sigGasConsumer := sdkante.DefaultSigVerificationGasConsumer

	return sdk.ChainAnteDecorators(
		sdkante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		sdkante.NewRejectExtensionOptionsDecorator(),
		sdkante.NewMempoolFeeDecorator(),
		sdkante.NewValidateBasicDecorator(),
		sdkante.TxTimeoutHeightDecorator{},
		sdkante.NewValidateMemoDecorator(ak),
		sdkante.NewConsumeGasForTxSizeDecorator(ak),
		sdkante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		sdkante.NewValidateSigCountDecorator(ak),
		// TODO review and reconcile new sdk DeductFeeDecorator next 2 lines
		// sdkante.NewDeductFeeDecorator(ak, bankKeeper, nil),
		NewDeductFeeDecorator(ak, bankKeeper, stakingKeeper),
		sdkante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		sdkante.NewSigVerificationDecorator(ak, signModeHandler),
		sdkante.NewIncrementSequenceDecorator(ak),
	)
}
