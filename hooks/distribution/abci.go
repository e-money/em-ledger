package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/supply"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/db"
)

var (
	previousProposerKey = []byte("emdistr/previousproposer")
)

// Adapted from cosmos-sdk/x/distribution/abci.go
// A custom version was needed to keep the address of the previousProposer out of the consensus-state.

// set the proposer for determining distribution during endblock
// and distribute rewards for the previous block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k distr.Keeper, sk supply.Keeper, db db.DB) {
	// determine the total power signing the block
	var previousTotalPower, sumPreviousPrecommitPower int64
	for _, voteInfo := range req.LastCommitInfo.GetVotes() {
		previousTotalPower += voteInfo.Validator.Power
		if voteInfo.SignedLastBlock {
			sumPreviousPrecommitPower += voteInfo.Validator.Power
		}
	}

	previousProposer := db.Get(previousProposerKey)

	// TODO this is Tendermint-dependent
	// ref https://github.com/cosmos/cosmos-sdk/issues/3095
	if ctx.BlockHeight() > 1 {
		feeCollector := sk.GetModuleAccount(ctx, auth.FeeCollectorName)
		coins := feeCollector.GetCoins()

		// Only call AllocateTokens if there are in fact tokens to allocate.
		if !coins.IsZero() {
			k.AllocateTokens(ctx, sumPreviousPrecommitPower, previousTotalPower, previousProposer, req.LastCommitInfo.GetVotes())
		}
	}

	previousProposer = sdk.ConsAddress(req.Header.ProposerAddress)
	db.Set(previousProposerKey, previousProposer)
}
