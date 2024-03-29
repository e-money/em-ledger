package queries

import (
	"context"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/e-money/em-ledger/x/queries/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const stakingDenom = "ungm"

var _ types.QueryServer = Querier{}

type Querier struct {
	accK AccountKeeper
	bk   BankKeeper
	sk   SlashingKeeper
}

func NewQuerier(accK AccountKeeper, bk BankKeeper, sk SlashingKeeper) *Querier {
	return &Querier{accK: accK, bk: bk, sk: sk}
}

func (k Querier) Circulating(c context.Context, req *types.QueryCirculatingRequest) (*types.QueryCirculatingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	total := calculateCirculatingSupply(ctx, k.accK, k.bk)

	return &types.QueryCirculatingResponse{Total: total}, nil
}

func calculateCirculatingSupply(ctx sdk.Context, accK AccountKeeper, bk BankKeeper) sdk.Coins {
	denomsSupply, stakingDenomIdx := getDenomsSupply(ctx, bk)

	ngmbalance := calcStakingSpendableSupply(ctx, accK, bk)

	// Replace staking token balance with the one calculated above, which omits
	// vesting and staked tokens.
	denomsSupply[stakingDenomIdx] = sdk.NewCoin(stakingDenom, ngmbalance)

	return denomsSupply
}

func getDenomsSupply(ctx sdk.Context, bk BankKeeper) (sdk.Coins, int) {
	denoms := bk.GetAllDenomMetaData(ctx)
	sort.Slice(denoms, func(i, j int) bool {
		return denoms[i].Base < denoms[j].Base
	})

	var (
		denomsSupply    sdk.Coins
		stakingDenomIdx int
	)
	for idx, denom := range denoms {
		if denom.Base == stakingDenom {
			stakingDenomIdx = idx
			// 0 : the staking supply is calculated later
			denomsSupply = append(
				denomsSupply, sdk.NewCoin(stakingDenom, sdk.ZeroInt()),
			)
			continue
		}
		denomsSupply = append(denomsSupply, bk.GetSupply(ctx, denom.Base))
	}

	return denomsSupply, stakingDenomIdx
}

func calcStakingSpendableSupply(ctx sdk.Context, accK AccountKeeper, bk BankKeeper) sdk.Int {
	stakingAccounts := []sdk.AccAddress{
		accK.GetModuleAccount(ctx, stakingtypes.NotBondedPoolName).GetAddress(),
		accK.GetModuleAccount(ctx, stakingtypes.BondedPoolName).GetAddress(),
	}

	bondedAndUnbondingBalance := sdk.ZeroInt()
	for _, acc := range stakingAccounts {
		bal := bk.GetBalance(ctx, acc, "ungm")
		bondedAndUnbondingBalance = bondedAndUnbondingBalance.Add(bal.Amount)
	}

	totalSupply := bk.GetSupply(ctx, stakingDenom)
	return totalSupply.Amount.Sub(bondedAndUnbondingBalance)
}

func (k Querier) Spendable(c context.Context, req *types.QuerySpendableRequest) (*types.QuerySpendableResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	spendableBalance := k.bk.SpendableCoins(ctx, address)
	return &types.QuerySpendableResponse{Balance: spendableBalance}, nil
}

func (k Querier) MissedBlocks(c context.Context, req *types.QueryMissedBlocksRequest) (*types.QueryMissedBlocksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	consAddr, err := sdk.ConsAddressFromBech32(req.ConsAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid validator consensus address: "+err.Error())
	}

	missedBlocksCnt, blocksCnt := k.sk.GetMissedBlocks(ctx, consAddr)
	return &types.QueryMissedBlocksResponse{
		MissedBlocksInfo: types.MissedBlocksInfo{
			ConsAddress:         req.ConsAddress,
			MissedBlocksCounter: missedBlocksCnt,
			TotalBlocksCounter:  blocksCnt,
		},
	}, nil
}
