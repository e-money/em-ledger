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

type ABCITestSuite struct {
	suite.Suite
	keeper bep3.Keeper

	ctx           sdk.Context
	addrs         []sdk.AccAddress
	swapIDs       []tmbytes.HexBytes
	randomNumbers []tmbytes.HexBytes
}

func (suite *ABCITestSuite) SetupTest() {
	ctx, bep3Keeper, accountKeeper, _, appModule := app.CreateTestComponents(suite.T())

	// Set up auth GenesisState
	_, addrs := app.GeneratePrivKeyAddressPairs(12)
	coins := cs(c("bnb", 10000000000), c("ukava", 10000000000))

	for _, addr := range addrs {
		account := accountKeeper.NewAccountWithAddress(ctx, addr)
		if err := account.SetCoins(coins); err != nil {
			panic(err)
		}
		accountKeeper.SetAccount(ctx, account)
	}

	appModule.InitGenesis(ctx, NewBep3GenState(addrs[11]))

	suite.ctx = ctx
	suite.addrs = addrs
	suite.keeper = bep3Keeper
	suite.ResetKeeper()
}

func (suite *ABCITestSuite) ResetKeeper() {
	var swapIDs []tmbytes.HexBytes
	var randomNumbers []tmbytes.HexBytes
	for i := 0; i < 10; i++ {
		// Set up atomic swap variables
		amount := cs(c("bnb", int64(10000)))
		timestamp := ts(i)
		swapTimeSpan := bep3.DefaultSwapTimeSpan
		randomNumber, _ := bep3.GenerateSecureRandomNumber()
		randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)

		// Create atomic swap and check err to confirm creation
		err := suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, swapTimeSpan,
			suite.addrs[11], suite.addrs[i], TestSenderOtherChain, TestRecipientOtherChain,
			amount, true)
		suite.Nil(err)

		// Store swap's calculated ID and secret random number
		swapID := bep3.CalculateSwapID(randomNumberHash, suite.addrs[11], TestSenderOtherChain)
		swapIDs = append(swapIDs, swapID)
		randomNumbers = append(randomNumbers, randomNumber[:])
	}
	suite.swapIDs = swapIDs
	suite.randomNumbers = randomNumbers
}

// getContextPlusSec returns a context forward or backward in time and block
// index. Assuming 1 second finality.
func (suite *ABCITestSuite) getContextPlusSec(plusSeconds uint64) sdk.Context {
	offset := int64(plusSeconds)
	ctx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Duration(offset) * time.Second))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + offset)

	return ctx
}

func (suite *ABCITestSuite) TestBeginBlocker_UpdateExpiredAtomicSwaps() {
	testCases := []struct {
		name            string
		firstCtx        sdk.Context
		secondCtx       sdk.Context
		expectedStatus  bep3.SwapStatus
		expectInStorage bool
	}{
		{
			name:            "normal",
			firstCtx:        suite.ctx,
			secondCtx:       suite.getContextPlusSec(10),
			expectedStatus:  bep3.Open,
			expectInStorage: true,
		},
		{
			name:            "after expiration",
			firstCtx:        suite.getContextPlusSec(bep3.DefaultSwapTimeSpan),
			secondCtx:       suite.getContextPlusSec(bep3.DefaultSwapTimeSpan + 10),
			expectedStatus:  bep3.Expired,
			expectInStorage: true,
		},
		{
			name:            "after completion",
			firstCtx:        suite.getContextPlusSec(1),
			secondCtx:       suite.getContextPlusSec(10),
			expectedStatus:  bep3.Completed,
			expectInStorage: true,
		},
		{
			name:            "after deletion",
			firstCtx:        suite.getContextPlusSec(bep3.DefaultSwapTimeSpan),
			secondCtx:       suite.getContextPlusSec(bep3.DefaultSwapTimeSpan + bep3.DefaultLongtermStorageDuration),
			expectedStatus:  bep3.NULL,
			expectInStorage: false,
		},
	}

	for _, tc := range testCases {
		// Reset keeper and run the initial begin blocker
		suite.ResetKeeper()
		suite.Run(tc.name, func() {
			// Complete Swap Requests: Claim or Refund
			as := suite.keeper.GetAllAtomicSwaps(suite.ctx)
			for _, s := range as {
				suite.Nil(s.Validate())
			}
			bep3.BeginBlocker(tc.firstCtx, suite.keeper)

			switch tc.expectedStatus {
			case bep3.Completed:
				for i, swapID := range suite.swapIDs {
					err := suite.keeper.ClaimAtomicSwap(tc.firstCtx, suite.addrs[5], swapID, suite.randomNumbers[i])
					suite.Nil(err)
				}
			case bep3.NULL:
				for _, swapID := range suite.swapIDs {
					err := suite.keeper.RefundAtomicSwap(tc.firstCtx, suite.addrs[5], swapID)
					suite.Nil(err)
				}
			}

			// Run the second begin blocker
			// Check final swap requests status or result and storage status.
			bep3.BeginBlocker(tc.secondCtx, suite.keeper)

			// Check each swap's availability and status
			for _, swapID := range suite.swapIDs {
				storedSwap, found := suite.keeper.GetAtomicSwap(tc.secondCtx, swapID)
				if tc.expectInStorage {
					suite.True(found)
				} else {
					suite.False(found)
				}
				suite.Equal(tc.expectedStatus, storedSwap.Status)
			}
		})
	}
}

