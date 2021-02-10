package queries

import (
	"encoding/json"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/e-money/em-ledger/x/queries/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(accK AccountKeeper, bk BankKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryCirculating:
			return queryCirculatingSupply(ctx, accK, bk)
		case types.QuerySpendable:
			return querySpendableBalance(ctx, bk, path[1:])
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query for endpoint: %s", path[0])
		}
	}
}

func queryCirculatingSupply(ctx sdk.Context, accK AccountKeeper, bk BankKeeper) (res []byte, err error) {
	var total sdk.Coins

	accK.IterateAccounts(ctx, func(account authtypes.AccountI) bool {
		if ma, ok := account.(*authtypes.ModuleAccount); ok {
			switch ma.Name {
			case stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName:
				return false
			}
		}

		coins := bk.SpendableCoins(ctx, account.GetAddress())
		total = total.Add(coins...)
		return false
	})

	return json.Marshal(total)
}

func querySpendableBalance(ctx sdk.Context, k BankKeeper, path []string) (res []byte, err error) {
	address, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, err
	}

	spendableBalance := k.SpendableCoins(ctx, address)
	return json.Marshal(spendableBalance)
}
