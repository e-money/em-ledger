package queries

import (
	"encoding/json"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/e-money/em-ledger/x/queries/types"
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

func calculateCirculatingSupply(ctx sdk.Context, accK AccountKeeper, bk BankKeeper) sdk.Coins {
	denomsSupply, stakingDenomIdx := getDenomsSupply(ctx, bk)

	ngmbalance := calcStakingSpendableSupply(ctx, accK, bk)

	// Replace staking token balance with the one calculated above, which omits
	// vesting and staked tokens.
	denomsSupply[stakingDenomIdx] = sdk.NewCoin(stakingDenom, ngmbalance)

	return denomsSupply
}

func calcStakingSpendableSupply(ctx sdk.Context, accK AccountKeeper, bk BankKeeper) sdk.Int {
	stakingAccounts := map[string]bool{
		accK.GetModuleAccount(ctx, stakingtypes.NotBondedPoolName).	GetAddress().String(): true,
		accK.GetModuleAccount(ctx, stakingtypes.BondedPoolName).GetAddress().String(): true,
	}

	ngmbalance := sdk.ZeroInt()

	bk.IterateAllBalances(
		ctx, func(address sdk.AccAddress, coin sdk.Coin) bool {
			if coin.Denom != stakingDenom {
				return false
			}

			if stakingAccounts[address.String()] {
				return false
			}

			spendableCoins := bk.SpendableCoins(ctx, address)
			ngmbalance = ngmbalance.Add(spendableCoins.AmountOf(stakingDenom))

			return false
		},
	)

	return ngmbalance
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
