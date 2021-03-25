package queries

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/e-money/em-ledger/x/queries/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	accK AccountKeeper
	bk   BankKeeper
}

func NewQuerier(accK AccountKeeper, bk BankKeeper) *Querier {
	return &Querier{accK: accK, bk: bk}
}

func (k Querier) Circulating(c context.Context, req *types.QueryCirculatingRequest) (*types.QueryCirculatingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	var total sdk.Coins

	k.accK.IterateAccounts(ctx, func(account authtypes.AccountI) bool {
		if ma, ok := account.(*authtypes.ModuleAccount); ok {
			switch ma.Name {
			case stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName:
				return false
			}
		}

		coins := k.bk.SpendableCoins(ctx, account.GetAddress())
		total = total.Add(coins...)
		return false
	})

	return &types.QueryCirculatingResponse{Total: total}, nil
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
