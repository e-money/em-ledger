// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package networktest

import (
	"bufio"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"io/ioutil"
	"os"
	"strings"
)

const KeyPwd = "pwd12345"
const Bip39Pwd = ""

type (
	KeyStore struct {
		path    string
		keybase keyring.Keyring

		Authority,
		Key1,
		Key2,
		Key3 Key

		MultiKey Key

		Validators []Key
	}

	Key struct {
		name    string
		keybase keyring.Keyring
		privkey string
		pubkey  cryptotypes.PubKey
		address sdk.AccAddress
	}
)

func newKey(name string, keybase keyring.Keyring) Key {
	// Extract key information to prevent future keystore access. Makes concurrent key usage possible.
	var (
		privkey, _ = keybase.ExportPrivKeyArmor(name, KeyPwd)
		info, _    = keybase.Key(name)
	)

	var address sdk.AccAddress
	var pubKey cryptotypes.PubKey
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

func (ks KeyStore) Keyring() keyring.Keyring {
	return ks.keybase
}

func (k Key) GetAddress() string {
	if k.address.Empty() {
		info, err := k.keybase.Key(k.name)
		if err != nil {
			panic(err)
		}

		k.address = info.GetAddress()
	}

	return k.address.String()
}

func (k Key) GetPublicKey() cryptotypes.PubKey {
	return k.pubkey
}

func (k Key) Sign(bz []byte) ([]byte, error) {
	signed, _, err := k.keybase.Sign(k.name, bz)
	return signed, err
}

func NewKeystore() (*KeyStore, error) {
	path, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}

	keybase, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, path, nil)
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

func (ks KeyStore) String() string {
	keyinfos, err := ks.keybase.List()
	if err != nil {
		return err.Error()
	}

	var sb strings.Builder
	for _, info := range keyinfos {
		sb.WriteString(fmt.Sprintf("%v - %v (%v)\n", info.GetName(), info.GetAddress().String(), info.GetAlgo()))
	}

	return sb.String()
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
		hdPath := sdk.GetConfig().GetFullFundraiserPath()
		_, err := ks.keybase.NewAccount(accountName, mnemonic, Bip39Pwd, hdPath, hd.Secp256k1)
		if err != nil {
			panic(err)
		}
	}
}

func initializeKeystore(kb keyring.Keyring) {
	keyDerivationPath := sdk.FullFundraiserPath

	_, err := kb.NewAccount("authoritykey",
		"play witness auto coast domain win tiny dress glare bamboo rent mule delay exact arctic vacuum laptop hidden siren sudden six tired fragile penalty",
		KeyPwd, keyDerivationPath, hd.Secp256k1)
	if err != nil {
		panic(err.Error())
	}
	_, _ = kb.NewAccount("key1",
		"document weekend believe whip diesel earth hope elder quiz pact assist quarter public deal height pulp roof organ animal health month holiday front pencil",
		KeyPwd, keyDerivationPath, hd.Secp256k1)

	_, _ = kb.NewAccount("key2",
		"treat ocean valid motor life marble syrup lady nephew grain cherry remember lion boil flock outside cupboard column dad rare build nut hip ostrich",
		KeyPwd, keyDerivationPath, hd.Secp256k1)

	_, _ = kb.NewAccount("key3",
		"rice short length buddy zero snake picture enough steak admit balance garage exit crazy cloud this sweet virus can aunt embrace picnic stick wheel",
		KeyPwd, keyDerivationPath, hd.Secp256k1)

	// Create a multisig key entry consisting of key1, key2 and key3 with a threshold of 2
	pks := make([]cryptotypes.PubKey, 3)
	for i, keyname := range []string{"key1", "key2", "key3"} {
		keyinfo, err := kb.Key(keyname)
		if err != nil {
			panic(err)
		}

		pks[i] = keyinfo.GetPubKey()
	}

	pk := multisig.NewLegacyAminoPubKey(2, pks)
	_, err = kb.SaveMultisig("multikey", pk)
	if err != nil {
		panic(err)
	}
}
