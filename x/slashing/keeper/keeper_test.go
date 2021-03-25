// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	sdkslashing "github.com/cosmos/cosmos-sdk/x/slashing"
	sdkslashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/e-money/em-ledger/x/slashing/types"
	"github.com/stretchr/testify/require"
)

const (
	blockWindow = int64(30) // Number of blocks in uptime window. This is contingent upon blocktime being about 1 minute.
)

// Have to change these parameters for tests
// lest the tests take forever
func keeperTestParams() sdkslashingtypes.Params {
	params := types.DefaultParams()
	params.DowntimeJailDuration = 5 * time.Minute
	return params
}

// ______________________________________________________________

// Test a validator through uptime, downtime, revocation,
// unrevocation, starting height reset, and revocation again
func TestHandleAbsentValidator(t *testing.T) {
	// initial setup
	ctx, keeper, _, bk, sk, database := createTestComponents(t)
	keeper.SetParams(ctx, keeperTestParams())

	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, val := addrs[0], pks[0]
	sh := staking.NewHandler(sk)
	slh := sdkslashing.NewHandler(keeper.Keeper)
	_, err := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.NoError(t, err)
	//require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	require.Equal(
		t, bk.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, initTokens.Sub(amt))),
	)
	require.Equal(t, amt, sk.Validator(ctx, addr).GetBondedTokens())

	// will exist since the validator has been bonded
	info, found := keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	//require.Equal(t, int64(0), info.StartHeight)
	//require.Equal(t, int64(0), info.IndexOffset)
	//require.Equal(t, int64(0), info.MissedBlocksCounter)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	height := int64(0)
	slashable := false
	nextBlocktime := blockTimeGenerator(time.Minute)

	// 1000 first blocks OK
	for ; height < 1000; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		batch := database.NewBatch()
		slashable = height > blockWindow
		keeper.HandleValidatorSignature(ctx, batch, val.Address(), power, true, blockWindow, slashable)
		batch.Write()
	}
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	//require.Equal(t, int64(0), info.StartHeight)
	//require.Equal(t, int64(0), info.MissedBlocksCounter)

	//for ; height < keeper.SignedBlocksWindow(ctx)+(keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx)); height++ {
	nextHeight := height + blockWindow - 3 // Approach the limit of missed signed blocks
	for ; height < nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		batch := database.NewBatch()
		keeper.HandleValidatorSignature(ctx, batch, val.Address(), power, false, blockWindow, slashable)
		batch.Write()
	}
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	//require.Equal(t, int64(0), info.StartHeight)
	//require.Equal(t, keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx), info.MissedBlocksCounter)

	// validator should be bonded still
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, stakingtypes.Bonded, validator.GetStatus())
	bondPool := sk.GetBondedPool(ctx)
	require.True(sdk.IntEq(t, amt, bk.GetAllBalances(ctx, bondPool.GetAddress()).AmountOf(sk.BondDenom(ctx))))

	// 501st block missed
	ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
	batch := database.NewBatch()
	keeper.HandleValidatorSignature(ctx, batch, val.Address(), power, false, blockWindow, slashable)
	batch.Write()
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	//require.Equal(t, int64(0), info.StartHeight)
	// counter now reset to zero
	//require.Equal(t, int64(0), info.MissedBlocksCounter)

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should have been jailed
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, stakingtypes.Unbonding, validator.GetStatus())

	slashAmt := amt.ToDec().Mul(keeper.SlashFractionDowntime(ctx)).RoundInt64()

	// validator should have been slashed
	require.Equal(t, amt.Int64()-slashAmt, validator.GetTokens().Int64())

	// 502nd block *also* missed (since the LastCommit would have still included the just-unbonded validator)
	ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
	batch = database.NewBatch()
	keeper.HandleValidatorSignature(ctx, batch, val.Address(), power, false, blockWindow, slashable)
	batch.Write()
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	//require.Equal(t, int64(0), info.StartHeight)
	//require.Equal(t, int64(1), info.MissedBlocksCounter)

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should not have been slashed any more, since it was already jailed
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, amt.Int64()-slashAmt, validator.GetTokens().Int64())
	require.True(t, validator.Jailed)

	// unrevocation should fail prior to jail expiration
	_, err = slh(ctx, sdkslashingtypes.NewMsgUnjail(addr))
	require.Error(t, err)
	//require.False(t, got.IsOK())

	// unrevocation should succeed after jail expiration
	height++
	ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(5))

	_, err = slh(ctx, sdkslashingtypes.NewMsgUnjail(addr))
	require.NoError(t, err)
	//require.True(t, got.IsOK())

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should be rebonded now
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, stakingtypes.Bonded, validator.GetStatus())

	// validator should have been slashed
	bondPool = sk.GetBondedPool(ctx)
	require.Equal(t, amt.Int64()-slashAmt, bk.GetAllBalances(ctx, bondPool.GetAddress()).AmountOf(sk.BondDenom(ctx)).Int64())

	// Validator start height should not have been changed
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	//require.Equal(t, int64(0), info.StartHeight)
	// we've missed 2 blocks more than the maximum, so the counter was reset to 0 at 1 block more and is now 1
	//require.Equal(t, int64(1), info.MissedBlocksCounter)

	// validator should not be immediately jailed again
	height++
	ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
	batch = database.NewBatch()
	keeper.HandleValidatorSignature(ctx, batch, val.Address(), power, false, blockWindow, slashable)
	batch.Write()
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, stakingtypes.Bonded, validator.GetStatus())

	// 500 signed blocks
	nextHeight = height + 501
	for ; height < nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		batch = database.NewBatch()
		keeper.HandleValidatorSignature(ctx, batch, val.Address(), power, false, blockWindow, slashable)
		batch.Write()
	}

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should be jailed again after 500 unsigned blocks
	nextHeight = height + blockWindow + 1
	for ; height <= nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		batch = database.NewBatch()
		keeper.HandleValidatorSignature(ctx, batch, val.Address(), power, false, blockWindow, slashable)
		batch.Write()
	}

	// end block
	staking.EndBlocker(ctx, sk)

	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, stakingtypes.Unbonding, validator.GetStatus())
	require.True(t, validator.IsJailed())
}

