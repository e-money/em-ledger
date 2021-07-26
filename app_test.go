package emoney

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"

	apptypes "github.com/e-money/em-ledger/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	authtypes "github.com/e-money/em-ledger/x/authority/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func TestSimAppExportAndBlockedAddrs(t *testing.T) {
	encCfg, db, app := getEmSimApp(t, rand.Bytes(sdk.AddrLen))
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

func getEmSimApp(t *testing.T, authorityAcc sdk.AccAddress) (EncodingConfig, *dbm.MemDB, *EMoneyApp) {
	encCfg := MakeEncodingConfig()
	db := dbm.NewMemDB()
	app := NewApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true,
		map[int64]bool{}, t.TempDir(), 0, encCfg, EmptyAppOptions{},
	)

	for acc := range maccPerms {
		require.True(
			t,
			app.bankKeeper.BlockedAddr(app.accountKeeper.GetModuleAddress(acc)),
			"ensure that blocked addresses are properly set in bank keeper",
		)
	}

	genesisState := ModuleBasics.DefaultGenesis(encCfg.Marshaler)
	//authorityState := authority.NewGenesisState(rand.Bytes(sdk.AddrLen), sdk.NewDecCoins())
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

	return encCfg, db, app
}

// EmptyAppOptions is a stub implementing AppOptions
type EmptyAppOptions struct{}

// Get implements AppOptions
func (ao EmptyAppOptions) Get(o string) interface{} {
	return nil
}

type emAppTests struct {
	ctx     sdk.Context
	homeDir string
	app     *EMoneyApp
}

func (et *emAppTests) initEmApp(t *testing.T) {
	homeDir := filepath.Join(t.TempDir(), "x_upgrade_keeper_test")

	authorityAddr, err := sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
	require.NoError(t, err)

	_, _, app := getEmSimApp(t, authorityAddr)
	app.upgradeKeeper = upgradekeeper.NewKeeper( // recreate keeper in order to use a custom home path
		make(map[int64]bool), app.GetKey(upgradetypes.StoreKey), app.AppCodec(), homeDir,
	)
	t.Log("home dir:", homeDir)

	et.homeDir = homeDir
	et.app = app
	et.ctx = app.BaseApp.NewContext(false, tmproto.Header{
		Time:   time.Now(),
		Height: 10,
	})
}

func Test_Upgrade(t *testing.T) {
	apptypes.ConfigureSDK()

	tests := []struct {
		suite        emAppTests
		name         string
		plan         upgradetypes.Plan
		setupUpgCond func(simApp emAppTests, plan *upgradetypes.Plan)
		expPass      bool
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
			expPass: true,
		},
		{
			name: "successful height schedule",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {},
			expPass:      true,
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
			expPass: false,
		},
		{
			name: "successful overwrite",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {

				_, err := simApp.app.authorityKeeper.ScheduleUpgrade(simApp.ctx, upgradetypes.Plan{
					Name:   "alt-good",
					Info:   "new text here",
					Height: 543210000,
				})
				require.NoError(t, err)
			},
			expPass: true,
		},
		{
			name: "successful overwrite future with earlier date",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {
				_, err := simApp.app.authorityKeeper.ScheduleUpgrade(simApp.ctx, upgradetypes.Plan{
					Name:   "alt-good",
					Info:   "new text here",
					Height: 543210000,
				})
				require.NoError(t, err)
			},
			expPass: true,
		},
		{
			name: "successful overwrite earlier with future date",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 543210000,
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {
				_, err := simApp.app.authorityKeeper.ScheduleUpgrade(simApp.ctx, upgradetypes.Plan{
					Name:   "alt-good",
					Info:   "new text here",
					Height: 123450000,
				})
				require.NoError(t, err)
			},
			expPass: true,
		},
		{
			name: "unsuccessful schedule: missing plan name",
			plan: upgradetypes.Plan{
				Height: 123450000,
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {},
			expPass:      false,
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
			expPass: false,
		},
		{
			name: "unsuccessful height schedule: due date in past",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 1,
			},
			setupUpgCond: func(simApp emAppTests, plan *upgradetypes.Plan) {},
			expPass:      false,
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
				_, err := simEmApp.app.authorityKeeper.ApplyUpgrade(
					simEmApp.ctx, upgradetypes.Plan{
						Name:   "all-good",
						Info:   "some text here",
						Height: 123450000,
					})
				if err != nil {
					panic(err)
				}
			},
			expPass: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.suite = emAppTests{}
				tt.suite.initEmApp(t)

				// setup test case
				tt.setupUpgCond(tt.suite, &tt.plan)

				// schedule upgrade plan
				var err error
				_, err = tt.suite.app.authorityKeeper.ScheduleUpgrade(tt.suite.ctx, tt.plan)
				schedPlan, hasPlan := tt.suite.app.authorityKeeper.GetUpgradePlan(tt.suite.ctx)

				// validate plan side-effect
				if tt.expPass {
					require.Truef(t, hasPlan, "hasPlan: %t there should be a plan", hasPlan)
					require.Equalf(t, schedPlan, tt.plan, "queried %v != %v", schedPlan, tt.plan)
				} else {
					require.Falsef(t, hasPlan, "hasPlan: %t plan should not exist", hasPlan)
					require.NotEqualf(t, schedPlan, tt.plan, "queried %v == %v", schedPlan, tt.plan)
				}

				// apply and confirm plan deletion
				if err == nil {
					tt.suite.app.upgradeKeeper.SetUpgradeHandler(tt.plan.Name, func(_ sdk.Context, _ upgradetypes.Plan) {})
					_, err = tt.suite.app.authorityKeeper.ApplyUpgrade(tt.suite.ctx, tt.plan)
					schedPlan, hasPlan = tt.suite.app.authorityKeeper.GetUpgradePlan(tt.suite.ctx)
					require.Falsef(t, hasPlan, "hasPlan: %t plan should not exist", hasPlan)
					require.NotEqualf(t, schedPlan, tt.plan, "queried %v == %v", schedPlan, tt.plan)
				}

				if (err != nil) == tt.expPass {
					t.Errorf(
						"Upgrade() error = %v, expPass %v", err, tt.expPass,
					)
					return
				}
			},
		)
	}
}
