package emoney

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/baseapp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	apptypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/authority"
	authtypes "github.com/e-money/em-ledger/x/authority/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

var configOnce sync.Once

func TestSimAppExportAndBlockedAddrs(t *testing.T) {
	encCfg, db, app, _ := getEmSimApp(t, rand.Bytes(sdk.AddrLen))
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
			ConsensusParams: &abci.ConsensusParams{
				Block: &abci.BlockParams{
					MaxGas: 100,
				},
			},
		},
	)

	return encCfg, db, app, homeFolder
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
			expSchedPass: true,
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
				simEmApp.app.upgradeKeeper.SetUpgradeHandler("all-good", func(_ sdk.Context, _ upgradetypes.Plan) {})
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

func Test_UpgradeByTime(t *testing.T) {

	configOnce.Do(apptypes.ConfigureSDK)

	tests := []struct {
		suite        emAppTests
		name         string
		plan         upgradetypes.Plan
		blockTimes   []time.Time
		setupUpgCond func(simApp emAppTests, blockTimes []time.Time, plan *upgradetypes.Plan)
		expPass      bool
		qrySched     bool
		qryApply     bool
	}{
		{
			name: "successful time plan + 1minute",
			plan: upgradetypes.Plan{
				Name: "all-good",
				Info: "some text here",
			},
			setupUpgCond: func(simApp emAppTests, blockTimes []time.Time, plan *upgradetypes.Plan) {
				const duration = 10 * time.Minute
				plan.Time = simApp.ctx.BlockTime().Add(duration)
				blockTimes = []time.Time{
					simApp.ctx.BlockTime(),
					simApp.ctx.BlockTime().Add(time.Minute),
					plan.Time.Add(-time.Nanosecond),
				}
			},
			expPass: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.suite = emAppTests{}.initEmApp(t)

				// setup test case
				tt.setupUpgCond(tt.suite, tt.blockTimes, &tt.plan)

				// schedule upgrade plan
				var err error
				_, err = tt.suite.app.authorityKeeper.ScheduleUpgrade(
					tt.suite.ctx, tt.suite.authority, tt.plan,
				)
				schedPlan, hasPlan := tt.suite.app.authorityKeeper.GetUpgradePlan(tt.suite.ctx)

				// validate plan side effect
				if tt.expPass {
					require.NoError(t, err, "Upgrade() error = %v, expSchedPass %v", err, tt.expPass)
					require.Truef(
						t, hasPlan, "hasPlan: %t there should be a plan",
						hasPlan,
					)
					require.Equalf(
						t, schedPlan, tt.plan, "queried %v != %v", schedPlan,
						tt.plan,
					)
				} else {
					require.Falsef(
						t, hasPlan, "hasPlan: %t plan should not exist", hasPlan,
					)
					require.NotEqualf(
						t, schedPlan, tt.plan, "queried %v == %v", schedPlan,
						tt.plan,
					)
				}

				// advance chain and confirm plan execution
				tt.suite.app.authorityKeeper.GetUpgradePlan(tt.suite.ctx)
				for _, blockTime := range tt.blockTimes {
					tt.suite.ctx = tt.suite.ctx.WithBlockTime(blockTime)
					require.Falsef(t, schedPlan.ShouldExecute(tt.suite.ctx), "premature timing for executing plan")
				}
				// move forward to execution time
				tt.suite.ctx = tt.suite.ctx.WithBlockTime(schedPlan.Time)
				require.Truef(t, schedPlan.ShouldExecute(tt.suite.ctx), "plan should be ripe for execution")

				executePlan(
					tt.suite.ctx, t, tt.suite.app.upgradeKeeper,
					tt.suite.app.authorityKeeper, tt.plan,
				)
			},
		)
	}
}