// Test a new validator entering the validator set
// Ensure that SigningInfo.StartHeight is set correctly
// and that they are not immediately jailed
func TestHandleNewValidator(t *testing.T) {
	nextBlocktime := blockTimeGenerator(time.Minute)
	ctx, keeper, _, bk, sk, database := createTestComponents(t)
	keeper.SetParams(ctx, keeperTestParams())

	addr, val := addrs[0], pks[0]
	amt := sdk.TokensFromConsensusPower(100)
	sh := staking.NewHandler(sk)

	// 1000 first blocks not a validator
	ctx = ctx.WithBlockHeight(1001).WithBlockTime(nextBlocktime(1001))

	// Validator created
	_, err := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.NoError(t, err)
	//require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	require.Equal(
		t, bk.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, initTokens.Sub(amt))),
	)
	require.Equal(t, amt, sk.Validator(ctx, addr).GetBondedTokens())

	// Now a validator, for two blocks
	batch := database.NewBatch()
	keeper.HandleValidatorSignature(ctx, batch, val.Address(), 100, true, blockWindow, true)
	batch.Write()
	ctx = ctx.WithBlockHeight(blockWindow + 2).WithBlockTime(nextBlocktime(2))
	batch = database.NewBatch()
	keeper.HandleValidatorSignature(ctx, batch, val.Address(), 100, false, blockWindow, true)
	batch.Write()

	info, found := keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	//require.Equal(t, blockWindow+1, info.StartHeight)
	//require.Equal(t, int64(2), info.IndexOffset)
	//require.Equal(t, int64(1), info.MissedBlocksCounter)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)

	// validator should be bonded still, should not have been jailed or slashed
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, stakingtypes.Bonded, validator.GetStatus())
	bondPool := sk.GetBondedPool(ctx)
	expTokens := sdk.TokensFromConsensusPower(100)
	require.Equal(t, expTokens.Int64(), bk.GetAllBalances(ctx, bondPool.GetAddress()).AmountOf(sk.BondDenom(ctx)).Int64())
}

