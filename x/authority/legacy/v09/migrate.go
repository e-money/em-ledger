package v09

import v10authority "github.com/e-money/em-ledger/x/authority/types"

func Migrate(authorityGenState GenesisState) *v10authority.GenesisState {
	return &v10authority.GenesisState{
		AuthorityKey: authorityGenState.AuthorityKey.String(),
		MinGasPrices: authorityGenState.MinGasPrices,
	}
}
