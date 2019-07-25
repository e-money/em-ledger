package emoney

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/cosmos/cosmos-sdk/x/genaccounts"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	appName = "emoneyd"
)

var (
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		supply.AppModuleBasic{},
		staking.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		auth.FeeCollectorName: nil,
		//distr.ModuleName:          nil,
		//mint.ModuleName:           {supply.Minter},
		staking.BondedPoolName:    {supply.Burner, supply.Staking},
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
		//gov.ModuleName:            {supply.Burner},
	}
)

type sandboxApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	keyMain    *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	keyParams  *sdk.KVStoreKey
	keySupply  *sdk.KVStoreKey
	keyStaking *sdk.KVStoreKey

	tkeyParams  *sdk.TransientStoreKey
	tkeyStaking *sdk.TransientStoreKey

	accountKeeper auth.AccountKeeper
	paramsKeeper  params.Keeper
	bankKeeper    bank.Keeper
	supplyKeeper  supply.Keeper
	stakingKeeper staking.Keeper

	mm *module.Manager
}

type GenesisState map[string]json.RawMessage

func NewApp(logger log.Logger, db db.DB) *sandboxApp {
	cdc := MakeCodec()
	txDecoder := auth.DefaultTxDecoder(cdc)

	bApp := bam.NewBaseApp(appName, logger, db, txDecoder)

	application := &sandboxApp{
		BaseApp:     bApp,
		cdc:         cdc,
		keyMain:     sdk.NewKVStoreKey("main"),
		keyAccount:  sdk.NewKVStoreKey(auth.StoreKey),
		keyParams:   sdk.NewKVStoreKey(params.StoreKey),
		tkeyParams:  sdk.NewTransientStoreKey(params.TStoreKey),
		keyStaking:  sdk.NewKVStoreKey(staking.StoreKey),
		tkeyStaking: sdk.NewTransientStoreKey(staking.TStoreKey),
		keySupply:   sdk.NewKVStoreKey(supply.StoreKey),
	}

	application.paramsKeeper = params.NewKeeper(cdc, application.keyParams, application.tkeyParams, params.DefaultCodespace)

	authSubspace := application.paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := application.paramsKeeper.Subspace(bank.DefaultParamspace)
	stakingSubspace := application.paramsKeeper.Subspace(staking.DefaultParamspace)

	application.accountKeeper = auth.NewAccountKeeper(cdc, application.keyAccount, authSubspace, auth.ProtoBaseAccount)
	application.bankKeeper = bank.NewBaseKeeper(application.accountKeeper, bankSubspace, bank.DefaultCodespace)
	application.supplyKeeper = supply.NewKeeper(cdc, application.keySupply, application.accountKeeper, application.bankKeeper, supply.DefaultCodespace, maccPerms)

	application.stakingKeeper = staking.NewKeeper(cdc, application.keyStaking, application.tkeyStaking, application.supplyKeeper,
		stakingSubspace, staking.DefaultCodespace)

	application.MountStores(application.keyMain, application.keyAccount, application.tkeyParams, application.keyParams,
		application.keySupply, application.keyStaking, application.tkeyStaking)

	application.mm = module.NewManager(
		genaccounts.NewAppModule(application.accountKeeper),
		auth.NewAppModule(application.accountKeeper),
		bank.NewAppModule(application.bankKeeper, application.accountKeeper),
		supply.NewAppModule(application.supplyKeeper, application.accountKeeper),
		staking.NewAppModule(application.stakingKeeper, nil, application.accountKeeper, application.supplyKeeper),
	)

	application.mm.SetOrderEndBlockers(staking.ModuleName)
	application.mm.SetOrderInitGenesis(genaccounts.ModuleName, staking.ModuleName, auth.ModuleName, bank.ModuleName, supply.ModuleName)

	application.mm.RegisterRoutes(application.Router(), application.QueryRouter())

	application.SetInitChainer(application.InitChainer)
	application.SetAnteHandler(auth.NewAnteHandler(application.accountKeeper, application.supplyKeeper, auth.DefaultSigVerificationGasConsumer))
	application.SetEndBlocker(application.EndBlocker)

	err := application.LoadLatestVersion(application.keyMain)
	if err != nil {
		panic(err)
	}

	return application
}

// application updates every end block
func (app *sandboxApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	fmt.Println(" *** Iterating accounts")
	for _, acc := range app.accountKeeper.GetAllAccounts(ctx) {
		fmt.Println(acc)
		//coins := acc.GetCoins()
		//for _, c := range coins {
		//	one := sdk.NewInt64Coin(c.Denom, 1)
		//	coins = coins.Add(sdk.NewCoins(one))
		//}
		//
		//app.bankKeeper.SetCoins(ctx, acc.GetAddress(), coins)
	}
	return app.mm.EndBlock(ctx, req)
}

// application update at chain initialization
func (app *sandboxApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) (res abci.ResponseInitChain) {
	var genesisState GenesisState
	app.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)

	res = app.mm.InitGenesis(ctx, genesisState)
	if len(req.Validators) > 0 {
		// NOTE : Initially manually set the list of validators here. Should eventually be set by the Staking module.
		res.Validators = req.Validators
	}

	return res
}

func MakeCodec() *codec.Codec {
	cdc := codec.New()
	ModuleBasics.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}
