package slashing

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestBeginBlocker(t *testing.T) {
	ctx, ck, sk, _, keeper := createTestInput(t, DefaultParams())
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, pk := addrs[2], pks[2]

	// bond the validator
	got := staking.NewHandler(sk)(ctx, NewTestMsgCreateValidator(addr, pk, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)
	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, initTokens.Sub(amt))),
	)
	require.Equal(t, amt, sk.Validator(ctx, addr).GetBondedTokens())

	val := abci.Validator{
		Address: pk.Address(),
		Power:   amt.Int64(),
	}

	// mark the validator as having signed
	req := abci.RequestBeginBlock{
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{{
				Validator:       val,
				SignedLastBlock: true,
			}},
		},
	}
	BeginBlocker(ctx, req, keeper)

	info, found := keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(pk.Address()))
	require.True(t, found)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	require.Equal(t, sdk.ConsAddress(pk.Address()), info.Address)

	height := int64(0)

	// for 1000 blocks, mark the validator as having signed
	now := time.Now()
	for ; height < 1000; height++ {
		now = now.Add(5 * time.Second)
		ctx = ctx.WithBlockHeight(height).WithBlockTime(now)
		req = abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{{
					Validator:       val,
					SignedLastBlock: true,
				}},
			},
		}
		BeginBlocker(ctx, req, keeper)
	}
	// for 500 blocks, mark the validator as having not signed
	for ; height < 1500; height++ {
		now = now.Add(5 * time.Second)
		ctx = ctx.WithBlockHeight(height).WithBlockTime(now)
		req = abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{{
					Validator:       val,
					SignedLastBlock: false,
				}},
			},
		}
		BeginBlocker(ctx, req, keeper)
	}

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should be jailed
	validator, found := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk))
	require.True(t, found)
	require.Equal(t, sdk.Unbonding, validator.GetStatus())
}
