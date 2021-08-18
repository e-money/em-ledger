package v040

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v040"
	v036supply "github.com/cosmos/cosmos-sdk/x/bank/legacy/v036"
	v038bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v038"
	v040bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v040"
	v039crisis "github.com/cosmos/cosmos-sdk/x/crisis/legacy/v039"
	v040crisis "github.com/cosmos/cosmos-sdk/x/crisis/legacy/v040"
	v036distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v036"
	v038distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v038"
	v040distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v040"
	v038evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v038"
	v040evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v040"
	v039genutil "github.com/cosmos/cosmos-sdk/x/genutil/legacy/v039"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	v036params "github.com/cosmos/cosmos-sdk/x/params/legacy/v036"
	v038staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v038"
	v040staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v040"
	v038upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/legacy/v038"
	"github.com/e-money/em-ledger/x/authority"
	v09authority "github.com/e-money/em-ledger/x/authority/legacy/v09"
	v09liquidityprovider "github.com/e-money/em-ledger/x/liquidityprovider/legacy/v09"
	v09slashing "github.com/e-money/em-ledger/x/slashing/legacy/v09"
)

func migrateGenutil(oldGenState v039genutil.GenesisState) *types.GenesisState {
	return &types.GenesisState{
		GenTxs: oldGenState.GenTxs,
	}
}

// Migrate migrates exported state from v0.9.x to a v1.0.x genesis state.
func Migrate(appState types.AppMap, clientCtx client.Context) types.AppMap {
	v09Codec := codec.NewLegacyAmino()
	v039auth.RegisterLegacyAminoCodec(v09Codec)
	v036distr.RegisterLegacyAminoCodec(v09Codec)
	v036params.RegisterLegacyAminoCodec(v09Codec)
	v038upgrade.RegisterLegacyAminoCodec(v09Codec)

	v09liquidityprovider.RegisterLegacyAminoCodec(v09Codec)

	v040Codec := clientCtx.JSONMarshaler

	if appState[v038bank.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var bankGenState v038bank.GenesisState
		v09Codec.MustUnmarshalJSON(appState[v038bank.ModuleName], &bankGenState)

		// unmarshal x/auth genesis state to retrieve all account balances
		var authGenState v039auth.GenesisState
		v09Codec.MustUnmarshalJSON(appState[v039auth.ModuleName], &authGenState)

		// unmarshal x/supply genesis state to retrieve total supply
		var supplyGenState v036supply.GenesisState
		v09Codec.MustUnmarshalJSON(appState[v036supply.ModuleName], &supplyGenState)

		// delete deprecated x/bank genesis state
		delete(appState, v038bank.ModuleName)

		// delete deprecated x/supply genesis state
		delete(appState, v036supply.ModuleName)

		// Remove Liquidity provider accounts before migrating acocunts.
		for i, v09account := range authGenState.Accounts {
			switch v09account := v09account.(type) {
			case *v09liquidityprovider.LiquidityProviderAccount:
				authGenState.Accounts[i] = v09account.BaseAccount
			}
		}

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v040bank.ModuleName] = v040Codec.MustMarshalJSON(v040bank.Migrate(bankGenState, authGenState, supplyGenState))
	}

	// remove balances from existing accounts
	if appState[v039auth.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var authGenState v039auth.GenesisState
		v09Codec.MustUnmarshalJSON(appState[v039auth.ModuleName], &authGenState)

		// delete deprecated x/auth genesis state
		delete(appState, v039auth.ModuleName)

		// Remove Liquidity provider accounts before migrating acocunts.
		for i, v09account := range authGenState.Accounts {
			switch v09account := v09account.(type) {
			case *v09liquidityprovider.LiquidityProviderAccount:
				authGenState.Accounts[i] = v09account.BaseAccount
			}
		}

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v040auth.ModuleName] = v040Codec.MustMarshalJSON(v040auth.Migrate(authGenState))
	}

	// Migrate x/authority
	if appState[authority.ModuleName] != nil {
		var authorityGenState v09authority.GenesisState
		v09Codec.MustUnmarshalJSON(appState[authority.ModuleName], &authorityGenState)

		delete(appState, authority.ModuleName)

		appState[authority.ModuleName] = v040Codec.MustMarshalJSON(v09authority.Migrate(authorityGenState))
	}

	// Migrate x/crisis.
	if appState[v039crisis.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var crisisGenState v039crisis.GenesisState
		v09Codec.MustUnmarshalJSON(appState[v039crisis.ModuleName], &crisisGenState)

		// delete deprecated x/crisis genesis state
		delete(appState, v039crisis.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v040crisis.ModuleName] = v040Codec.MustMarshalJSON(v040crisis.Migrate(crisisGenState))
	}

	// Migrate x/distribution.
	if appState[v038distr.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var distributionGenState v038distr.GenesisState
		v09Codec.MustUnmarshalJSON(appState[v038distr.ModuleName], &distributionGenState)

		// delete deprecated x/distribution genesis state
		delete(appState, v038distr.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v040distr.ModuleName] = v040Codec.MustMarshalJSON(v040distr.Migrate(distributionGenState))
	}

	// Migrate x/evidence.
	if appState[v038evidence.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var evidenceGenState v038evidence.GenesisState
		v09Codec.MustUnmarshalJSON(appState[v038bank.ModuleName], &evidenceGenState)

		// delete deprecated x/evidence genesis state
		delete(appState, v038evidence.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v040evidence.ModuleName] = v040Codec.MustMarshalJSON(v040evidence.Migrate(evidenceGenState))
	}

	// Migrate x/slashing.
	if appState[v09slashing.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var slashingGenState v09slashing.GenesisState
		v09Codec.MustUnmarshalJSON(appState[v09slashing.ModuleName], &slashingGenState)

		// delete deprecated x/slashing genesis state
		delete(appState, v09slashing.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v09slashing.ModuleName] = v040Codec.MustMarshalJSON(v09slashing.Migrate(slashingGenState))
	}

	// Migrate x/staking.
	if appState[v038staking.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var stakingGenState v038staking.GenesisState
		v09Codec.MustUnmarshalJSON(appState[v038staking.ModuleName], &stakingGenState)

		// delete deprecated x/staking genesis state
		delete(appState, v038staking.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v040staking.ModuleName] = v040Codec.MustMarshalJSON(v040staking.Migrate(stakingGenState))
	}

	// Migrate x/genutil
	if appState[v039genutil.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var genutilGenState v039genutil.GenesisState
		v09Codec.MustUnmarshalJSON(appState[v039genutil.ModuleName], &genutilGenState)

		// delete deprecated x/staking genesis state
		delete(appState, v039genutil.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[ModuleName] = v040Codec.MustMarshalJSON(migrateGenutil(genutilGenState))
	}

	return appState
}
