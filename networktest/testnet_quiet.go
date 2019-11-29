// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd,quiet

package networktest

import (
	"io/ioutil"
)

func init() {
	// Silence the output of the testnet component
	output = ioutil.Discard
}
