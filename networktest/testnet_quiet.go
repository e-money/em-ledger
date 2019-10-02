// +build bdd,quiet

package network_test

import (
	"io/ioutil"
)

func init() {
	// Silence the output of the testnet component
	output = ioutil.Discard
}
