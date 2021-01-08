package testapp

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

const (
	Bech32MainPrefix = "app"
)

func SetBech32AddressPrefixes(config *sdk.Config) {
	config.SetBech32PrefixForAccount(Bech32MainPrefix, Bech32MainPrefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator, Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic)
	config.SetBech32PrefixForConsensusNode(Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus, Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic)
}

// Migrated from github/kava/app/test_common.go
// GeneratePrivKeyAddressPairsFromRand generates (deterministically) a total of n private keys and addresses.
// TODO only generate secp256 keys?
func GeneratePrivKeyAddressPairs(n int) (keys []crypto.PrivKey, addrs []sdk.AccAddress) {
	r := rand.New(rand.NewSource(12345)) // make the generation deterministic
	keys = make([]crypto.PrivKey, n)
	addrs = make([]sdk.AccAddress, n)
	for i := 0; i < n; i++ {
		secret := make([]byte, 32)
		_, err := r.Read(secret)
		if err != nil {
			panic("Could not read randomness")
		}
		if r.Int63()%2 == 0 {
			keys[i] = secp256k1.GenPrivKeySecp256k1(secret)
		} else {
			keys[i] = ed25519.GenPrivKeyFromSecret(secret)
		}
		addrs[i] = sdk.AccAddress(keys[i].PubKey().Address())
	}
	return
}
