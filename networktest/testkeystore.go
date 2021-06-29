// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package networktest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	KeyPwd           = "pwd12345"
	Bip39Pwd         = ""
	DeputyKey        = "deputykey"
	AuthKey          = "authoritykey"
	Key1             = "key1"
	Key2             = "key2"
	Key3             = "key3"
	Key4             = "key4"
	Key5             = "key5"
	Key6             = "key6"
	MultiKey         = "multikey"
	MultiKey2        = "multikey2"
	LocalNetReuse    = "localnet_reuse"
	startForReUseEnv = "REUSE"
)

type (
	KeyStore struct {
		path    string
		keybase keyring.Keyring

		Authority,
		DeputyKey,
		Key1,
		Key2,
		Key3,
		Key4,
		Key5,
		Key6 Key

		MultiKey, MultiKey2 Key

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

func (k Key) GetName() string {
	return k.name
}

func (k Key) Sign(bz []byte) ([]byte, error) {
	signed, _, err := k.keybase.Sign(k.name, bz)
	return signed, err
}

// NewKeystore creates a keystore considering reusableLocation as the fixed
// location and reusableIsUp on whether to persist the keystore on disk.
func NewKeystore(reusableLocation, reusableIsUp bool) (*KeyStore, error) {
	var (
		path string
		err  error
	)

	// random tmp path
	if !reusableIsUp && !reusableLocation {
		path, err = ioutil.TempDir("", "")
		if err != nil {
			return nil, err
		}
	} else {
		path = "/tmp/" + LocalNetReuse
	}

	// persist keystore at fixed location.
	if !reusableIsUp && reusableLocation {
		if err := os.RemoveAll(path); err != nil {
			panic(err)
		}
		if err := os.Mkdir(path, 0o700); err != nil {
			panic(err)
		}
	}

	var keybase keyring.Keyring
	if !reusableIsUp {
		keybase, err = keyring.New(
			sdk.KeyringServiceName(), keyring.BackendTest, path, nil,
		)
	} else {
		// in memory, do not rewrite running setup
		keybase, err = keyring.New(
			sdk.KeyringServiceName(), keyring.BackendMemory, path, nil,
		)
	}
	if err != nil {
		return nil, err
	}

	initializeKeystore(keybase)

	ks := &KeyStore{
		keybase:   keybase,
		path:      path,
		Authority: newKey(AuthKey, keybase),
		DeputyKey: newKey(DeputyKey, keybase),
		Key1:      newKey(Key1, keybase),
		Key2:      newKey(Key2, keybase),
		Key3:      newKey(Key3, keybase),
		Key4:      newKey(Key4, keybase),
		Key5:      newKey(Key5, keybase),
		Key6:      newKey(Key6, keybase),
		MultiKey:  newKey(MultiKey, keybase),
		MultiKey2:  newKey(MultiKey2, keybase),
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

func (ks KeyStore) addDeputyKey() {
	mn := "play witness auto coast domain win tiny dress glare bamboo rent mule delay exact arctic vacuum laptop hidden siren sudden six tired fragile penalty"
	// create the deputy account
	deputyAccount, err := ks.keybase.NewAccount("deputykey", mn, "", sdk.FullFundraiserPath, hd.Secp256k1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("deputy address: %s\nmnemonic: %s\n", deputyAccount.GetAddress().String(), mn)
}

func (ks KeyStore) addValidatorKeys(workDir string, numberNodes int) {
	for i := 0; i < numberNodes; i++ {
		fileName := filepath.Join(workDir, fmt.Sprintf("node%d", i), "key_seed.json")
		bz, err := ioutil.ReadFile(fileName)
		if err != nil {
			panic(fmt.Sprintf("failed to load key see %q: %s", fileName, err))
		}
		var obj struct {
			Mnemonic string `json:"secret"`
		}
		if err := json.Unmarshal(bz, &obj); err != nil {
			panic(err)
		}
		accountName := fmt.Sprintf("validator%v", i)
		hdPath := sdk.GetConfig().GetFullFundraiserPath()
		_, err = ks.keybase.NewAccount(accountName, obj.Mnemonic, Bip39Pwd, hdPath, hd.Secp256k1)
		if err != nil {
			panic(err)
		}
	}
}

func initializeKeystore(kb keyring.Keyring) {
	keyDerivationPath := sdk.FullFundraiserPath

	const mnemonic1 = "then nuclear favorite advance plate glare shallow enhance replace embody list dose quick scale service sentence hover announce advance nephew phrase order useful this"
	ac1, err := kb.NewAccount("authoritykey", mnemonic1, "", keyDerivationPath, hd.Secp256k1)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Created auth account %s from mnemonic: %s\n", ac1.GetAddress(), mnemonic1)

	mn := "play witness auto coast domain win tiny dress glare bamboo rent mule delay exact arctic vacuum laptop hidden siren sudden six tired fragile penalty"
	// create the deputy account
	deputyAccount, _ := kb.NewAccount("deputykey", mn, "", keyDerivationPath, hd.Secp256k1)
	fmt.Printf("deputy address: %s\nmnemonic: %s\n", deputyAccount.GetAddress().String(), mn)

	const mnemonic2 = "document weekend believe whip diesel earth hope elder quiz pact assist quarter public deal height pulp roof organ animal health month holiday front pencil"
	ac2, _ := kb.NewAccount("key1", mnemonic2, "", keyDerivationPath, hd.Secp256k1)
	fmt.Printf("Created key1 account %s from mnemonic: %s\n", ac2.GetAddress(), mnemonic2)

	const mnemonic3 = "treat ocean valid motor life marble syrup lady nephew grain cherry remember lion boil flock outside cupboard column dad rare build nut hip ostrich"
	ac3, _ := kb.NewAccount("key2", mnemonic3, "", keyDerivationPath, hd.Secp256k1)
	fmt.Printf("Created account %s from mnemonic: %s\n", ac3.GetAddress(), mnemonic3)

	const mnemonic4 = "rice short length buddy zero snake picture enough steak admit balance garage exit crazy cloud this sweet virus can aunt embrace picnic stick wheel"
	ac4, _ := kb.NewAccount("key3", mnemonic4, "", keyDerivationPath, hd.Secp256k1)
	fmt.Printf("Created account %s from mnemonic: %s\n", ac4.GetAddress(), mnemonic4)

	const mnemonic5 = "census museum crew rude tower vapor mule rib weasel faith page cushion rain inherit much cram that blanket occur region track hub zero topple"
	ac5, _ := kb.NewAccount("key4", mnemonic5, "", keyDerivationPath, hd.Secp256k1)
	fmt.Printf("Created account %s from mnemonic: %s\n", ac5.GetAddress(), mnemonic5)

	const mnemonic6 = "flavor print loyal canyon expand salmon century field say frequent human dinosaur frame claim bridge affair web way direct win become merry crash frequent"
	ac6, _ := kb.NewAccount("key5", mnemonic6, "", keyDerivationPath, hd.Secp256k1)
	fmt.Printf("Created account %s from mnemonic: %s\n", ac6.GetAddress(), mnemonic6)

	const mnemonic7 = "very health column only surface project output absent outdoor siren reject era legend legal twelve setup roast lion rare tunnel devote style random food"
	ac7, _ := kb.NewAccount("key6", mnemonic7, "", keyDerivationPath, hd.Secp256k1)
	fmt.Printf("Created account %s from mnemonic: %s\n", ac7.GetAddress(), mnemonic7)

	createMultisig(kb, MultiKey, []string{"key1", "key2", "key3"}, 2, err)
	createMultisig(kb, MultiKey2, []string{"key1", "key3", "key5"}, 2, err)
}

func createMultisig(kb keyring.Keyring, keyName string, keys []string, threshold int, err error) keyring.Info{
	pks := make([]cryptotypes.PubKey, len(keys))
	for i, keyname := range keys{
		keyinfo, err := kb.Key(keyname)
		if err != nil {
			panic(err)
		}

		pks[i] = keyinfo.GetPubKey()
	}

	sort.Slice(
		pks, func(i, j int) bool {
			return bytes.Compare(pks[i].Address(), pks[j].Address()) < 0
		},
	)

	pk := multisig.NewLegacyAminoPubKey(threshold, pks)

	mSig, err := kb.SaveMultisig(keyName, pk)
	if err != nil {
		panic(err)
	}

	return mSig
}