// Test a jailed validator being "down" twice
// Ensure that they're only slashed once
func TestHandleAlreadyJailed(t *testing.T) {
	nextBlocktime := blockTimeGenerator(time.Minute)
	ctx, keeper, _, bk, sk, database := createTestComponents(t)
	keeper.SetParams(ctx, keeperTestParams())

	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, val := addrs[0], pks[0]
	sh := staking.NewHandler(sk)
	_, err := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	//require.True(t, got.IsOK())
	require.NoError(t, err)
	staking.EndBlocker(ctx, sk)

	preSlashingSupply := bk.GetSupply(ctx)

	// 1000 first blocks OK
	height := int64(0)
	slashable := false
	for ; height < 1000; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		batch := database.NewBatch()
		slashable = height > blockWindow
		keeper.HandleValidatorSignature(ctx, batch, val.Address(), power, true, blockWindow, slashable)
		batch.Write()
	}

	// 501 blocks missed
	for ; height < 1501; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		batch := database.NewBatch()
		keeper.HandleValidatorSignature(ctx, batch, val.Address(), power, false, blockWindow, slashable)
		batch.Write()
	}

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should have been jailed and slashed
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, stakingtypes.Unbonding, validator.GetStatus())

	// validator should have been slashed 1/10 percent
	resultingTokens := amt.Sub(amt.QuoRaw(1000))
	require.Equal(t, resultingTokens, validator.GetTokens())

	// another block missed
	ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
	batch := database.NewBatch()
	keeper.HandleValidatorSignature(ctx, batch, val.Address(), power, false, blockWindow, slashable)
	batch.Write()

	// validator should not have been slashed twice
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, resultingTokens, validator.GetTokens())
	require.Equal(t, preSlashingSupply, bk.GetSupply(ctx))
}

// Test a validator dipping in and out of the validator set
// Ensure that missed blocks are tracked correctly and that
// the start height of the signing info is reset correctly
func TestValidatorDippingInAndOut(t *testing.T) {
	nextBlocktime := blockTimeGenerator(time.Minute)

	// initial setup
	// keeperTestParams set the SignedBlocksWindow to 1000 and MaxMissedBlocksPerWindow to 500
	ctx, keeper, _, _, sk, database := createTestComponents(t)
	keeper.SetParams(ctx, keeperTestParams())

	params := sk.GetParams(ctx)
	params.MaxValidators = 1
	sk.SetParams(ctx, params)
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, val := addrs[0], pks[0]
	consAddr := sdk.ConsAddress(addr)
	sh := staking.NewHandler(sk)
	_, err := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.NoError(t, err)
	//require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	// 100 first blocks OK
	height := int64(1)
	slashable := false
	for ; height < int64(100); height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		batch := database.NewBatch()
		slashable = height > blockWindow
		keeper.HandleValidatorSignature(ctx, batch, val.Address(), power, true, height, slashable)
		batch.Write()
	}

	// kick first validator out of validator set
	newAmt := sdk.TokensFromConsensusPower(101)
	_, err = sh(ctx, NewTestMsgCreateValidator(addrs[1], pks[1], newAmt))
	//require.True(t, got.IsOK())
	require.NoError(t, err)
	validatorUpdates := staking.EndBlocker(ctx, sk)
	require.Equal(t, 2, len(validatorUpdates))
	validator, _ := sk.GetValidator(ctx, addr)
	require.Equal(t, stakingtypes.Unbonding, validator.Status)

	// 600 more blocks happened
	height = int64(700)
	nextBlocktime(600)
	ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(0))

	// validator added back in
	delTokens := sdk.TokensFromConsensusPower(50)
	_, err = sh(ctx, stakingtypes.NewMsgDelegate(sdk.AccAddress(addrs[2]), addrs[0], sdk.NewCoin(sk.GetParams(ctx).BondDenom, delTokens)))
	require.NoError(t, err)
	//require.True(t, got.IsOK())
	validatorUpdates = staking.EndBlocker(ctx, sk)
	require.Equal(t, 2, len(validatorUpdates))
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, stakingtypes.Bonded, validator.Status)
	newPower := int64(150)

	// validator misses a block
	batch := database.NewBatch()
	keeper.HandleValidatorSignature(ctx, batch, val.Address(), newPower, false, blockWindow, slashable)
	batch.Write()
	height++

	// shouldn't be jailed/kicked yet
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, stakingtypes.Bonded, validator.Status)

	// validator misses 500 more blocks, 501 total
	latest := height
	for ; height < latest+500; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		batch = database.NewBatch()
		keeper.HandleValidatorSignature(ctx, batch, val.Address(), newPower, false, blockWindow, slashable)
		batch.Write()
	}

	// should now be jailed & kicked
	staking.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, stakingtypes.Unbonding, validator.Status)

	// check all the signing information
	signInfo, found := keeper.GetValidatorSigningInfo(ctx, consAddr)
	require.True(t, found)
	require.Equal(t, consAddr.String(), signInfo.Address)
	//require.Equal(t, int64(0), signInfo.MissedBlocksCounter)
	//require.Equal(t, int64(0), signInfo.IndexOffset)
	// array should be cleared
	// TODO Check that the validators missed blocks registration have been purged.
	//for offset := int64(0); offset < keeper.SignedBlocksWindow(ctx); offset++ {
	//	missed := keeper.getValidatorMissedBlockBitArray(ctx, consAddr, offset)
	//	require.False(t, missed)
	//}

	// some blocks pass
	height = int64(5000)
	nextBlocktime(4000)
	ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(0))

	// validator rejoins and starts signing again
	sk.Unjail(ctx, consAddr)

	batch = database.NewBatch()
	keeper.HandleValidatorSignature(ctx, batch, val.Address(), newPower, true, blockWindow, slashable)
	batch.Write()
	height++

	// validator should not be kicked since we reset counter/array when it was jailed
	staking.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, stakingtypes.Bonded, validator.Status)

	// validator misses 501 blocks
	latest = height
	for ; height < latest+501; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		batch = database.NewBatch()
		keeper.HandleValidatorSignature(ctx, batch, val.Address(), newPower, false, blockWindow, slashable)
		batch.Write()
	}

	// validator should now be jailed & kicked
	staking.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, stakingtypes.Unbonding, validator.Status)
}

