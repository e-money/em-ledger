package types

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	baseAcc := auth.NewBaseAccount(addr, sdk.NewCoins(), priv.PubKey(), 1, 0)

	credit := sdk.NewCoin("x2eur", sdk.NewIntWithDecimal(1000, 2))
	lpAcc := NewLiquidityProviderAccount(baseAcc, sdk.NewCoins(credit))

	fmt.Println(lpAcc)

	// The credit window is not considered spendable.
	assert.Equal(t, sdk.NewCoins(), lpAcc.SpendableCoins(time.Now()))
}
