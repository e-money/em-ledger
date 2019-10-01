// +build bdd,quiet

package network_test

import (
	"fmt"
	"io/ioutil"
)

func init() {
	fmt.Println(" *** Silencing test output")
	output = ioutil.Discard
}
