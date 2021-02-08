// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package authority

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	emtypes "github.com/e-money/em-ledger/types"
)

type GenesisState struct {
	AuthorityKey     string                   `json:"key" yaml:"key"`
	RestrictedDenoms emtypes.RestrictedDenoms `json:"restricted_denoms" yaml:"restricted_denoms"`
	MinGasPrices     sdk.DecCoins             `json:"min_gas_prices" yaml:"min_gas_prices"`
}

func NewGenesisState(authorityKey sdk.AccAddress, restrictedDenoms emtypes.RestrictedDenoms, gasPrices sdk.DecCoins) GenesisState {
	return GenesisState{
		AuthorityKey:     authorityKey.String(),
		RestrictedDenoms: restrictedDenoms,
		MinGasPrices:     gasPrices,
	}
}

func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

func InitGenesis(ctx sdk.Context, keeper Keeper, state GenesisState) error {
	authKey, err := sdk.AccAddressFromBech32(state.AuthorityKey)
	if err != nil {
		return sdkerrors.Wrap(err, "authority key")
	}
	keeper.SetAuthority(ctx, authKey)
	keeper.SetRestrictedDenoms(ctx, state.RestrictedDenoms)
	keeper.SetGasPrices(ctx, authKey, state.MinGasPrices)
	return nil
}
