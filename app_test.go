package emoney

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibckeeper "github.com/cosmos/ibc-go/modules/core/keeper"
	ibctesting "github.com/cosmos/ibc-go/testing"
	apptypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/authority"
	authtypes "github.com/e-money/em-ledger/x/authority/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

var configOnce sync.Once

func mustGetAccAddress(addr string) sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		panic(err)
	}
	return acc
}

func TestSimAppExportAndBlockedAddrs(t *testing.T) {
	encCfg, db, app, _ := getEmSimApp(t, mustGetAccAddress("cosmos1lagqmceycrfpkyu7y6ayrk6jyvru5mkrkp8vkn"))
	app.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true,
		map[int64]bool{}, t.TempDir(), 0, encCfg, EmptyAppOptions{},
	)
	_, err := app2.ExportAppStateAndValidators(false, []string{})
	require.NoError(
		t, err, "ExportAppStateAndValidators should not have an error",
	)
}

func getEmSimApp(
	t *testing.T, authorityAcc sdk.AccAddress,
) (encCfg EncodingConfig, memDB *dbm.MemDB, eMoneyApp *EMoneyApp, homeFolder string) {
	t.Helper()

	encCfg = MakeEncodingConfig()
	db := dbm.NewMemDB()
	homeDir := t.TempDir()
	t.Log("home dir:", homeDir)

	app := NewApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true,
		map[int64]bool{}, homeDir, 0, encCfg, EmptyAppOptions{},
	)

	for acc := range maccPerms {
		require.True(
			t,
			app.bankKeeper.BlockedAddr(app.accountKeeper.GetModuleAddress(acc)),
			"ensure that blocked addresses are properly set in bank keeper",
		)
	}

	genesisState := ModuleBasics.DefaultGenesis(encCfg.Marshaler)
	authorityState := authtypes.GenesisState{AuthorityKey: authorityAcc.String(), MinGasPrices: sdk.NewDecCoins()}

	genesisState["authority"] = encCfg.Marshaler.MustMarshalJSON(&authorityState)

	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	app.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)

	return encCfg, db, app, homeFolder
}

func init() {
	ibctesting.DefaultTestingAppInit = getIBCApp
}

type emIBCApp struct {
	*EMoneyApp
	encCfg EncodingConfig
}

var tempDirs = make([]string, 0, 2)

func (app emIBCApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

func (app emIBCApp) GetStakingKeeper() stakingkeeper.Keeper {
	return app.stakingKeeper
}

func (app emIBCApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.ibcKeeper
}

func (app emIBCApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.scopedIBCKeeper
}

func (app emIBCApp) GetTxConfig() client.TxConfig {
	return app.encCfg.TxConfig
}

func (app emIBCApp) AppCodec() codec.Codec {
	return app.appCodec
}

func createIBCApp() (emIBCApp emIBCApp, genesis map[string]json.RawMessage) {
	var _ ibctesting.TestingApp = emIBCApp

	//configOnce.Do(apptypes.ConfigureSDK)
	authorityAcc := mustGetAccAddress("cosmos1lagqmceycrfpkyu7y6ayrk6jyvru5mkrkp8vkn")

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	emIBCApp.encCfg = MakeEncodingConfig()
	db := dbm.NewMemDB()
	homeDir, err := os.MkdirTemp("/tmp", "chain-")
	if err != nil {
		panic(err)
	}
	tempDirs = append(tempDirs, homeDir)
	logger.Info(fmt.Sprintf("home dir:%s", homeDir))

	emIBCApp.EMoneyApp = NewApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true,
		map[int64]bool{}, homeDir, 5, emIBCApp.encCfg, EmptyAppOptions{},
	)

	genesisState := ModuleBasics.DefaultGenesis(emIBCApp.encCfg.Marshaler)
	authorityState := authtypes.GenesisState{AuthorityKey: authorityAcc.String(), MinGasPrices: sdk.NewDecCoins()}

	genesisState["authority"] = emIBCApp.encCfg.Marshaler.MustMarshalJSON(&authorityState)

	return emIBCApp, genesisState
}

// Bridge the concrete type returning function with the interface returning func
func getIBCApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	return createIBCApp()
}

// IBCTestSuite is a testing suite to test keeper functions.
type IBCTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

// SetupTest creates a coordinator with 2 test chains.
func (suite *IBCTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)

	suite.Nil(suite.chainA)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.NotNil(suite.chainA)

	suite.Nil(suite.chainB)
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.NotNil(suite.chainB)
}

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort

	return path
}

func (suite *IBCTestSuite) Test_IBCTransfer() {
	// setup between chainA and chainB
	path := NewTransferPath(suite.chainA, suite.chainB)

	// Succeeds setting the clients
	// suite.coordinator.SetupClients(path)

	// Fails here
	suite.coordinator.Setup(path)

	suite.Require().Equal("07-tendermint-0", path.EndpointA.ClientID)
}

func Test_IBCSuite(t *testing.T) {
	suite.Run(t, new(IBCTestSuite))
}

func (s *IBCTestSuite) TearDownSuite() {
	s.T().Log("tearing down ibc test suite")
	for _, t := range tempDirs {
		err := os.RemoveAll(t)
		if err != nil {
			s.T().Log(fmt.Sprintf("removing %s temp dir %v", t, err))
		}
	}
}


// EmptyAppOptions is a stub implementing AppOptions
type EmptyAppOptions struct{}

// Get implements AppOptions
func (ao EmptyAppOptions) Get(o string) interface{} {
	return nil
}

type emAppTests struct {
	ctx       sdk.Context
	homeDir   string
	app       *EMoneyApp
	authority sdk.AccAddress
}

