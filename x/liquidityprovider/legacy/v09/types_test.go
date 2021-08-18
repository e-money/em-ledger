package v09

import (
	"testing"

	"github.com/stretchr/testify/require"

	apptypes "github.com/e-money/em-ledger/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
)

// From emoney-2 export
const json1 = `{
          "type": "e-money/LiquidityProviderAccount",
          "value": {
            "Account": {
              "type": "cosmos-sdk/Account",
              "value": {
                "account_number": "81",
                "address": "emoney1s73cel9vxllx700eaeuqr70663w5f0twzcks3l",
                "coins": [
                  {
                    "amount": "96850000",
                    "denom": "ungm"
                  }
                ],
                "public_key": {
                  "type": "tendermint/PubKeySecp256k1",
                  "value": "A+fi3hhzZjrM7+TPetNPW+0FGOrT2Q87nQS+XSs9rY4g"
                },
                "sequence": "14"
              }
            },
            "mintable": [
              {
                "amount": "338273454390",
                "denom": "eeur"
              }
            ]
          }
        }`

func init() {
	apptypes.ConfigureSDK()
}

func TestParseLPJson(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	v039auth.RegisterLegacyAminoCodec(cdc)
	RegisterLegacyAminoCodec(cdc)

	var account LiquidityProviderAccount
	cdc.MustUnmarshalJSON([]byte(json1), &account)
	require.NotNil(t, account.BaseAccount)

	require.Len(t, account.Mintable, 1)
	require.Equal(t, uint64(81), account.AccountNumber)

	balance := account.BaseAccount.GetCoins()
	require.Len(t, balance, 1)
	require.Equal(t, balance[0].Amount, sdk.NewInt(96850000))
}
