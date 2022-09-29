package queries

import (
	"encoding/json"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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
	total := calculateCirculatingSupply(ctx, accK, bk)
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

func calculateCirculatingSupply(ctx sdk.Context, accK AccountKeeper, bk BankKeeper) (circSupply sdk.Coins) {
	stakingAccounts := []sdk.AccAddress{
		accK.GetModuleAccount(ctx, stakingtypes.NotBondedPoolName).GetAddress(),
		accK.GetModuleAccount(ctx, stakingtypes.BondedPoolName).GetAddress(),
	}

	bondedAndUnbondingBalance := sdk.ZeroInt()
	for _, acc := range stakingAccounts {
		bal := bk.GetBalance(ctx, acc, "ungm")
		bondedAndUnbondingBalance = bondedAndUnbondingBalance.Add(bal.Amount)
	}

	total := bk.GetSupply(ctx).GetTotal()
	// Replace staking token balance with the one calculated above, which omits staked tokens.
	for i, c := range total {
		if c.Denom != stakingDenom {
			continue
		}

		total[i] = total[i].Sub(sdk.NewCoin("ungm", bondedAndUnbondingBalance))
		break
	}

	return total
}
