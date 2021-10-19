// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package emoney

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	ibcconnectiontypes "github.com/cosmos/ibc-go/modules/core/03-connection/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	sdkante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/ibc-go/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/modules/core"
	channelkeeper "github.com/cosmos/ibc-go/modules/core/04-channel/keeper"
	porttypes "github.com/cosmos/ibc-go/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/modules/core/keeper"
	embank "github.com/e-money/em-ledger/hooks/bank"
	apptypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/auth/ante"
	"github.com/e-money/em-ledger/x/authority"
	"github.com/e-money/em-ledger/x/buyback"
	emdistr "github.com/e-money/em-ledger/x/distribution"
	"github.com/e-money/em-ledger/x/inflation"
	"github.com/e-money/em-ledger/x/issuer"
	"github.com/e-money/em-ledger/x/liquidityprovider"
	lptypes "github.com/e-money/em-ledger/x/liquidityprovider/types"
	"github.com/e-money/em-ledger/x/market"
	"github.com/e-money/em-ledger/x/queries"
	emslashing "github.com/e-money/em-ledger/x/slashing"
	"github.com/e-money/em-ledger/x/staking"
	historykeeper "github.com/e-money/em-ledger/x/staking/keeper"
	"github.com/e-money/em-ledger/x/upgrade"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"
	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	db "github.com/tendermint/tm-db"
	dbm "github.com/tendermint/tm-db"
)

const (
	appName = "emoneyd"
)

var (
	DefaultNodeHome = os.ExpandEnv("$HOME/.emd")

	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		emdistr.AppModuleBasic{},
		// todo (reviewer) : gov was deactivated in the original app.go but for ICS-20 we should have the `upgradeclient` handlers
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		emslashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
		vesting.AppModuleBasic{},
		// em modules
		inflation.AppModuleBasic{},
		liquidityprovider.AppModuleBasic{},
		issuer.AppModuleBasic{},
		authority.AppModule{},
		market.AppModule{},
		buyback.AppModule{},
		queries.AppModule{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		emdistr.ModuleName:             nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		// em modules
		inflation.ModuleName:         {authtypes.Minter},
		emslashing.ModuleName:        nil, // TODO Remove this line?
		liquidityprovider.ModuleName: {authtypes.Minter, authtypes.Burner},
		buyback.ModuleName:           {authtypes.Burner},
	}

	// module accounts that are allowed to receive tokens
	allowedReceivingModAcc = map[string]bool{
		emdistr.ModuleName: true,
	}
)
var (
	_ simapp.App              = (*EMoneyApp)(nil)
	_ servertypes.Application = (*EMoneyApp)(nil)
)

type EMoneyApp struct {
	*baseapp.BaseApp
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	database     db.DB
	currentBatch db.Batch

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*sdk.KVStoreKey
	tkeys   map[string]*sdk.TransientStoreKey
	memKeys map[string]*sdk.MemoryStoreKey

	// keepers
	accountKeeper    authkeeper.AccountKeeper
	bankKeeper       *embank.ProxyKeeper
	historykeeper    historykeeper.HistoryKeeper
	capabilityKeeper *capabilitykeeper.Keeper
	distrKeeper      distrkeeper.Keeper
	stakingKeeper    stakingkeeper.Keeper
	slashingKeeper   emslashing.Keeper
	crisisKeeper     crisiskeeper.Keeper
	upgradeKeeper    upgradekeeper.Keeper
	paramsKeeper     paramskeeper.Keeper
	ibcKeeper        *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	evidenceKeeper   evidencekeeper.Keeper
	transferKeeper   ibctransferkeeper.Keeper
	feeGrantKeeper   feegrantkeeper.Keeper
	authzKeeper      authzkeeper.Keeper

	// make scoped keepers public for test purposes
	scopedIBCKeeper      capabilitykeeper.ScopedKeeper
	scopedTransferKeeper capabilitykeeper.ScopedKeeper

	// custom modules
	inflationKeeper inflation.Keeper
	lpKeeper        liquidityprovider.Keeper
	issuerKeeper    issuer.Keeper
	authorityKeeper authority.Keeper
	marketKeeper    *market.Keeper
	buybackKeeper   buyback.Keeper

	// the module manager
	mm *module.Manager

	// simulation manager
	configurator module.Configurator
}

