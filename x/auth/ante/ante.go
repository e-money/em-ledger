package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	channelkeeper "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/keeper"
	ibcante "github.com/cosmos/cosmos-sdk/x/ibc/core/ante"
)

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
// Forked from sdk v0.42.4
func NewAnteHandler(
	ak sdkante.AccountKeeper, bankKeeper types.BankKeeper, stakingKeeper StakingKeeper,
	signModeHandler signing.SignModeHandler,
	channelKeeper channelkeeper.Keeper,
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
		sdkante.NewRejectFeeGranterDecorator(),
		sdkante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		sdkante.NewValidateSigCountDecorator(ak),
		NewDeductFeeDecorator(ak, bankKeeper, stakingKeeper),
		sdkante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		sdkante.NewSigVerificationDecorator(ak, signModeHandler),
		sdkante.NewIncrementSequenceDecorator(ak),
		ibcante.NewAnteDecorator(channelKeeper),
	)
}
