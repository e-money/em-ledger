package queries

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/e-money/em-ledger/x/queries/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(accK AccountKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryCirculating:
			return queryCirculatingSupply(ctx, accK)
		case types.QuerySpendable:
			return querySpendableBalance(ctx, accK, path[1:])
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query for endpoint: %s", path[0])
		}
	}
}

func queryCirculatingSupply(ctx sdk.Context, accK AccountKeeper) (res []byte, err error) {
	var total sdk.Coins

	accK.IterateAccounts(ctx, func(account exported.Account) bool {
		if ma, ok := account.(*supply.ModuleAccount); ok {
			switch ma.Name {
			case staking.NotBondedPoolName, staking.BondedPoolName:
				return false
			}
		}

		coins := account.SpendableCoins(ctx.BlockTime())
		total = total.Add(coins...)
		return false
	})

	return json.Marshal(total)
}

func querySpendableBalance(ctx sdk.Context, accK AccountKeeper, path []string) (res []byte, err error) {
	address, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, err
	}

	account := accK.GetAccount(ctx, address)
	if account == nil {
		return nil, nil
	}

	spendableBalance := account.SpendableCoins(ctx.BlockTime())
	return json.Marshal(spendableBalance)
}