func (app *EMoneyApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

func (app *EMoneyApp) SimulationManager() *module.SimulationManager {
	panic("not supported")
}

type GenesisState map[string]json.RawMessage

// NewApp returns a reference to an initialized App.
func NewApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, skipUpgradeHeights map[int64]bool,
	homePath string, invCheckPeriod uint, encodingConfig EncodingConfig,
	appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp),
) *EMoneyApp {
	appCodec := encodingConfig.Marshaler
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(appName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		distrtypes.StoreKey, emslashing.StoreKey,
		paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey,
		lptypes.StoreKey, issuer.StoreKey, authority.StoreKey,
		market.StoreKey, market.StoreKeyIdx, buyback.StoreKey,
		inflation.StoreKey, feegrant.StoreKey, authzkeeper.StoreKey,
	)

	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &EMoneyApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
		database:          createApplicationDatabase(homePath),
	}

	app.paramsKeeper = initParamsKeeper(appCodec, legacyAmino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	bApp.SetParamStore(app.paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))

	// add capability keeper and ScopeToModule for ibc module
	app.capabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
	scopedIBCKeeper := app.capabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedTransferKeeper := app.capabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)

	// add keepers
	app.accountKeeper = authkeeper.NewAccountKeeper(
		appCodec, keys[authtypes.StoreKey], app.GetSubspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
	)

	app.bankKeeper = embank.Wrap(bankkeeper.NewBaseKeeper(
		appCodec, keys[banktypes.StoreKey], app.accountKeeper, app.GetSubspace(banktypes.ModuleName), app.ModuleAccountAddrs(),
	))

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec, keys[stakingtypes.StoreKey], app.accountKeeper, app.bankKeeper, app.GetSubspace(stakingtypes.ModuleName),
	)

	app.authzKeeper = authzkeeper.NewKeeper(keys[authzkeeper.StoreKey], appCodec, app.BaseApp.MsgServiceRouter())

	app.feeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], app.accountKeeper)

	app.historykeeper = historykeeper.NewHistoryKeeper(appCodec, keys[historykeeper.StoreKey], stakingKeeper, app.database)

	app.distrKeeper = distrkeeper.NewKeeper(
		appCodec, keys[distrtypes.StoreKey], app.GetSubspace(distrtypes.ModuleName), app.accountKeeper, app.bankKeeper,
		&stakingKeeper, authtypes.FeeCollectorName, app.ModuleAccountAddrs(),
	)
	app.slashingKeeper = emslashing.NewKeeper(
		appCodec, keys[emslashing.StoreKey], &stakingKeeper, app.GetSubspace(emslashing.ModuleName), app.bankKeeper,
		app.database, authtypes.FeeCollectorName,
	)
	app.crisisKeeper = crisiskeeper.NewKeeper(
		app.GetSubspace(crisistypes.ModuleName), invCheckPeriod, app.bankKeeper, authtypes.FeeCollectorName,
	)
	app.upgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath, app.BaseApp)

	app.registerUpgradeHandlers()

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.stakingKeeper = *stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks()),
	)

	// TODO is IBC relying on upgrade keeper's state?
	app.ibcKeeper = ibckeeper.NewKeeper(
		appCodec, keys[ibchost.StoreKey], app.GetSubspace(ibchost.ModuleName),
		app.historykeeper, app.upgradeKeeper, scopedIBCKeeper)

	// Create Transfer Keepers
	app.transferKeeper = ibctransferkeeper.NewKeeper(
		appCodec, keys[ibctransfertypes.StoreKey], app.GetSubspace(ibctransfertypes.ModuleName),
		app.ibcKeeper.ChannelKeeper, &app.ibcKeeper.PortKeeper,
		app.accountKeeper, app.bankKeeper, scopedTransferKeeper,
	)
	transferModule := transfer.NewAppModule(app.transferKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferModule)
	app.ibcKeeper.SetRouter(ibcRouter)

	// create evidence keeper with router
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, keys[evidencetypes.StoreKey], &app.stakingKeeper, app.slashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.evidenceKeeper = *evidenceKeeper

	app.inflationKeeper = inflation.NewKeeper(app.appCodec, keys[inflation.StoreKey], app.bankKeeper, app.accountKeeper, app.stakingKeeper, buyback.AccountName, authtypes.FeeCollectorName)
	app.lpKeeper = liquidityprovider.NewKeeper(app.appCodec, keys[lptypes.StoreKey], app.bankKeeper)
	app.issuerKeeper = issuer.NewKeeper(app.appCodec, keys[issuer.StoreKey], app.lpKeeper, app.inflationKeeper, app.bankKeeper)
	app.authorityKeeper = authority.NewKeeper(app.appCodec, keys[authority.StoreKey], app.issuerKeeper, app.bankKeeper, app, &app.upgradeKeeper)
	app.marketKeeper = market.NewKeeper(app.appCodec, keys[market.StoreKey], keys[market.StoreKeyIdx], app.accountKeeper, app.bankKeeper)
	app.buybackKeeper = buyback.NewKeeper(app.appCodec, keys[buyback.StoreKey], app.marketKeeper, app.accountKeeper, app.stakingKeeper, app.bankKeeper)

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	var skipGenesisInvariants = cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, app.accountKeeper, authsims.RandomGenesisAccounts),
		vesting.NewAppModule(app.accountKeeper, app.bankKeeper),
		bank.NewAppModule(appCodec, app.bankKeeper, app.accountKeeper),
		capability.NewAppModule(appCodec, *app.capabilityKeeper),
		feegrantmodule.NewAppModule(appCodec, app.accountKeeper, app.bankKeeper, app.feeGrantKeeper, app.interfaceRegistry),
		authzmodule.NewAppModule(appCodec, app.authzKeeper, app.accountKeeper, app.bankKeeper, interfaceRegistry),
		crisis.NewAppModule(&app.crisisKeeper, skipGenesisInvariants),
		emslashing.NewAppModule(appCodec, app.slashingKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		staking.NewAppModule(appCodec, app.stakingKeeper, app.accountKeeper, app.bankKeeper, app.historykeeper),
		upgrade.NewAppModule(app.upgradeKeeper),
		evidence.NewAppModule(app.evidenceKeeper),
		ibc.NewAppModule(app.ibcKeeper),
		params.NewAppModule(app.paramsKeeper),
		transferModule,
		emdistr.NewAppModule(distr.NewAppModule(appCodec, app.distrKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper), app.distrKeeper, app.accountKeeper, app.bankKeeper, app.database),
		liquidityprovider.NewAppModule(app.lpKeeper),
		issuer.NewAppModule(app.issuerKeeper),
		authority.NewAppModule(app.authorityKeeper),
		market.NewAppModule(app.marketKeeper),
		buyback.NewAppModule(app.buybackKeeper, app.bankKeeper),
		inflation.NewAppModule(app.inflationKeeper),
		queries.NewAppModule(app.accountKeeper, app.bankKeeper),
	)

	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		// todo (reviewer): check which modules make sense and which order
		upgradetypes.ModuleName,
		//// Cosmos #9800: capability module's begin blocker must come before any modules using capabilities (e.g. IBC)
		capabilitytypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		ibchost.ModuleName,

		authority.ModuleName,
		market.ModuleName,
		inflation.ModuleName,
		emslashing.ModuleName,
		emdistr.ModuleName,
		buyback.ModuleName,
		//bep3.ModuleName, // <- TODO Forces app-state change in BeginBlock
	)

	app.mm.SetOrderEndBlockers(crisistypes.ModuleName, stakingtypes.ModuleName,
		feegrant.ModuleName, authz.ModuleName)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName, authtypes.ModuleName, banktypes.ModuleName, emdistr.ModuleName, stakingtypes.ModuleName,
		emslashing.ModuleName, crisistypes.ModuleName,
		ibchost.ModuleName, genutiltypes.ModuleName, evidencetypes.ModuleName, ibctransfertypes.ModuleName,
		inflation.ModuleName, issuer.ModuleName, authority.ModuleName, market.ModuleName, buyback.ModuleName,
		liquidityprovider.ModuleName, feegrant.ModuleName, authz.ModuleName,
	)

	app.mm.RegisterInvariants(&app.crisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	// TODO hack to initialize application
	// from sdk.bank/module.go: AppModule RegisterServices
	// 	m := keeper.NewMigrator(am.keeper.(keeper.BaseKeeper))
	sdkBk := app.bankKeeper.GetBankKeeper()
	bankMigModule := bank.NewAppModule(appCodec, *sdkBk, app.accountKeeper)
	proxy := app.mm.Modules[banktypes.ModuleName]
	app.mm.Modules[banktypes.ModuleName] = bankMigModule
	app.mm.RegisterServices(app.configurator)

	app.mm.Modules[banktypes.ModuleName] = proxy

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	anteHandler, err := ante.NewAnteHandler(
		ante.EmAnteHandlerOptions{
			AccountKeeper:    app.accountKeeper,
			BankKeeper:       app.bankKeeper,
			FeegrantKeeper:   app.feeGrantKeeper,
			SignModeHandler:  encodingConfig.TxConfig.SignModeHandler(),
			SigGasConsumer:   sdkante.DefaultSigVerificationGasConsumer,
			StakingKeeper:    app.stakingKeeper,
			IBCChannelkeeper: channelkeeper.Keeper{},
		},
	)
	if err != nil {
		panic(fmt.Errorf("failed to create AnteHandler: %s", err))
	}

	app.SetAnteHandler(anteHandler)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(fmt.Sprintf("failed to load latest version: %s", err))
		}
	}

	app.scopedIBCKeeper = scopedIBCKeeper
	app.scopedTransferKeeper = scopedTransferKeeper

	return app
}