func TestHandlePendingPenalties(t *testing.T) {
	specs := map[string]struct {
		srcPenalties    types.Penalties
		srcValidatorSet map[string]bool
		exp             types.Penalties
	}{
		"payout when validator not in the active set": {
			srcPenalties: types.Penalties{Elements: []types.Penalty{{
				Validator: "a",
				Amounts:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1)))},
			}},
			srcValidatorSet: map[string]bool{},
		},
		"validator in the active set": {
			srcPenalties: types.Penalties{Elements: []types.Penalty{{
				Validator: "a",
				Amounts:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1)))},
			}},
			srcValidatorSet: map[string]bool{
				"a": true,
			},
			exp: types.Penalties{Elements: []types.Penalty{{
				Validator: "a",
				Amounts:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1)))},
			}},
		},
		"non penalties": {
			srcValidatorSet: map[string]bool{
				"a": true,
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, keeper, accountKeeper, bankKeeper, _, database := createTestComponents(t)

			err := bankKeeper.SetBalances(ctx, accountKeeper.GetModuleAddress(types.PenaltyAccount), sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1))))
			require.NoError(t, err)

			batch := database.NewBatch()
			keeper.setPendingPenalties(batch, spec.srcPenalties)
			require.NoError(t, batch.Write())

			// when
			batch = database.NewBatch()
			fn := func() map[string]bool { return spec.srcValidatorSet }
			keeper.handlePendingPenalties(ctx, batch, fn)
			require.NoError(t, batch.Write())

			// then
			assert.Equal(t, spec.exp, keeper.getPendingPenalties())
		})
	}

}

func blockTimeGenerator(blocktime time.Duration) func(int) time.Time {
	// TODO This might be a tiny bit over-engineered. Move to a simple struct?
	now := time.Now()

	return func(blockcount int) time.Time {
		now = now.Add(time.Duration(blockcount) * blocktime)
		return now
	}
}
