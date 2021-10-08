package queries

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"math"

	"github.com/e-money/em-ledger/x/queries/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

const stakingDenom = "ungm"

func NewLegacyQuerier(accK AccountKeeper, bk BankKeeper) sdk.Querier {
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
	total, err := calculateCirculatingSupply(ctx, accK, bk)
	if err != nil {
		return nil, err
	}
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

func calculateCirculatingSupply(ctx sdk.Context, accK AccountKeeper, bk BankKeeper) (circSupply sdk.Coins, err error) {
	total, _, err := bk.GetPaginatedTotalSupply(ctx, &query.PageRequest{Limit: math.MaxUint64})
	if err != nil {
		return circSupply, err
	}

	stakingAccounts := map[string]bool{
		accK.GetModuleAccount(ctx, stakingtypes.NotBondedPoolName).GetAddress().String(): true,
		accK.GetModuleAccount(ctx, stakingtypes.BondedPoolName).GetAddress().String():    true,
	}

	ngmbalance := sdk.ZeroInt()

	bk.IterateAllBalances(ctx, func(address sdk.AccAddress, coin sdk.Coin) bool {
		if coin.Denom != stakingDenom {
			return false
		}

		if _, stakingModule := stakingAccounts[address.String()]; stakingModule {
			return false
		}

		spendableCoins := bk.SpendableCoins(ctx, address)
		ngmbalance = ngmbalance.Add(spendableCoins.AmountOf("ungm"))

		return false
	})

	// Replace staking token balance with the one calculated above, which omits vesting and staked tokens.
	for i, c := range total {
		if c.Denom != stakingDenom {
			continue
		}

		total[i] = sdk.NewCoin(stakingDenom, ngmbalance)
		break
	}

	return total, nil
}
