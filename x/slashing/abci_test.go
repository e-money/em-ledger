// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package slashing

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/auth"
	db "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestBeginBlocker(t *testing.T) {
	database := db.NewMemDB()
	ctx, ck, sk, _, keeper, supplyKeeper := createTestInput(t, DefaultParams(), database)
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr1, pk1 := addrs[2], pks[2]
	addr2, pk2 := addrs[1], pks[1]

	// Verify that the penalty account is available and empty
	penalties := supplyKeeper.GetModuleAccount(ctx, PenaltyAccount).GetCoins()
	require.True(t, penalties.IsZero())

	// bond the validators
	_, err := staking.NewHandler(sk)(ctx, NewTestMsgCreateValidator(addr1, pk1, amt))
	require.NoError(t, err)
	_, err = staking.NewHandler(sk)(ctx, NewTestMsgCreateValidator(addr2, pk2, amt))
	require.NoError(t, err)

	staking.EndBlocker(ctx, sk)
	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr1)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, initTokens.Sub(amt))),
	)
	require.Equal(t, amt, sk.Validator(ctx, addr1).GetBondedTokens())

	val := abci.Validator{
		Address: pk1.Address(),
		Power:   amt.Int64(),
	}

	val2 := abci.Validator{
		Address: pk2.Address(),
		Power:   amt.Int64(),
	}

	// mark the validator as having signed
	req := abci.RequestBeginBlock{
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{
				{
					Validator:       val,
					SignedLastBlock: true,
				},
				{
					Validator:       val2,
					SignedLastBlock: true,
				},
			},
		},
	}

	batch := database.NewBatch()
	BeginBlocker(ctx, req, keeper, batch)
	batch.Write()

	info, found := keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(pk1.Address()))
	require.True(t, found)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	require.Equal(t, sdk.ConsAddress(pk1.Address()), info.Address)

	height := int64(0)

	// for 1000 blocks, mark the validators as having signed
	now := time.Now()
	for ; height < 1000; height++ {
		now = now.Add(5 * time.Minute)
		ctx = ctx.WithBlockHeight(height).WithBlockTime(now)
		req = abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val,
						SignedLastBlock: true,
					},
					{
						Validator:       val2,
						SignedLastBlock: true,
					},
				},
			},
		}
		batch = database.NewBatch()
		BeginBlocker(ctx, req, keeper, batch)
		batch.Write()
	}
	// for 500 blocks, mark the validator as having not signed. Other validator keeps signing.
	for ; height < 1500; height++ {
		now = now.Add(time.Minute)
		ctx = ctx.WithBlockHeight(height).WithBlockTime(now)
		req = abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val,
						SignedLastBlock: false,
					},
					{
						Validator:       val2,
						SignedLastBlock: true,
					},
				},
			},
		}
		batch = database.NewBatch()
		BeginBlocker(ctx, req, keeper, batch)
		batch.Write()
	}

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should be jailed
	validator, found := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk1))
	require.True(t, found)
	require.Equal(t, sdk.Unbonding, validator.GetStatus())

	// Verify that a fine has been added to the penalty account due to the jailing
	penalties = supplyKeeper.GetModuleAccount(ctx, PenaltyAccount).GetCoins()
	require.False(t, penalties.IsZero())

	// Verify that the penalty is distributed among the remaining validators
	now = now.Add(5 * time.Minute)
	ctx = ctx.WithBlockHeight(height).WithBlockTime(now)
	req = abci.RequestBeginBlock{
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{
				{
					Validator:       val2,
					SignedLastBlock: true,
				},
			},
		},
	}
	batch = database.NewBatch()
	BeginBlocker(ctx, req, keeper, batch)
	batch.Write()

	penalties = supplyKeeper.GetModuleAccount(ctx, PenaltyAccount).GetCoins()
	require.True(t, penalties.IsZero())

	// Penalty should now be in the fee account, ready to be distributed
	feeAccount := supplyKeeper.GetModuleAccount(ctx, auth.FeeCollectorName).GetCoins()
	require.False(t, feeAccount.IsZero())
}
