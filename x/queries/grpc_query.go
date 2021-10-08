package queries

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	total, err := calculateCirculatingSupply(ctx, k.accK, k.bk)
	if err != nil {
		return nil, status.Error(
			codes.Unavailable, fmt.Sprintf("supply store query erred:%v", err),
		)
	}

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