func (app *EMoneyApp) registerUpgradeHandlers() {
	const upg44Plan = "v44-upg-test"

	fmt.Println("")
	fmt.Println("*** ------------------------------------------------------- ")

	fmt.Println("Entered registerUpgradeHandlers")

	fmt.Println("*** ------------------------------------------------------- ")
	fmt.Println("")

	app.upgradeKeeper.SetUpgradeHandler(
		upg44Plan,
		func(ctx sdk.Context, _ upgradetypes.Plan, _ module.VersionMap) (module.VersionMap, error) {
			app.ibcKeeper.ConnectionKeeper.SetParams(ctx, ibcconnectiontypes.DefaultParams())

			fromVM := make(map[string]uint64)
			for _, mod := range app.mm.Modules {
				fromVM[mod.Name()] = mod.ConsensusVersion()
			}
			// override versions for _new_ modules as to not skip InitGenesis
			fromVM[authz.ModuleName] = 0
			fromVM[feegrant.ModuleName] = 0

			ctx.Logger().Info("Upgraded to " + upg44Plan)

			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		},
	)

	fmt.Println("")
	fmt.Println("*** ------------------------------------------------------- ")

	fmt.Println("after SetUpgradeHandler", upg44Plan, "has handler:", app.upgradeKeeper.HasHandler(upg44Plan))

	fmt.Println("*** ------------------------------------------------------- ")
	fmt.Println("")

	upgradeInfo, err := app.upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if upgradeInfo.Name == upg44Plan && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		fmt.Println("")
		fmt.Println("*** ------------------------------------------------------- ")

		fmt.Println("entered setStoreLoader check", upg44Plan, "has handler:", app.upgradeKeeper.HasHandler(upg44Plan))

		fmt.Println("*** ------------------------------------------------------- ")
		fmt.Println("")

		storeUpgrades := store.StoreUpgrades{
			Added: []string{authz.ModuleName, feegrant.ModuleName},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	fmt.Println("")
	fmt.Println("*** ------------------------------------------------------- ")

	fmt.Println("after SetStoreLoader", upg44Plan, "has handler:", app.upgradeKeeper.HasHandler(upg44Plan))

	fmt.Println("*** ------------------------------------------------------- ")
	fmt.Println("")
}

func createApplicationDatabase(rootDir string) db.DB {
	datadirectory := filepath.Join(rootDir, "data")
	emoneydb, err := db.NewGoLevelDB("emoney", datadirectory)
	if err != nil {
		panic(err)
	}

	return emoneydb
}

// load a particular height
func (app *EMoneyApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

func (app *EMoneyApp) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "emz")
}

