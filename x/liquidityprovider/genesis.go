// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package liquidityprovider

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
)

func defaultGenesisState() *types.GenesisState {
	return &types.GenesisState{}
}

func InitGenesis(ctx sdk.Context, keeper Keeper, gs types.GenesisState) error {
	for _, lp := range gs.Accounts {
		addr, err := sdk.AccAddressFromBech32(lp.Account)
		if err != nil {
			return sdkerrors.Wrapf(err, "address: %s", lp.Account)
		}
		_, err = keeper.CreateLiquidityProvider(ctx, addr, lp.Mintable)
		if err != nil {
			return sdkerrors.Wrap(err, "liquidity provider")
		}
	}
	return nil
}
