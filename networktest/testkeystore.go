package networktest

import (
	"bufio"
	"fmt"
	keys2 "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
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

		Validators []Key
	}

	Key struct {
		name    string
		keybase keys.Keybase
	}
)

func newKey(name string, keybase keys.Keybase) Key {
	return Key{
		name:    name,
		keybase: keybase,
	}
}

func (k Key) GetAddress() string {
	info, err := k.keybase.Get(k.name)
	if err != nil {
		panic(err) // TODO Better errorhandling
	}

	accAddress := info.GetAddress()
	return accAddress.String()
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
		Validators: []Key{
			newKey("validator1", keybase),
			newKey("validator2", keybase),
			newKey("validator3", keybase),
			newKey("validator4", keybase),
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
}