func (app *EMoneyApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	// Due to how the current Tendermint implementation calculates blocktime, a byzantine 1/3 of voting power can move time forward an arbitrary amount.
	// Have non-malicious nodes shut down if this appears to be happening.
	// This will effectively halt the chain and require off-chain coordination to remedy.
	walltime := time.Now().UTC()
	if walltime.Add(time.Hour).Before(ctx.BlockTime()) {
		s := fmt.Sprintf("Blocktime %v is too far ahead of local wall clock %v.\nSuspending node without processing block %v.\n", ctx.BlockTime(), walltime, ctx.BlockHeight())
		fmt.Println(s)
		panic(s)
	}

	app.currentBatch = app.database.NewBatch() // store in app state as ctx is different in end block
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	ctx = apptypes.WithCurrentBatch(ctx, app.currentBatch)

	return app.mm.BeginBlock(ctx, req)
}

// application updates every end block
func (app *EMoneyApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	block := ctx.BlockHeader()
	proposerAddress := block.GetProposerAddress()
	app.Logger(ctx).Info(fmt.Sprintf("Endblock: Block %v was proposed by %v", ctx.BlockHeight(), sdk.ValAddress(proposerAddress)))

	response := app.mm.EndBlock(ctx, req)
	err := app.currentBatch.Write() // Write non-IAVL state to database
	if err != nil {                 // todo (reviewer): should we panic or ignore? panics are not handled downstream will cause a crash
		panic(err)
	}
	return response
}

