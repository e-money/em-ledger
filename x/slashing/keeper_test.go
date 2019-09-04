package slashing

import (
	"emoney/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	blockWindow = int64(30) // Number of blocks in uptime window. This is contingent upon blocktime being about 1 minute.
)

// Have to change these parameters for tests
// lest the tests take forever
func keeperTestParams() types.Params {
	params := types.DefaultParams()
	params.DowntimeJailDuration = 5 * time.Minute
	return params
}

// ______________________________________________________________

// Test that a validator is slashed correctly
// when we discover evidence of infraction
func TestHandleDoubleSign(t *testing.T) {
	// initial setup
	ctx, ck, sk, _, keeper, supplyKeeper := createTestInput(t, keeperTestParams())
	// validator added pre-genesis
	ctx = ctx.WithBlockHeight(-1)
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	operatorAddr, val := addrs[0], pks[0]
	got := staking.NewHandler(sk)(ctx, NewTestMsgCreateValidator(operatorAddr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)
	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(operatorAddr)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, initTokens.Sub(amt))),
	)
	require.Equal(t, amt, sk.Validator(ctx, operatorAddr).GetBondedTokens())

	// handle a signature to set signing info
	keeper.HandleValidatorSignature(ctx, val.Address(), amt.Int64(), true, blockWindow)

	// Keep track of token supplies before the slashing
	preSlashSupply := supplyKeeper.GetSupply(ctx)
	feeAccount := supplyKeeper.GetModuleAccount(ctx, auth.FeeCollectorName)
	require.True(t, feeAccount.GetCoins().IsZero())

	oldTokens := sk.Validator(ctx, operatorAddr).GetTokens()

	// double sign less than max age
	keeper.HandleDoubleSign(ctx, val.Address(), 0, time.Unix(0, 0), power)

	// should be jailed
	require.True(t, sk.Validator(ctx, operatorAddr).IsJailed())

	// tokens should be decreased
	newTokens := sk.Validator(ctx, operatorAddr).GetTokens()
	require.True(t, newTokens.LT(oldTokens))

	// New evidence
	keeper.HandleDoubleSign(ctx, val.Address(), 0, time.Unix(0, 0), power)

	// tokens should be the same (capped slash)
	require.True(t, sk.Validator(ctx, operatorAddr).GetTokens().Equal(newTokens))

	// Jump to past the unbonding period
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1, 0).Add(sk.GetParams(ctx).UnbondingTime)})

	// No tokens should have been burned, but rather sent to the fee distribution account
	require.Equal(t, preSlashSupply, supplyKeeper.GetSupply(ctx))
	feeAccount = supplyKeeper.GetModuleAccount(ctx, auth.FeeCollectorName)
	assert.Equal(t, sdk.NewInt(5000000), feeAccount.GetCoins().AmountOf(sk.BondDenom(ctx)))

	// Still shouldn't be able to unjail
	msgUnjail := types.NewMsgUnjail(operatorAddr)
	res := handleMsgUnjail(ctx, msgUnjail, keeper)
	require.False(t, res.IsOK())

	// Should be able to unbond now
	del, _ := sk.GetDelegation(ctx, sdk.AccAddress(operatorAddr), operatorAddr)
	validator, _ := sk.GetValidator(ctx, operatorAddr)

	totalBond := validator.TokensFromShares(del.GetShares()).TruncateInt()
	msgUnbond := staking.NewMsgUndelegate(sdk.AccAddress(operatorAddr), operatorAddr, sdk.NewCoin(sk.GetParams(ctx).BondDenom, totalBond))
	res = staking.NewHandler(sk)(ctx, msgUnbond)
	require.True(t, res.IsOK())
}

// ______________________________________________________________

// Test that a validator is slashed correctly
// when we discover evidence of infraction
func TestPastMaxEvidenceAge(t *testing.T) {
	// initial setup
	ctx, ck, sk, _, keeper, _ := createTestInput(t, keeperTestParams())
	// validator added pre-genesis
	ctx = ctx.WithBlockHeight(-1)
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	operatorAddr, val := addrs[0], pks[0]
	got := staking.NewHandler(sk)(ctx, NewTestMsgCreateValidator(operatorAddr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)
	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(operatorAddr)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, initTokens.Sub(amt))),
	)
	require.Equal(t, amt, sk.Validator(ctx, operatorAddr).GetBondedTokens())

	// handle a signature to set signing info
	keeper.HandleValidatorSignature(ctx, val.Address(), power, true, blockWindow)

	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1, 0).Add(keeper.MaxEvidenceAge(ctx))})

	oldPower := sk.Validator(ctx, operatorAddr).GetConsensusPower()

	// double sign past max age
	keeper.HandleDoubleSign(ctx, val.Address(), 0, time.Unix(0, 0), power)

	// should still be bonded
	require.True(t, sk.Validator(ctx, operatorAddr).IsBonded())

	// should still have same power
	require.Equal(t, oldPower, sk.Validator(ctx, operatorAddr).GetConsensusPower())
}

