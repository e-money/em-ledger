package v040

import (
	"sort"
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

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

func (s *IntegrationTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())

	s.app = app
	s.ctx = ctx
}

func (s *IntegrationTestSuite) TestGetStateGenesisData() {
	inputDenomData := GetDenomMetaData()
	sort.Slice(inputDenomData, func(i, j int) bool {
		return inputDenomData[i].Base < inputDenomData[j].Base
	})
	g := banktypes.DefaultGenesisState()
	g.DenomMetadata = inputDenomData
	bk := s.app.BankKeeper
	bk.InitGenesis(s.ctx, g)

	// wrapper around keeper get denom data all at once
	exportedBankGenesis := bk.ExportGenesis(s.ctx)

	s.Require().Equal(len(inputDenomData), len(exportedBankGenesis.DenomMetadata))

	for i := 0; i < len(inputDenomData); i++ {
		inputDenom := inputDenomData[i]
		genesisDenom := exportedBankGenesis.DenomMetadata[i]
		s.Require().Equal(inputDenom, genesisDenom)
	}
}

func TestMigrateSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
