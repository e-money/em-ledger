// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package authority

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/authority/types"
)

func NewGenesisState(authorityKey sdk.AccAddress, gasPrices sdk.DecCoins) types.GenesisState {
	return types.GenesisState{
		AuthorityKey: authorityKey.String(),
		MinGasPrices: gasPrices,
	}
}

func DefaultGenesisState() *types.GenesisState {
	return &types.GenesisState{}
}

func InitGenesis(ctx sdk.Context, keeper Keeper, state types.GenesisState) error {
	authKey, err := sdk.AccAddressFromBech32(state.AuthorityKey)
	if err != nil {
		return sdkerrors.Wrap(err, "authority key")
	}
	keeper.SetAuthority(ctx, authKey)
	keeper.SetGasPrices(ctx, authKey, state.MinGasPrices)
	return nil
}
