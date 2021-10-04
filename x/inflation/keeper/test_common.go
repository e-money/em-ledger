// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// nolint:deadcode unused
package keeper

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/e-money/em-ledger/x/inflation/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

type testInput struct {
	ctx        sdk.Context
	cdc        codec.Codec
	mintKeeper Keeper
	encConfig  simappparams.EncodingConfig
}

func newTestInput(t *testing.T) testInput {
	t.Helper()
	encConfig := MakeTestEncodingConfig()
	var (
		keyInflation = sdk.NewKVStoreKey(types.ModuleName)
		bankKey      = sdk.NewKVStoreKey(banktypes.ModuleName)
		authCapKey   = sdk.NewKVStoreKey("authCapKey")
		keyParams    = sdk.NewKVStoreKey("params")
		stakingKey   = sdk.NewKVStoreKey("staking")
		authKey      = sdk.NewKVStoreKey(authtypes.StoreKey)
		tkeyParams   = sdk.NewTransientStoreKey("transient_params")

		blockedAddrs = make(map[string]bool)
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyInflation, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(stakingKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(bankKey, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.NoError(t, err)

	maccPerms := map[string][]string{
		types.ModuleName:               {authtypes.Minter},
		authtypes.FeeCollectorName:     nil,
		"buyback":                      {authtypes.Burner},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	}

	pk := paramskeeper.NewKeeper(encConfig.Marshaler, encConfig.Amino, keyParams, tkeyParams)

	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: "test-chain"}, true, log.NewNopLogger())

	accountKeeper := authkeeper.NewAccountKeeper(
		encConfig.Marshaler, authCapKey, pk.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
	)

	bankKeeper := bankkeeper.NewBaseKeeper(
		encConfig.Marshaler, bankKey, accountKeeper, pk.Subspace(banktypes.ModuleName), blockedAddrs,
	)

	stakingKeeper := mockStakingKeeper{}

	inflationKeeper := NewKeeper(
		encConfig.Marshaler, keyInflation, bankKeeper, accountKeeper, stakingKeeper, "buyback", authtypes.FeeCollectorName,
	)
	inflationKeeper.SetState(ctx, types.NewInflationState(time.Now(), "ejpy", "0.05", "echf", "0.10", "eeur", "0.01"))

	//// set module accounts
	//feeCollectorAcc := supply.NewEmptyModuleAccount(auth.FeeCollectorName)
	//minterAcc := supply.NewEmptyModuleAccount(types.ModuleName, supply.Minter)
	//notBondedPool := supply.NewEmptyModuleAccount(staking.NotBondedPoolName, supply.Burner)
	//bondPool := supply.NewEmptyModuleAccount(staking.BondedPoolName, supply.Burner)
	//
	//supplyKeeper.SetModuleAccount(ctx, feeCollectorAcc)
	//supplyKeeper.SetModuleAccount(ctx, minterAcc)
	//supplyKeeper.SetModuleAccount(ctx, notBondedPool)
	//supplyKeeper.SetModuleAccount(ctx, bondPool)

	return testInput{
		ctx:        ctx,
		cdc:        encConfig.Marshaler,
		mintKeeper: inflationKeeper,
		encConfig:  encConfig,
	}
}

func MakeTestEncodingConfig() simappparams.EncodingConfig {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	encodingConfig := simappparams.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             cdc,
	}

	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics := module.NewBasicManager(
		bank.AppModuleBasic{},
		auth.AppModuleBasic{},
	)

	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

type mockStakingKeeper struct{}

func (m mockStakingKeeper) GetParams(_ sdk.Context) stakingtypes.Params {
	return stakingtypes.NewParams(5*time.Minute, 40, 50, 0, "ungm")
}
