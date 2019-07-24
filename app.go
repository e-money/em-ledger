package tmsandbox

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/genaccounts"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	//"github.com/cosmos/cosmos-sdk/x/supply"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	appName = "sandbox"
)

var (
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		//supply.AppModuleBasic{},
	)
)

type sandboxApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	keyMain          *sdk.KVStoreKey
	keyAccount       *sdk.KVStoreKey
	keyFeeCollection *sdk.KVStoreKey
	keyParams        *sdk.KVStoreKey
	tkeyParams       *sdk.TransientStoreKey

	accountKeeper auth.AccountKeeper
	paramsKeeper  params.Keeper
	bankKeeper    bank.Keeper
	//supplyKeeper        supply.Keeper

	mm *module.Manager
}

type GenesisState map[string]json.RawMessage

func NewApp(logger log.Logger, db db.DB) *sandboxApp {
	cdc := MakeCodec()
	txDecoder := auth.DefaultTxDecoder(cdc)

	bApp := bam.NewBaseApp(appName, logger, db, txDecoder)

	application := &sandboxApp{
		BaseApp:    bApp,
		cdc:        cdc,
		keyMain:    sdk.NewKVStoreKey("main"),
		keyAccount: sdk.NewKVStoreKey(auth.StoreKey),
		keyParams:  sdk.NewKVStoreKey(params.StoreKey),
		tkeyParams: sdk.NewTransientStoreKey(params.TStoreKey),
	}

	application.paramsKeeper = params.NewKeeper(cdc, application.keyParams, application.tkeyParams, params.DefaultCodespace)

	authSubspace := application.paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := application.paramsKeeper.Subspace(bank.DefaultParamspace)

	application.accountKeeper = auth.NewAccountKeeper(cdc, application.keyAccount, authSubspace, auth.ProtoBaseAccount)
	application.bankKeeper = bank.NewBaseKeeper(application.accountKeeper, bankSubspace, bank.DefaultCodespace)

	application.MountStores(application.keyMain, application.keyAccount, application.keyFeeCollection, application.tkeyParams, application.keyParams)

	application.mm = module.NewManager(
		genaccounts.NewAppModule(application.accountKeeper),
		auth.NewAppModule(application.accountKeeper),
		bank.NewAppModule(application.bankKeeper, application.accountKeeper),
	)

	application.mm.SetOrderInitGenesis(genaccounts.ModuleName, auth.ModuleName, bank.ModuleName)

	application.mm.RegisterRoutes(application.Router(), application.QueryRouter())

	application.SetInitChainer(application.InitChainer)
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