func (et emAppTests) initEmApp(t *testing.T) emAppTests {
	t.Helper()

	var err error
	et.authority, err = sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
	require.NoError(t, err)

	_, _, app, homeDir := getEmSimApp(t, et.authority)

	et.homeDir = homeDir
	et.app = app
	et.ctx = app.BaseApp.NewContext(false, tmproto.Header{
		Time:   time.Now(),
		Height: 10,
	})

	return et
}

func Test_Upgrade(t *testing.T) {
	configOnce.Do(apptypes.ConfigureSDK)

	tests := []struct {
		suite        emAppTests
		name         string
		plan         upgradetypes.Plan
		setupUpgCond func(simApp emAppTests, plan *upgradetypes.Plan)
		expSchedPass bool
		qrySched     bool
		qryApply     bool
	}{
		{
			name: "successful time schedule",
			plan: upgradetypes.Plan{
				Name: "all-good",
				Info: "some text here",
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {
				plan.Time = simApp.ctx.BlockTime().Add(time.Hour)
			},
			// As of v44, time scheduled upgrades are deprecated
			expSchedPass: false,
		},
		{
			name: "successful height schedule",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {},
			expSchedPass: true,
		},
		{
			name: "setting both time and height schedule",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {
				plan.Time = simApp.ctx.BlockTime().Add(time.Hour)
			},
			expSchedPass: false,
		},
		{
			name: "successful overwrite",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {

				_, err := simApp.app.authorityKeeper.ScheduleUpgrade(
					simApp.ctx, simApp.authority, upgradetypes.Plan{
						Name:   "alt-good",
						Info:   "new text here",
						Height: 543210000,
					},
				)
				require.NoError(t, err)
			},
			expSchedPass: true,
		},
		{
			name: "successful overwrite future with earlier date",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {
				_, err := simApp.app.authorityKeeper.ScheduleUpgrade(
					simApp.ctx, simApp.authority, upgradetypes.Plan{
						Name:   "alt-good",
						Info:   "new text here",
						Height: 543210000,
					},
				)
				require.NoError(t, err)
			},
			expSchedPass: true,
		},
		{
			name: "successful overwrite earlier with future date",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 543210000,
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {
				_, err := simApp.app.authorityKeeper.ScheduleUpgrade(
					simApp.ctx, simApp.authority, upgradetypes.Plan{
						Name:   "alt-good",
						Info:   "new text here",
						Height: 123450000,
					},
				)
				require.NoError(t, err)
			},
			expSchedPass: true,
		},
		{
			name: "unsuccessful schedule: missing plan name",
			plan: upgradetypes.Plan{
				Height: 123450000,
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {},
			expSchedPass: false,
		},
		{
			name: "unsuccessful time schedule: initialized, uninitialized due date in past",
			plan: upgradetypes.Plan{
				Name: "all-good",
				Info: "some text here",
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {
				plan.Time = simApp.ctx.BlockTime()
			},
			expSchedPass: false,
		},
		{
			name: "unsuccessful height schedule: due date in past",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 1,
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {},
			expSchedPass: false,
		},
		{
			name: "unsuccessful schedule: schedule already executed",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simEmApp emAppTests, plan *upgradetypes.Plan) {
				simEmApp.app.upgradeKeeper.SetUpgradeHandler("all-good", func(_ sdk.Context, upgradeHandler upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
					return vm, nil
				})

				simEmApp.app.upgradeKeeper.ApplyUpgrade(
					simEmApp.ctx, upgradetypes.Plan{
						Name:   "all-good",
						Info:   "some text here",
						Height: 123450000,
					},
				)
			},
			expSchedPass: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.suite = emAppTests{}.initEmApp(t)

				// setup test case
				tt.setupUpgCond(tt.suite, &tt.plan)

				// schedule upgrade plan
				var err error
				_, err = tt.suite.app.authorityKeeper.ScheduleUpgrade(
					tt.suite.ctx, tt.suite.authority, tt.plan,
				)
				schedPlan, hasPlan := tt.suite.app.authorityKeeper.GetUpgradePlan(tt.suite.ctx)

				// validate plan side effect
				if tt.expSchedPass {
					require.NoError(t, err, "Upgrade() error = %v, expSchedPass %v", err, tt.expSchedPass)
					require.Truef(t, hasPlan, "hasPlan: %t there should be a plan", hasPlan)
					require.Equalf(t, schedPlan, tt.plan, "queried %v != %v", schedPlan, tt.plan)
				} else {
					require.Falsef(t, hasPlan, "hasPlan: %t plan should not exist", hasPlan)
					require.NotEqualf(t, schedPlan, tt.plan, "queried %v == %v", schedPlan, tt.plan)
				}

				// apply and confirm plan deletion
				if tt.expSchedPass {
					executePlan(
						tt.suite.ctx, t, tt.suite.app.upgradeKeeper,
						tt.suite.app.authorityKeeper, tt.plan,
					)
				}
			},
		)
	}
}

func executePlan(
	ctx sdk.Context, t *testing.T, uk upgradekeeper.Keeper, ak authority.Keeper,
	plan upgradetypes.Plan,
) {
	uk.SetUpgradeHandler(plan.Name, func(_ sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return vm, nil
	})

	uk.ApplyUpgrade(ctx, plan)

	schedPlan, hasPlan := ak.GetUpgradePlan(ctx)
	require.Falsef(t, hasPlan, "hasPlan: %t plan should not exist", hasPlan)
	require.NotEqualf(t, schedPlan, plan, "queried %v == %v", schedPlan, plan)
}
