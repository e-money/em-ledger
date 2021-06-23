package liquidityprovider

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	apptypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
	"github.com/stretchr/testify/require"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tidwall/gjson"
	"testing"
)

func init() {
	apptypes.ConfigureSDK()
}

func TestGenesisStruct1(t *testing.T) {
	input := `{
		"accounts": []
	}`

	gs := types.GenesisState{}
	err := ModuleCdc.UnmarshalJSON([]byte(input), &gs)
	require.NoError(t, err)
	require.Empty(t, gs.Accounts)
}

func TestGenesisStruct2(t *testing.T) {
	input := `{
		"accounts": [
			{ 
				"address" : "emoney16j4trwyg8a3pfwqu2ely96tkzl05eh4vvyyfts",
				"mintable": [
					{
						"denom": "eeur",
						"amount": "50000000"
					}
				]
			},
			{ 
				"address" : "emoney1cs4323dyzu0wxfj4vc62m8q3xsczfavqx9x3zd",
				"mintable": [
					{
						"denom": "echf",
						"amount": "900"
					},
					{
						"denom": "esek",
						"amount": "4100000"
					}
				]
			}
		]
	}`

	gs := types.GenesisState{}
	err := ModuleCdc.UnmarshalJSON([]byte(input), &gs)
	require.NoError(t, err)
	require.Len(t, gs.Accounts, 2)

	require.Len(t, gs.Accounts[0].Mintable, 1)
	require.Len(t, gs.Accounts[1].Mintable, 2)

	require.True(t, gs.Accounts[0].Mintable.IsValid())
	require.True(t, gs.Accounts[1].Mintable.IsValid())
}

func TestSerialize(t *testing.T) {
	gs := &types.GenesisState{
		Accounts: []types.GenesisAcc{
			{
				Address: randomAddress().String(),
				Mintable: sdk.Coins{
					sdk.Coin{
						Denom:  "eeur",
						Amount: sdk.NewInt(6000000),
					},
					sdk.Coin{
						Denom:  "echf",
						Amount: sdk.NewInt(130000),
					},
				},
			},
			{
				Address: randomAddress().String(),
				Mintable: sdk.Coins{
					sdk.Coin{
						Denom:  "esek",
						Amount: sdk.NewInt(750000),
					},
				},
			},
		},
	}

	json, err := ModuleCdc.MarshalJSON(gs)
	require.NoError(t, err)
	doc := gjson.ParseBytes(json)

	require.Len(t, doc.Get("accounts").Array(), 2)

	require.Len(t, doc.Get("accounts.0.mintable").Array(), 2)
	require.Len(t, doc.Get("accounts.1.mintable").Array(), 1)
}

func randomAddress() sdk.AccAddress {
	return tmrand.Bytes(sdk.AddrLen)
}