// InitChainer application update at chain initialization
func (app *EMoneyApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) (res abci.ResponseInitChain) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	app.upgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())

	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

func (app *EMoneyApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// AppCodec returns SimApp's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *EMoneyApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns SimApp's InterfaceRegistry
func (app *EMoneyApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *EMoneyApp) GetKey(storeKey string) *sdk.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *EMoneyApp) GetTKey(storeKey string) *sdk.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *EMoneyApp) GetMemKey(storeKey string) *sdk.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *EMoneyApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.paramsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *EMoneyApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	rpc.RegisterRoutes(clientCtx, apiSvr.Router)
	// Register legacy tx routes.
	authrest.RegisterTxRoutes(clientCtx, apiSvr.Router)
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	ModuleBasics.RegisterRESTRoutes(clientCtx, apiSvr.Router)
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(clientCtx, apiSvr.Router)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *EMoneyApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *EMoneyApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.interfaceRegistry)
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(ctx client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// GetMaccs returns a copy of the module accounts
func GetMaccs() map[string]bool {
	maccs := make(map[string]bool)
	for k := range maccPerms {
		maccs[k] = true
	}
	return maccs
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey sdk.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(emslashing.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)

	return paramsKeeper
}

func (app EMoneyApp) SetMinimumGasPrices(gasPricesStr string) (err error) {
	if _, err = sdk.ParseDecCoins(gasPricesStr); err != nil {
		return
	}

	baseapp.SetMinGasPrices(gasPricesStr)(app.BaseApp)
	return
}

func init() {
	sdk.DefaultPowerReduction = sdk.OneInt()
}
