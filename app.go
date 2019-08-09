package emoney

import (
	"emoney/x/mint"
	"encoding/json"
	"fmt"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"os"
)

const (
	appName = "emoneyd"
)

var (
	DefaultCLIHome  = os.ExpandEnv("$HOME/.emcli")
	DefaultNodeHome = os.ExpandEnv("$HOME/.emd")

	ModuleBasics = module.NewBasicManager(
		genaccounts.AppModuleBasic{},
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		supply.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		auth.FeeCollectorName:     nil,
		distr.ModuleName:          nil,
		mint.ModuleName:           {supply.Minter},
		staking.BondedPoolName:    {supply.Burner, supply.Staking},
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
		//gov.ModuleName:            {supply.Burner},
	}
)

type emoneyApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	keyMain    *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	keyParams  *sdk.KVStoreKey
	keySupply  *sdk.KVStoreKey
	keyStaking *sdk.KVStoreKey
	keyMint    *sdk.KVStoreKey
	keyDistr   *sdk.KVStoreKey

	tkeyParams  *sdk.TransientStoreKey
	tkeyStaking *sdk.TransientStoreKey

	accountKeeper auth.AccountKeeper
	paramsKeeper  params.Keeper
	bankKeeper    bank.Keeper
	supplyKeeper  supply.Keeper
	stakingKeeper staking.Keeper
	mintKeeper    mint.Keeper
	distrKeeper   distr.Keeper

	mm *module.Manager
}

type GenesisState map[string]json.RawMessage

func NewApp(logger log.Logger, db db.DB) *emoneyApp {
	cdc := MakeCodec()
	txDecoder := auth.DefaultTxDecoder(cdc)

	bApp := bam.NewBaseApp(appName, logger, db, txDecoder)

	application := &emoneyApp{
		BaseApp:     bApp,
		cdc:         cdc,
		keyMain:     sdk.NewKVStoreKey("main"),
		keyAccount:  sdk.NewKVStoreKey(auth.StoreKey),
		keyParams:   sdk.NewKVStoreKey(params.StoreKey),
		tkeyParams:  sdk.NewTransientStoreKey(params.TStoreKey),
		keyStaking:  sdk.NewKVStoreKey(staking.StoreKey),
		tkeyStaking: sdk.NewTransientStoreKey(staking.TStoreKey),
		keyMint:     sdk.NewKVStoreKey(mint.StoreKey),
		keyDistr:    sdk.NewKVStoreKey(distr.StoreKey),
		keySupply:   sdk.NewKVStoreKey(supply.StoreKey),
	}

	application.paramsKeeper = params.NewKeeper(cdc, application.keyParams, application.tkeyParams, params.DefaultCodespace)

	authSubspace := application.paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := application.paramsKeeper.Subspace(bank.DefaultParamspace)
	stakingSubspace := application.paramsKeeper.Subspace(staking.DefaultParamspace)
	mintSubspace := application.paramsKeeper.Subspace(mint.DefaultParamspace)
	distrSubspace := application.paramsKeeper.Subspace(distr.DefaultParamspace)

	application.accountKeeper = auth.NewAccountKeeper(cdc, application.keyAccount, authSubspace, auth.ProtoBaseAccount)
	application.bankKeeper = bank.NewBaseKeeper(application.accountKeeper, bankSubspace, bank.DefaultCodespace)
	application.supplyKeeper = supply.NewKeeper(cdc, application.keySupply, application.accountKeeper, application.bankKeeper, supply.DefaultCodespace, maccPerms)
	application.stakingKeeper = staking.NewKeeper(cdc, application.keyStaking, application.tkeyStaking, application.supplyKeeper,
		stakingSubspace, staking.DefaultCodespace)
	application.distrKeeper = distr.NewKeeper(application.cdc, application.keyDistr, distrSubspace, &application.stakingKeeper,
		application.supplyKeeper, distr.DefaultCodespace, auth.FeeCollectorName)

	application.mintKeeper = mint.NewKeeper(application.cdc, application.keyMint, mintSubspace, application.supplyKeeper, auth.FeeCollectorName)

	application.MountStores(application.keyMain, application.keyAccount, application.tkeyParams, application.keyParams,
		application.keySupply, application.keyStaking, application.tkeyStaking, application.keyMint, application.keyDistr)

	application.stakingKeeper = *application.stakingKeeper.SetHooks(staking.NewMultiStakingHooks(application.distrKeeper.Hooks()))

	application.mm = module.NewManager(
		genaccounts.NewAppModule(application.accountKeeper),
		genutil.NewAppModule(application.accountKeeper, application.stakingKeeper, application.BaseApp.DeliverTx),
		auth.NewAppModule(application.accountKeeper),
		bank.NewAppModule(application.bankKeeper, application.accountKeeper),
		supply.NewAppModule(application.supplyKeeper, application.accountKeeper),
		staking.NewAppModule(application.stakingKeeper, nil, application.accountKeeper, application.supplyKeeper),
		mint.NewAppModule(application.mintKeeper),
		distr.NewAppModule(application.distrKeeper, application.supplyKeeper),
	)

	application.mm.SetOrderBeginBlockers(mint.ModuleName, distr.ModuleName)
	application.mm.SetOrderEndBlockers(staking.ModuleName)
	application.mm.SetOrderInitGenesis(genaccounts.ModuleName, distr.ModuleName, staking.ModuleName, auth.ModuleName, bank.ModuleName, mint.ModuleName, supply.ModuleName, genutil.ModuleName)

	application.mm.RegisterRoutes(application.Router(), application.QueryRouter())

	application.SetInitChainer(application.InitChainer)
	application.SetAnteHandler(auth.NewAnteHandler(application.accountKeeper, application.supplyKeeper, auth.DefaultSigVerificationGasConsumer))
	application.SetBeginBlocker(application.BeginBlocker)
	application.SetEndBlocker(application.EndBlocker)

	err := application.LoadLatestVersion(application.keyMain)
	if err != nil {
		panic(err)
	}

	return application
}

