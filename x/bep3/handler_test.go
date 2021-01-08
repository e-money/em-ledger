package bep3_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/bep3"
	app "github.com/e-money/em-ledger/x/bep3/testapp"
	"github.com/stretchr/testify/suite"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

type HandlerTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	handler sdk.Handler
	keeper  bep3.Keeper
	addrs   []sdk.AccAddress
}

func (suite *HandlerTestSuite) SetupTest() {
	ctx, bep3Keeper, accountKeeper, _, appModule := app.CreateTestComponents(suite.T())

	// Set up genesis state and initialize
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	coins := cs(c("bnb", 10000000000), c("ukava", 10000000000))

	for _, addr := range addrs {
		account := accountKeeper.NewAccountWithAddress(ctx, addr)
		if err := account.SetCoins(coins); err != nil {
			panic(err)
		}
		accountKeeper.SetAccount(ctx, account)
	}

	appModule.InitGenesis(ctx, NewBep3GenState(addrs[0]))

	suite.addrs = addrs
	suite.handler = bep3.NewHandler(bep3Keeper)
	suite.keeper = bep3Keeper
	suite.ctx = ctx
}

func (suite *HandlerTestSuite) AddAtomicSwap() (tmbytes.HexBytes, tmbytes.HexBytes) {
	expireTimeSpan := bep3.DefaultSwapTimeSpan
	amount := cs(c("bnb", int64(50000)))
	timestamp := ts(0)
	randomNumber, _ := bep3.GenerateSecureRandomNumber()
	randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)

	// Create atomic swap and check err to confirm creation
	err := suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, expireTimeSpan,
		suite.addrs[0], suite.addrs[1], TestSenderOtherChain, TestRecipientOtherChain,
		amount, true)
	suite.Nil(err)

	swapID := bep3.CalculateSwapID(randomNumberHash, suite.addrs[0], TestSenderOtherChain)
	return swapID, randomNumber[:]
}

func (suite *HandlerTestSuite) TestMsgCreateAtomicSwap() {
	amount := cs(c("bnb", int64(10000)))
	timestamp := ts(0)
	randomNumber, _ := bep3.GenerateSecureRandomNumber()
	randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)

	msg := bep3.NewMsgCreateAtomicSwap(
		suite.addrs[0], suite.addrs[2], TestRecipientOtherChain,
		TestSenderOtherChain, randomNumberHash, timestamp, amount,
		bep3.DefaultSwapTimeSpan)

	res, err := suite.handler(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
}

func (suite *HandlerTestSuite) TestMsgClaimAtomicSwap() {
	// Attempt claim msg on fake atomic swap
	badRandomNumber, _ := bep3.GenerateSecureRandomNumber()
	badRandomNumberHash := bep3.CalculateRandomHash(badRandomNumber[:], ts(0))
	badSwapID := bep3.CalculateSwapID(badRandomNumberHash, suite.addrs[0], TestSenderOtherChain)
	badMsg := bep3.NewMsgClaimAtomicSwap(suite.addrs[0], badSwapID, badRandomNumber[:])
	badRes, err := suite.handler(suite.ctx, badMsg)
	suite.Require().Error(err)
	suite.Require().Nil(badRes)

	// Add an atomic swap before attempting new claim msg
	swapID, randomNumber := suite.AddAtomicSwap()
	msg := bep3.NewMsgClaimAtomicSwap(suite.addrs[0], swapID, randomNumber)
	res, err := suite.handler(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
}

// getContextPlusSec returns a context forward or backward in time and block
// index. Assuming 1 second finality.
func (suite *HandlerTestSuite) getContextPlusSec(plusSeconds uint64) sdk.Context {
	offset := int64(plusSeconds)
	ctx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Duration(offset) * time.Second))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + offset)

	return ctx
}

func (suite *HandlerTestSuite) TestMsgRefundAtomicSwap() {
	// Attempt refund msg on fake atomic swap
	badRandomNumber, _ := bep3.GenerateSecureRandomNumber()
	badRandomNumberHash := bep3.CalculateRandomHash(badRandomNumber[:], ts(0))
	badSwapID := bep3.CalculateSwapID(badRandomNumberHash, suite.addrs[0], TestSenderOtherChain)
	badMsg := bep3.NewMsgRefundAtomicSwap(suite.addrs[0], badSwapID)
	badRes, err := suite.handler(suite.ctx, badMsg)
	suite.Require().Error(err)
	suite.Require().Nil(badRes)

	// Add an atomic swap and build refund msg
	swapID, _ := suite.AddAtomicSwap()
	msg := bep3.NewMsgRefundAtomicSwap(suite.addrs[0], swapID)

	// Attempt to refund active atomic swap
	res1, err := suite.handler(suite.ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res1)

	// Expire the atomic swap with begin blocker and attempt refund
	laterCtx := suite.getContextPlusSec(bep3.DefaultSwapTimeSpan)
	bep3.BeginBlocker(laterCtx, suite.keeper)
	res2, err := suite.handler(laterCtx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res2)
}

func (suite *HandlerTestSuite) TestInvalidMsg() {
	res, err := suite.handler(suite.ctx, sdk.NewTestMsg())
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
