package v040

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	tmtypes "github.com/tendermint/tendermint/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app *simapp.SimApp
	ctx sdk.Context
}

func (suite *IntegrationTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())

	suite.app = app
	suite.ctx = ctx
}

func (suite *IntegrationTestSuite) TestKeeperSetGetDenomMetaData() {
	app, ctx := suite.app, suite.ctx

	inputDenomData := GetDenomMetaData()

	for i := 0; i < len(inputDenomData); i++ {
		app.BankKeeper.SetDenomMetaData(ctx, inputDenomData[i])
		storeDenomData := app.BankKeeper.GetDenomMetaData(ctx, inputDenomData[i].Base)
		suite.Require().Equal(inputDenomData[i], storeDenomData)
	}
}

func (suite *IntegrationTestSuite) TestInitDenomInGenesis() {
	inputDenomData := GetDenomMetaData()
	g := banktypes.DefaultGenesisState()
	g.DenomMetadata = inputDenomData
	bk := suite.app.BankKeeper
	// set denom data to state
	bk.InitGenesis(suite.ctx, g)

	for i := 0; i < len(inputDenomData); i++ {
		denom := inputDenomData[i]
		genDenomBase := bk.GetDenomMetaData(suite.ctx, denom.Base)
		suite.Require().Equal(denom, genDenomBase)
	}
}

func (suite *IntegrationTestSuite) TestGetStateGenesisData() {
	inputDenomData := GetDenomMetaData()
	sort.Slice(inputDenomData, func(i, j int) bool {
		return inputDenomData[i].Base < inputDenomData[j].Base
	})
	g := banktypes.DefaultGenesisState()
	g.DenomMetadata = inputDenomData
	bk := suite.app.BankKeeper
	bk.InitGenesis(suite.ctx, g)

	// wrapper around keeper get denom data all at once
	exportedBankGenesis := bk.ExportGenesis(suite.ctx)

	suite.Require().Equal(len(inputDenomData), len(exportedBankGenesis.DenomMetadata))

	for i := 0; i < len(inputDenomData); i++ {
		inputDenom := inputDenomData[i]
		genesisDenom := exportedBankGenesis.DenomMetadata[i]
		suite.Require().Equal(inputDenom, genesisDenom)
	}
}

func (suite *IntegrationTestSuite) TestExportGenesisData() {
	serverCtx := server.NewDefaultContext()
	serverCtx.Config.RootDir = os.TempDir()
	tmp, err := os.CreateTemp(serverCtx.Config.RootDir, "gen")
	serverCtx.Config.Genesis = tmp.Name()
	suite.Require().NoError(tmp.Close())

	encodingConfig := simapp.MakeTestEncodingConfig()

	genDoc := suite.newDefaultGenesisDoc(encodingConfig.Marshaler)

	expGenFile := serverCtx.Config.GenesisFile()
	err = genutil.ExportGenesisFile(genDoc, expGenFile)
	suite.Require().NoError(err)

	suite.app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simapp.DefaultConsensusParams,
			AppStateBytes:   genDoc.AppState,
		},
	)
	suite.app.Commit()

	inputDenomData := GetDenomMetaData()

	app, sdkctx := suite.app, suite.ctx

	for i := 0; i < len(inputDenomData); i++ {
		app.BankKeeper.SetDenomMetaData(sdkctx, inputDenomData[i])
		storeDenomData := app.BankKeeper.GetDenomMetaData(sdkctx, inputDenomData[i].Base)
		suite.Require().Equal(inputDenomData[i], storeDenomData)
	}

	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 2}})
	suite.app.Commit()

	expApp, err := app.ExportAppStateAndValidators(false, nil)
	suite.Require().NoError(err)

	var expAlteredState simapp.GenesisState
	err = json.Unmarshal(expApp.AppState, &expAlteredState)
	suite.Require().NoError(err)
	jsonAppState := string(expApp.AppState)
	fmt.Println(jsonAppState)
}

func (suite *IntegrationTestSuite) newDefaultGenesisDoc(cdc codec.Marshaler) *tmtypes.GenesisDoc {
	genesisState := simapp.NewDefaultGenesisState(cdc)

	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	suite.Require().NoError(err)

	genDoc := &tmtypes.GenesisDoc{}
	genDoc.ChainID = "lilmermaid"
	genDoc.Validators = nil
	genDoc.AppState = stateBytes

	return genDoc
}

func getSendEnabledParam() banktypes.Params {
	return banktypes.Params{
		SendEnabled: []*banktypes.SendEnabled{
			{
				Denom:   "eeur",
				Enabled: true,
			},
			{
				Denom:   "ungm",
				Enabled: true,
			},
			{
				Denom:   "ngm",
				Enabled: true,
			},
		},
	}
}

func (suite *IntegrationTestSuite) _TestInjectingDenomData2eMoney2Gen() {
	const emoney2Gen = "./emoney-2.export.json"
	genDoc, err := tmtypes.GenesisDocFromFile(emoney2Gen)
	suite.Require().NoError(err)
	var appState map[string]json.RawMessage
	err = json.Unmarshal(genDoc.AppState, &appState)
	suite.Require().NoError(err)

	encodingConfig := simapp.MakeTestEncodingConfig()
	var bkState banktypes.GenesisState
	bankSeg := appState[banktypes.ModuleName]
	bankSegS := string(bankSeg)
	fmt.Println(bankSegS)
	encodingConfig.Amino.MustUnmarshalJSON(bankSeg, &bkState)

	bkState.DenomMetadata = GetDenomMetaData()
	bkState.Params = getSendEnabledParam()

	bankSegS2 := encodingConfig.Amino.MustMarshalJSON(bkState)
	fmt.Println(string(bankSegS2))
	appState[banktypes.ModuleName] = encodingConfig.Amino.MustMarshalJSON(bkState)
}

func TestMigrateSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
