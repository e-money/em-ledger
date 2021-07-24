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

type upgTests struct {
	ctx     sdk.Context
	homeDir string
	app     *EMoneyApp
}

func (ut *upgTests) initApp(t *testing.T) {
	homeDir := filepath.Join(t.TempDir(), "x_upgrade_keeper_test")

	authorityAddr, err := sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
	require.NoError(t, err)

	_, _, app := getEmSimApp(t, authorityAddr)
	app.upgradeKeeper = upgradekeeper.NewKeeper( // recreate keeper in order to use a custom home path
		make(map[int64]bool), app.GetKey(upgradetypes.StoreKey), app.AppCodec(), homeDir,
	)
	t.Log("home dir:", homeDir)

	ut.homeDir = homeDir
	ut.app = app
	ut.ctx = app.BaseApp.NewContext(false, tmproto.Header{
		Time:   time.Now(),
		Height: 10,
	})
}

func Test_ScheduleUpgrade(t *testing.T) {
	apptypes.ConfigureSDK()

	tests := []struct {
		suite        upgTests
		name         string
		plan         upgradetypes.Plan
		setupUpgCond func(simApp upgTests, plan *upgradetypes.Plan)
		expPass      bool
	}{
		{
			suite: upgTests{},
			name:  "successful time schedule",
			plan: upgradetypes.Plan{
				Name: "all-good",
				Info: "some text here",
			},
			setupUpgCond: func(simApp upgTests, plan *upgradetypes.Plan) {
				plan.Time = simApp.ctx.BlockTime().Add(time.Hour)
			},
			expPass: true,
		},
		{
			suite: upgTests{},
			name:  "successful height schedule",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simApp upgTests, plan *upgradetypes.Plan) {},
			expPass:      true,
		},
		{
			suite: upgTests{},
			name:  "successful schedule",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simApp upgTests, plan *upgradetypes.Plan) {},
			expPass:      true,
		},
		{
			suite: upgTests{},
			name:  "successful overwrite",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simApp upgTests, plan *upgradetypes.Plan) {

				err := simApp.app.upgradeKeeper.ScheduleUpgrade(simApp.ctx, upgradetypes.Plan{
					Name:   "alt-good",
					Info:   "new text here",
					Height: 543210000,
				})
				require.NoError(t, err)
			},
			expPass: true,
		},
		{
			suite: upgTests{},
			name:  "successful overwrite",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simApp upgTests, plan *upgradetypes.Plan) {
				err := simApp.app.upgradeKeeper.ScheduleUpgrade(simApp.ctx, upgradetypes.Plan{
					Name:   "alt-good",
					Info:   "new text here",
					Height: 543210000,
				})
				require.NoError(t, err)
			},
			expPass: true,
		},
		{
			suite: upgTests{},
			name:  "successful IBC overwrite with non IBC plan",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simApp upgTests, plan *upgradetypes.Plan) {
				err := simApp.app.upgradeKeeper.ScheduleUpgrade(simApp.ctx, upgradetypes.Plan{
					Name:   "alt-good",
					Info:   "new text here",
					Height: 543210000,
				})
				require.NoError(t, err)
			},
			expPass: true,
		},
		{
			suite: upgTests{},
			name:  "unsuccessful schedule: invalid plan",
			plan: upgradetypes.Plan{
				Height: 123450000,
			},
			setupUpgCond: func(simApp upgTests, plan *upgradetypes.Plan) {},
			expPass:      false,
		},
		{
			suite: upgTests{},
			name:  "unsuccessful time schedule: due date in past",
			plan: upgradetypes.Plan{
				Name: "all-good",
				Info: "some text here",
			},
			setupUpgCond: func(simApp upgTests, plan *upgradetypes.Plan) {
				plan.Time = simApp.ctx.BlockTime()
			},
			expPass: false,
		},
		{
			suite: upgTests{},
			name:  "unsuccessful height schedule: due date in past",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 1,
			},
			setupUpgCond: func(simApp upgTests, plan *upgradetypes.Plan) {},
			expPass:      false,
		},
		{
			suite: upgTests{},
			name:  "unsuccessful schedule: schedule already executed",
			plan: upgradetypes.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setupUpgCond: func(simApp upgTests, plan *upgradetypes.Plan) {
				simApp.app.upgradeKeeper.SetUpgradeHandler("all-good", func(_ sdk.Context, _ upgradetypes.Plan) {})
				simApp.app.upgradeKeeper.ApplyUpgrade(simApp.ctx, upgradetypes.Plan{
					Name:   "all-good",
					Info:   "some text here",
					Height: 123450000,
				})
			},
			expPass: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.suite.initApp(t)

				// setupUpgCond test case
				tt.setupUpgCond(tt.suite, &tt.plan)

				err := tt.suite.app.upgradeKeeper.ScheduleUpgrade(tt.suite.ctx, tt.plan)

				if (err != nil) == tt.expPass {
					t.Errorf(
						"scheduleUpgrade() error = %v, expPass %v", err, tt.expPass,
					)
					return
				}
			},
		)
	}
}
