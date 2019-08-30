package slashing

import (
	"fmt"
	"sort"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	blockTimes []time.Time
)

// slashing begin block functionality
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, sk Keeper) {
	signedBlocksWindow := sk.SignedBlocksWindowDuration(ctx)

	blockTimes = append(blockTimes, ctx.BlockTime())
	blockTimes = truncateByWindow(blockTimes, signedBlocksWindow)

	// Iterate over all the validators which *should* have signed this block
	// store whether or not they have actually signed it and slash/unbond any
	// which have missed too many blocks in a row (downtime slashing)
	for _, voteInfo := range req.LastCommitInfo.GetVotes() {
		sk.HandleValidatorSignature(ctx, voteInfo.Validator.Address, voteInfo.Validator.Power, voteInfo.SignedLastBlock, int64(len(blockTimes)))
	}

	// Iterate through any newly discovered evidence of infraction
	// Slash any validators (and since-unbonded stake within the unbonding period)
	// who contributed to valid infractions
	for _, evidence := range req.ByzantineValidators {
		switch evidence.Type {
		case tmtypes.ABCIEvidenceTypeDuplicateVote:
			sk.HandleDoubleSign(ctx, evidence.Validator.Address, evidence.Height, evidence.Time, evidence.Validator.Power)
		default:
			sk.Logger(ctx).Error(fmt.Sprintf("ignored unknown evidence type: %s", evidence.Type))
		}
	}
}

func truncateByWindow(times []time.Time, signedBlocksWindow time.Duration) []time.Time {
	if len(times) == 0 {
		return times
	}

	// Remove timestamps outside of the time window we are watching
	threshold := times[len(times)-1].Add(-1 * signedBlocksWindow)

	index := sort.Search(len(times), func(i int) bool {
		return times[i].After(threshold)
	})

	return times[index:]
}