func (suite *ABCITestSuite) TestBeginBlocker_DeleteClosedAtomicSwapsFromLongtermStorage() {
	type Action int
	const (
		NULL   Action = 0x00
		Refund Action = 0x01
		Claim  Action = 0x02
	)

	testCases := []struct {
		name            string
		firstCtx        sdk.Context
		action          Action
		secondCtx       sdk.Context
		expectInStorage bool
	}{
		{
			name:            "no action with long storage duration",
			firstCtx:        suite.ctx,
			action:          NULL,
			secondCtx:       suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + int64(bep3.DefaultLongtermStorageDuration)),
			expectInStorage: true,
		},
		{
			name:            "claim with short storage duration",
			firstCtx:        suite.ctx,
			action:          Claim,
			secondCtx:       suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 5000),
			expectInStorage: true,
		},
		{
			name:            "claim with long storage duration",
			firstCtx:        suite.ctx,
			action:          Claim,
			secondCtx:       suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + int64(bep3.DefaultLongtermStorageDuration)),
			expectInStorage: false,
		},
		{
			name:            "refund with short storage duration",
			firstCtx:        suite.ctx,
			action:          Refund,
			secondCtx:       suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 5000),
			expectInStorage: true,
		},
		{
			name:            "refund with long storage duration",
			firstCtx:        suite.ctx,
			action:          Refund,
			secondCtx:       suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + int64(bep3.DefaultLongtermStorageDuration)),
			expectInStorage: false,
		},
	}

	for _, tc := range testCases {
		// Reset keeper and run the initial begin blocker
		suite.ResetKeeper()
		suite.Run(tc.name, func() {
			bep3.BeginBlocker(tc.firstCtx, suite.keeper)

			switch tc.action {
			case Claim:
				for i, swapID := range suite.swapIDs {
					err := suite.keeper.ClaimAtomicSwap(tc.firstCtx, suite.addrs[5], swapID, suite.randomNumbers[i])
					suite.Nil(err)
				}
			case Refund:
				for _, swapID := range suite.swapIDs {
					swap, _ := suite.keeper.GetAtomicSwap(tc.firstCtx, swapID)
					refundCtx := suite.ctx.WithBlockTime(time.Unix(int64(swap.ExpireTimestamp), 0))
					bep3.BeginBlocker(refundCtx, suite.keeper)
					err := suite.keeper.RefundAtomicSwap(refundCtx, suite.addrs[5], swapID)
					suite.Nil(err)
					// Add expiration timestamp to second ctx block timestamp
					tc.secondCtx = tc.secondCtx.WithBlockTime(
						time.Unix(tc.secondCtx.BlockTime().Unix()+int64(swap.ExpireTimestamp)-swap.Timestamp, 0))
				}
			}

			// Run the second begin blocker
			bep3.BeginBlocker(tc.secondCtx, suite.keeper)

			// Check each swap's availability and status
			for _, swapID := range suite.swapIDs {
				_, found := suite.keeper.GetAtomicSwap(tc.secondCtx, swapID)
				if tc.expectInStorage {
					suite.True(found)
				} else {
					suite.False(found)
				}
			}
		})
	}
}

func TestABCITestSuite(t *testing.T) {
	suite.Run(t, new(ABCITestSuite))
}
