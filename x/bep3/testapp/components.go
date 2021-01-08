package testapp

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/e-money/em-ledger/x/bep3"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func CreateTestComponents(t *testing.T) (sdk.Context, bep3.Keeper, auth.AccountKeeper, supply.Keeper, bep3.AppModule) {
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	keys := sdk.NewKVStoreKeys(bep3.StoreKey, auth.StoreKey, supply.StoreKey, params.StoreKey)
	for _, k := range keys {
		ms.MountStoreWithDB(k, sdk.StoreTypeIAVL, db)
	}

	tkeys := sdk.NewTransientStoreKeys(params.TStoreKey)
	for _, k := range tkeys {
		ms.MountStoreWithDB(k, sdk.StoreTypeTransient, db)
	}

	err := ms.LoadLatestVersion()
	require.NoError(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain"}, true, log.NewNopLogger())
	ctx = ctx.WithBlockTime(time.Now())

	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	sdk.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	bep3.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	cdc.Seal()

	mAccPerms := map[string][]string{
		bep3.ModuleName: {supply.Minter, supply.Burner},
	}

	paramsKeeper := params.NewKeeper(cdc, keys[params.StoreKey], tkeys[params.TStoreKey])
	var (
		authSubspace = paramsKeeper.Subspace(auth.DefaultParamspace)
		bankSubspace = paramsKeeper.Subspace(bank.DefaultParamspace)
		bep3Subspace = paramsKeeper.Subspace(bep3.DefaultParamspace)
	)

	var (
		accountKeeper = auth.NewAccountKeeper(cdc, keys[auth.StoreKey], authSubspace, auth.ProtoBaseAccount)
		bankKeeper    = bank.NewBaseKeeper(accountKeeper, bankSubspace, make(map[string]bool))
		supplyKeeper  = supply.NewKeeper(cdc, keys[supply.StoreKey], accountKeeper, bankKeeper, mAccPerms)
		bep3Keeper    = bep3.NewKeeper(cdc, keys[bep3.StoreKey], supplyKeeper, accountKeeper, bep3Subspace, make(map[string]bool))
	)

	supplyKeeper.SetSupply(ctx, supply.NewSupply(sdk.NewCoins()))

	return ctx, bep3Keeper, accountKeeper, supplyKeeper, bep3.NewAppModule(bep3Keeper, accountKeeper, supplyKeeper)
}
