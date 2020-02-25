// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package networktest

import (
	"bufio"
	"fmt"
	keys2 "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/multisig"
	"io/ioutil"
	"os"
	"strings"
)

const KeyPwd = "pwd12345"

type (
	KeyStore struct {
		path    string
		keybase keys.Keybase

		Authority,
		Key1,
		Key2,
		Key3 Key

		MultiKey Key

		Validators []Key
	}

	Key struct {
		name    string
		keybase keys.Keybase
		privkey crypto.PrivKey
		pubkey  crypto.PubKey
		address sdk.AccAddress
	}
)

func newKey(name string, keybase keys.Keybase) Key {
	// Extract key information to prevent future keystore access. Makes concurrent key usage possible.
	var (
		privkey, _ = keybase.ExportPrivateKeyObject(name, KeyPwd)
		info, _    = keybase.Get(name)
	)

	var address sdk.AccAddress
	var pubKey crypto.PubKey
	if info != nil {
		pubKey = info.GetPubKey()
		address = info.GetAddress()
	}

	return Key{
		name:    name,
		keybase: keybase,
		privkey: privkey,
		pubkey:  pubKey,
		address: address,
	}
}

func (k Key) GetAddress() string {
	if k.address.Empty() {
		info, err := k.keybase.Get(k.name)
		if err != nil {
			panic(err)
		}

		k.address = info.GetAddress()
	}

	return k.address.String()
}

func (k Key) GetPublicKey() crypto.PubKey {
	if k.pubkey != nil {
		return k.pubkey
	}

	return k.privkey.PubKey()
}

func (k Key) Sign(bz []byte) ([]byte, error) {
	return k.privkey.Sign(bz)
}

func NewKeystore() (*KeyStore, error) {
	path, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}

	keybase, err := keys2.NewKeyBaseFromDir(path)
	if err != nil {
		return nil, err
	}

	initializeKeystore(keybase)

	// TODO This looks kind of horrible. Refactor to something prettier
	ks := &KeyStore{
		keybase:   keybase,
		path:      path,
		Authority: newKey("authoritykey", keybase),
		Key1:      newKey("key1", keybase),
		Key2:      newKey("key2", keybase),
		Key3:      newKey("key3", keybase),
		MultiKey:  newKey("multikey", keybase),
		Validators: []Key{
			newKey("validator0", keybase),
			newKey("validator1", keybase),
			newKey("validator2", keybase),
			newKey("validator3", keybase),
		},
	}

	return ks, nil
}

func (ks KeyStore) Close() {
	_ = os.RemoveAll(ks.path)
}

func (ks KeyStore) GetPath() string {
	return ks.path
}

func (ks KeyStore) addValidatorKeys(testnetoutput string) {
	scan := bufio.NewScanner(strings.NewReader(testnetoutput))
	seeds := make([]string, 0)
	for scan.Scan() {
		s := scan.Text()
		if strings.Contains(s, "Key mnemonic for Validator") {
			seed := strings.Split(s, ":")[1]
			seeds = append(seeds, strings.TrimSpace(seed))
		}
	}

	for i, mnemonic := range seeds {
		accountName := fmt.Sprintf("validator%v", i)
		_, err := ks.keybase.CreateAccount(accountName, mnemonic, "", KeyPwd, 0, 0)
		if err != nil {
			panic(err)
		}
	}
}

func initializeKeystore(kb keys.Keybase) {
	_, _ = kb.CreateAccount("authoritykey",
		"play witness auto coast domain win tiny dress glare bamboo rent mule delay exact arctic vacuum laptop hidden siren sudden six tired fragile penalty",
		"", KeyPwd, 0, 0)

	_, _ = kb.CreateAccount("key1",
		"document weekend believe whip diesel earth hope elder quiz pact assist quarter public deal height pulp roof organ animal health month holiday front pencil",
		"", KeyPwd, 0, 0)

	_, _ = kb.CreateAccount("key2",
		"treat ocean valid motor life marble syrup lady nephew grain cherry remember lion boil flock outside cupboard column dad rare build nut hip ostrich",
		"", KeyPwd, 0, 0)

	_, _ = kb.CreateAccount("key3",
		"rice short length buddy zero snake picture enough steak admit balance garage exit crazy cloud this sweet virus can aunt embrace picnic stick wheel",
		"", KeyPwd, 0, 0)

	// Create a multisig key entry consisting of key1, key2 and key3 with a threshold of 2
	pks := make([]crypto.PubKey, 3)
	for i, keyname := range []string{"key1", "key2", "key3"} {
		keyinfo, err := kb.Get(keyname)
		if err != nil {
			panic(err)
		}

		pks[i] = keyinfo.GetPubKey()
	}

	pk := multisig.NewPubKeyMultisigThreshold(2, pks)
	_, err := kb.CreateMulti("multikey", pk)
	if err != nil {
		panic(err)
	}
}
