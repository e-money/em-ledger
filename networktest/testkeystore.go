package networktest

import (
	keys2 "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"io/ioutil"
	"os"
)

const KeyPwd = "pwd12345"

type KeyStore struct {
	path    string
	keybase keys.Keybase

	Authority,
	Key1,
	Key2 Key
}

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

	ks := &KeyStore{
		keybase:   keybase,
		path:      path,
		Authority: newKey("authoritykey", keybase),
		Key1:      newKey("key1", keybase),
		Key2:      newKey("key2", keybase),
	}

	return ks, nil
}

func (ks KeyStore) Close() {
	_ = os.RemoveAll(ks.path)
}

func (ks KeyStore) GetPath() string {
	return ks.path
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
