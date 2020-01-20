// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	apptypes "github.com/e-money/em-ledger/types"
)

func init() {
	// Be able to parse emoney bech32 encoded addresses.
	apptypes.ConfigureSDK()
}

func TestMsgUnjailGetSignBytes(t *testing.T) {
	addr := sdk.AccAddress("abcd")
	msg := NewMsgUnjail(sdk.ValAddress(addr))
	bytes := msg.GetSignBytes()
	require.Equal(
		t,
		`{"type":"cosmos-sdk/MsgUnjail","value":{"address":"emoneyvaloper1v93xxeqhz8086"}}`,
		string(bytes),
	)
}
