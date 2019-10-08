// +build bdd,quiet

package networktest

import (
	"io/ioutil"
)

func init() {
	// Silence the output of the testnet component
	output = ioutil.Discard
}