// Test a validator through uptime, downtime, revocation,
// unrevocation, starting height reset, and revocation again
func TestHandleAbsentValidator(t *testing.T) {
	// initial setup
	ctx, ck, sk, _, keeper, _ := createTestInput(t, keeperTestParams())
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, val := addrs[0], pks[0]
	sh := staking.NewHandler(sk)
	slh := NewHandler(keeper)
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, initTokens.Sub(amt))),
	)
	require.Equal(t, amt, sk.Validator(ctx, addr).GetBondedTokens())

	// will exist since the validator has been bonded
	info, found := keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	//require.Equal(t, int64(0), info.StartHeight)
	//require.Equal(t, int64(0), info.IndexOffset)
	//require.Equal(t, int64(0), info.MissedBlocksCounter)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	height := int64(0)
	nextBlocktime := blockTimeGenerator(time.Minute)

	// 1000 first blocks OK
	for ; height < 1000; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		keeper.HandleValidatorSignature(ctx, val.Address(), power, true, blockWindow)
	}
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	//require.Equal(t, int64(0), info.StartHeight)
	//require.Equal(t, int64(0), info.MissedBlocksCounter)

	//for ; height < keeper.SignedBlocksWindow(ctx)+(keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx)); height++ {
	nextHeight := height + blockWindow - 3 // Approach the limit of missed signed blocks
	for ; height < nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		keeper.HandleValidatorSignature(ctx, val.Address(), power, false, blockWindow)
	}
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	//require.Equal(t, int64(0), info.StartHeight)
	//require.Equal(t, keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx), info.MissedBlocksCounter)

	// validator should be bonded still
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	bondPool := sk.GetBondedPool(ctx)
	require.True(sdk.IntEq(t, amt, bondPool.GetCoins().AmountOf(sk.BondDenom(ctx))))

	// 501st block missed
	ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
	keeper.HandleValidatorSignature(ctx, val.Address(), power, false, blockWindow)
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	//require.Equal(t, int64(0), info.StartHeight)
	// counter now reset to zero
	//require.Equal(t, int64(0), info.MissedBlocksCounter)

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should have been jailed
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())

	slashAmt := amt.ToDec().Mul(keeper.SlashFractionDowntime(ctx)).RoundInt64()

	// validator should have been slashed
	require.Equal(t, amt.Int64()-slashAmt, validator.GetTokens().Int64())

	// 502nd block *also* missed (since the LastCommit would have still included the just-unbonded validator)
	ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
	keeper.HandleValidatorSignature(ctx, val.Address(), power, false, blockWindow)
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
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
	got = slh(ctx, NewMsgUnjail(addr))
	require.False(t, got.IsOK())

	// unrevocation should succeed after jail expiration
	height++
	ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(5))

	got = slh(ctx, NewMsgUnjail(addr))
	require.True(t, got.IsOK())

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should be rebonded now
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// validator should have been slashed
	bondPool = sk.GetBondedPool(ctx)
	require.Equal(t, amt.Int64()-slashAmt, bondPool.GetCoins().AmountOf(sk.BondDenom(ctx)).Int64())

	// Validator start height should not have been changed
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	//require.Equal(t, int64(0), info.StartHeight)
	// we've missed 2 blocks more than the maximum, so the counter was reset to 0 at 1 block more and is now 1
	//require.Equal(t, int64(1), info.MissedBlocksCounter)

	// validator should not be immediately jailed again
	height++
	ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
	keeper.HandleValidatorSignature(ctx, val.Address(), power, false, blockWindow)
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// 500 signed blocks
	nextHeight = height + 501
	for ; height < nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		keeper.HandleValidatorSignature(ctx, val.Address(), power, false, blockWindow)
	}

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should be jailed again after 500 unsigned blocks
	nextHeight = height + blockWindow + 1
	for ; height <= nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		keeper.HandleValidatorSignature(ctx, val.Address(), power, false, blockWindow)
	}

	// end block
	staking.EndBlocker(ctx, sk)

	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())
	require.True(t, validator.IsJailed())
}

// Test a new validator entering the validator set
// Ensure that SigningInfo.StartHeight is set correctly
// and that they are not immediately jailed
func TestHandleNewValidator(t *testing.T) {
	nextBlocktime := blockTimeGenerator(time.Minute)

	// initial setup
	ctx, ck, sk, _, keeper, _ := createTestInput(t, keeperTestParams())
	addr, val := addrs[0], pks[0]
	amt := sdk.TokensFromConsensusPower(100)
	sh := staking.NewHandler(sk)

	// 1000 first blocks not a validator
	ctx = ctx.WithBlockHeight(1001).WithBlockTime(nextBlocktime(1001))

	// Validator created
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, initTokens.Sub(amt))),
	)
	require.Equal(t, amt, sk.Validator(ctx, addr).GetBondedTokens())

	// Now a validator, for two blocks
	keeper.HandleValidatorSignature(ctx, val.Address(), 100, true, blockWindow)
	ctx = ctx.WithBlockHeight(blockWindow + 2).WithBlockTime(nextBlocktime(2))
	keeper.HandleValidatorSignature(ctx, val.Address(), 100, false, blockWindow)

	info, found := keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	//require.Equal(t, blockWindow+1, info.StartHeight)
	//require.Equal(t, int64(2), info.IndexOffset)
	//require.Equal(t, int64(1), info.MissedBlocksCounter)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)

	// validator should be bonded still, should not have been jailed or slashed
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	bondPool := sk.GetBondedPool(ctx)
	expTokens := sdk.TokensFromConsensusPower(100)
	require.Equal(t, expTokens.Int64(), bondPool.GetCoins().AmountOf(sk.BondDenom(ctx)).Int64())
}

