// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

//go:build bdd && quiet
// +build bdd,quiet

package networktest

import (
	"io/ioutil"
)

func init() {
	// Silence the output of the testnet component
	output = ioutil.Discard
}
