package v040

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v038"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	json := `{
    "type": "cosmos-sdk/Account",
    "value":
    {
        "account_number": "65",
        "address": "cosmos1xxkueklal9vejv9unqu80w9vptyepfa95pd53u",
        "coins":
        [
            {
                "amount": "99300000",
                "denom": "ungm"
            }
        ],
        "public_key":
        {
            "type": "tendermint/PubKeyMultisigThreshold",
            "value":
            {
                "pubkeys":
                [
                    {
                        "type": "tendermint/PubKeySecp256k1",
                        "value": "A2rU+hzTlIQSHhHhsQZidUndKcuRnNFIO2yclhg3357K"
                    },
                    {
                        "type": "tendermint/PubKeySecp256k1",
                        "value": "A/Xo27k+o+5i24gbxolCR8dbeLoO2g+fsN2nVd/7OLnp"
                    },
                    {
                        "type": "tendermint/PubKeySecp256k1",
                        "value": "A0F4FJNaoJ8NXNFTYIVqt0/nIzDGkbknhESfMRG7fQow"
                    }
                ],
                "threshold": "2"
            }
        },
        "sequence": "4"
    }
}`

	cdc := codec.NewLegacyAmino()
	v039auth.RegisterLegacyAminoCodec(cdc)

	var acc v038auth.GenesisAccount
	cdc.MustUnmarshalJSON([]byte(json), &acc)

	acc2, ok := acc.(*v039auth.BaseAccount)
	require.True(t, ok)

	mpk, ok := acc2.PubKey.(*multisig.LegacyAminoPubKey)
	require.True(t, ok)

	require.Equal(t, mpk.Threshold, uint32(2))
	require.Len(t, mpk.PubKeys, 3)

	// println("Multisig threshold", mpk.Threshold)
	// println("Pubkey count", len(mpk.PubKeys))

	// ERROR IS HERE:
	// tmpvendor/cosmos-sdk/crypto/keys/multisig/amino.go:79
	// The iteration at the bottom of the method iterates over an empty array.
}