// Test a jailed validator being "down" twice
// Ensure that they're only slashed once
func TestHandleAlreadyJailed(t *testing.T) {
	nextBlocktime := blockTimeGenerator(time.Minute)

	// initial setup
	ctx, _, sk, _, keeper, supplyKeeper := createTestInput(t, DefaultParams())
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, val := addrs[0], pks[0]
	sh := staking.NewHandler(sk)
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	preSlashingSupply := supplyKeeper.GetSupply(ctx)

	// 1000 first blocks OK
	height := int64(0)
	for ; height < 1000; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		keeper.HandleValidatorSignature(ctx, val.Address(), power, true, blockWindow)
	}

	// 501 blocks missed
	for ; height < 1501; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		keeper.HandleValidatorSignature(ctx, val.Address(), power, false, blockWindow)
	}

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should have been jailed and slashed
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())

	// validator should have been slashed
	resultingTokens := amt.Sub(sdk.TokensFromConsensusPower(1))
	require.Equal(t, resultingTokens, validator.GetTokens())

	// another block missed
	ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
	keeper.HandleValidatorSignature(ctx, val.Address(), power, false, blockWindow)

	// validator should not have been slashed twice
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, resultingTokens, validator.GetTokens())
	require.Equal(t, preSlashingSupply, supplyKeeper.GetSupply(ctx))
}

// Test a validator dipping in and out of the validator set
// Ensure that missed blocks are tracked correctly and that
// the start height of the signing info is reset correctly
func TestValidatorDippingInAndOut(t *testing.T) {
	nextBlocktime := blockTimeGenerator(time.Minute)

	// initial setup
	// keeperTestParams set the SignedBlocksWindow to 1000 and MaxMissedBlocksPerWindow to 500
	ctx, _, sk, _, keeper, _ := createTestInput(t, keeperTestParams())
	params := sk.GetParams(ctx)
	params.MaxValidators = 1
	sk.SetParams(ctx, params)
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, val := addrs[0], pks[0]
	consAddr := sdk.ConsAddress(addr)
	sh := staking.NewHandler(sk)
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	// 100 first blocks OK
	height := int64(1)
	for ; height < int64(100); height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		keeper.HandleValidatorSignature(ctx, val.Address(), power, true, height)
	}

	// kick first validator out of validator set
	newAmt := sdk.TokensFromConsensusPower(101)
	got = sh(ctx, NewTestMsgCreateValidator(addrs[1], pks[1], newAmt))
	require.True(t, got.IsOK())
	validatorUpdates := staking.EndBlocker(ctx, sk)
	require.Equal(t, 2, len(validatorUpdates))
	validator, _ := sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Unbonding, validator.Status)

	// 600 more blocks happened
	height = int64(700)
	nextBlocktime(600)
	ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(0))

	// validator added back in
	delTokens := sdk.TokensFromConsensusPower(50)
	got = sh(ctx, newTestMsgDelegate(sdk.AccAddress(addrs[2]), addrs[0], delTokens))
	require.True(t, got.IsOK())
	validatorUpdates = staking.EndBlocker(ctx, sk)
	require.Equal(t, 2, len(validatorUpdates))
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Bonded, validator.Status)
	newPower := int64(150)

	// validator misses a block
	keeper.HandleValidatorSignature(ctx, val.Address(), newPower, false, blockWindow)
	height++

	// shouldn't be jailed/kicked yet
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Bonded, validator.Status)

	// validator misses 500 more blocks, 501 total
	latest := height
	for ; height < latest+500; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		keeper.HandleValidatorSignature(ctx, val.Address(), newPower, false, blockWindow)
	}

	// should now be jailed & kicked
	staking.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Unbonding, validator.Status)

	// check all the signing information
	signInfo, found := keeper.getValidatorSigningInfo(ctx, consAddr)
	require.True(t, found)
	require.Equal(t, consAddr, signInfo.Address)
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

	keeper.HandleValidatorSignature(ctx, val.Address(), newPower, true, blockWindow)
	height++

	// validator should not be kicked since we reset counter/array when it was jailed
	staking.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Bonded, validator.Status)

	// validator misses 501 blocks
	latest = height
	for ; height < latest+501; height++ {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(nextBlocktime(1))
		keeper.HandleValidatorSignature(ctx, val.Address(), newPower, false, blockWindow)
	}

	// validator should now be jailed & kicked
	staking.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Unbonding, validator.Status)
}

func blockTimeGenerator(blocktime time.Duration) func(int) time.Time {
	// TODO This might be a tiny bit over-engineered. Move to a simple struct?
	now := time.Now()

	return func(blockcount int) time.Time {
		now = now.Add(time.Duration(blockcount) * blocktime)
		return now
	}
}