func (app *emoneyApp) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "emz")
}

func (app *emoneyApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// application updates every end block
func (app *emoneyApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	//for _, acc := range app.accountKeeper.GetAllAccounts(ctx) {
	//	fmt.Printf("%v : %v [%T]\n", acc.GetAddress(), acc.GetCoins(), acc)
	//	//coins := acc.GetCoins()
	//	//for _, c := range coins {
	//	//	one := sdk.NewInt64Coin(c.Denom, 1)
	//	//	coins = coins.Add(sdk.NewCoins(one))
	//	//}
	//	//
	//	//app.bankKeeper.SetCoins(ctx, acc.GetAddress(), coins)
	//}

	block := ctx.BlockHeader()
	proposerAddress := block.GetProposerAddress()
	app.Logger(ctx).Info(fmt.Sprintf("Block %v proposed by %v", ctx.BlockHeight(), sdk.ValAddress(proposerAddress)))

	return app.mm.EndBlock(ctx, req)
}

// application update at chain initialization
func (app *emoneyApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) (res abci.ResponseInitChain) {
	var genesisState GenesisState
	app.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)
	return app.mm.InitGenesis(ctx, genesisState)
}

func init() {
	setGenesisDefaults()
}

func setGenesisDefaults() {
	// Override module defaults for use in testnets and the default init functionality.
	staking.DefaultGenesisState = stakingGenesisState

	mint.DefaultInflationState = mintDefaultInflationState()
}

func mintDefaultInflationState() func() mint.InflationState {
	mintDefaultInflationStateFn := mint.DefaultInflationState

	return func() mint.InflationState {
		state := mintDefaultInflationStateFn()
		return state
	}
}

func stakingGenesisState() stakingtypes.GenesisState {
	genesisState := stakingtypes.DefaultGenesisState()
	genesisState.Params.BondDenom = "ungm"

	return genesisState
}

func MakeCodec() *codec.Codec {
	cdc := codec.New()
	ModuleBasics.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}