func executePlan(
	ctx sdk.Context, t *testing.T, uk upgradekeeper.Keeper, ak authority.Keeper,
	plan upgradetypes.Plan,
) {
	uk.SetUpgradeHandler(plan.Name, func(_ sdk.Context, _ upgradetypes.Plan) {})

	uk.ApplyUpgrade(ctx, plan)

	schedPlan, hasPlan := ak.GetUpgradePlan(ctx)
	require.Falsef(t, hasPlan, "hasPlan: %t plan should not exist", hasPlan)
	require.NotEqualf(t, schedPlan, plan, "queried %v == %v", schedPlan, plan)
}

func TestUpdatingChainParams(t *testing.T) {
	configOnce.Do(apptypes.ConfigureSDK)

	et := emAppTests{}.initEmApp(t)
	header := tmproto.Header{Height: et.app.LastBlockHeight() + 1}
	et.app.BeginBlock(abci.RequestBeginBlock{Header: header})
	et.app.EndBlock(abci.RequestEndBlock{})
	et.app.Commit()

	// block --> 2
	header = tmproto.Header{Height: et.app.LastBlockHeight() + 1}
	et.app.BeginBlock(abci.RequestBeginBlock{Header: header})
	et.ctx = et.app.BaseApp.NewContext(true, header)

	paramChanges := []proposal.ParamChange{
		{
			Subspace: stakingtypes.ModuleName,
			Key:      "MaxValidators",
			Value:    "101",
		},
	}
	_, err := et.app.authorityKeeper.SetParams(et.ctx, et.authority, paramChanges)
	require.NoError(t, err)

	stateMaxValidators := et.app.stakingKeeper.MaxValidators(et.ctx)
	maxValidators := et.app.stakingKeeper.GetParams(et.ctx).MaxValidators
	require.Equal(t, uint32(101), stateMaxValidators)
	require.Equal(t, stateMaxValidators, maxValidators)

	et.app.EndBlock(abci.RequestEndBlock{})
	et.app.Commit()

	stateMaxValidators = et.app.stakingKeeper.MaxValidators(et.ctx)
	require.Equal(t, uint32(101), stateMaxValidators)
}

func TestUpdatingBlockParams(t *testing.T) {
	configOnce.Do(apptypes.ConfigureSDK)

	et := emAppTests{}.initEmApp(t)
	header := tmproto.Header{Height: et.app.LastBlockHeight() + 1}
	et.app.BeginBlock(abci.RequestBeginBlock{Header: header})
	et.app.EndBlock(abci.RequestEndBlock{})
	et.app.Commit()

	// block --> 2
	header = tmproto.Header{Height: et.app.LastBlockHeight() + 1}
	et.app.BeginBlock(abci.RequestBeginBlock{Header: header})
	et.ctx = et.app.BaseApp.NewContext(true, header)

	consensusParams := et.app.BaseApp.GetConsensusParams(et.ctx)
	blockParams := *consensusParams.Block

	subSpc := et.app.GetSubspace(baseapp.Paramspace)
	var subSpaceBlockParams abci.BlockParams
	subSpc.Get(et.ctx, baseapp.ParamStoreKeyBlockParams, &subSpaceBlockParams)

	require.Equal(t, blockParams.String(), subSpaceBlockParams.String())

	cdc := codec.NewLegacyAmino()
	blockParams.MaxBytes = int64(1024)
	bz, err := cdc.MarshalJSON(blockParams)
	require.NoError(t, err)
	strBlockParams := string(bz)
	fmt.Println(strBlockParams)

	paramChanges := []proposal.ParamChange{
		{
			Subspace: baseapp.Paramspace,
			Key:      "BlockParams",
			Value:    strBlockParams,
		},
	}
	_, err = et.app.authorityKeeper.SetParams(et.ctx, et.authority, paramChanges)
	require.NoError(t, err)

	et.app.EndBlock(abci.RequestEndBlock{})
	et.app.Commit()

	consensusParams = et.app.BaseApp.GetConsensusParams(et.ctx)
	require.Equal(t, consensusParams.Block.String(), blockParams.String())
}
