package keeper

import (
	apptypes "github.com/e-money/em-ledger/types"
	"sort"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	batch := apptypes.GetCurrentBatch(ctx)
	if batch == nil {
		panic("batch object not found") // todo (reviewer): panic in begin blocker is not handled downstream and will crash the node.
	}
	signedBlocksWindow := k.SignedBlocksWindowDuration(ctx)

	blockTimes := k.getBlockTimes()
	blockTimes = append(blockTimes, ctx.BlockTime())
	slashable := false
	slashable, blockTimes = truncateByWindow(ctx.BlockTime(), blockTimes, signedBlocksWindow)
	k.setBlockTimes(batch, blockTimes)

	// Iterate over all the validators which *should* have signed this block
	// store whether or not they have actually signed it and slash/unbond any
	// which have missed too many blocks in a row (downtime slashing)
	for _, voteInfo := range req.LastCommitInfo.GetVotes() {
		k.HandleValidatorSignature(ctx, batch, voteInfo.Validator.Address, voteInfo.Validator.Power, voteInfo.SignedLastBlock, int64(len(blockTimes)), slashable)
	}
}

func truncateByWindow(blockTime time.Time, times []time.Time, signedBlocksWindow time.Duration) (bool, []time.Time) {

	if len(times) > 0 && times[0].Add(signedBlocksWindow).Before(blockTime) {
		// Remove timestamps outside of the time window we are watching
		threshold := blockTime.Add(-signedBlocksWindow)

		index := sort.Search(len(times), func(i int) bool {
			return times[i].After(threshold)
		})

		return true, times[index:]
	}

	return false, times
}
