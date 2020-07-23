// +build bdd

package emoney

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/sjson"
	"time"
)

var _ = Describe("Restricted denominationsq", func() {

	var (
		Key1 = testnet.Keystore.Key1
		Key2 = testnet.Keystore.Key2
		Key3 = testnet.Keystore.Key3
	)

	//tearDownAfterTests = false

	It("starts a new testnet", func() {
		awaitReady, err := testnet.RestartWithModifications(
			func(bz []byte) []byte {
				type Obj map[string]interface{}

				o := Obj{
					"Denom": "ungm",
					"Allowed": []string{
						Key1.GetAddress(),
					},
				}

				bz, _ = sjson.SetBytes(bz, "app_state.authority.restricted_denoms", []Obj{o})

				return bz
			})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(awaitReady()).To(BeTrue())
	})

	It("tests transfer restrictions", func() {
		time.Sleep(8 * time.Second) // Allow for a few blocks to avoid querying at height 1.

		emcli := testnet.NewEmcli()
		{
			// Key1 is whitelisted
			txid, success, err := emcli.Send(Key1, Key2, "5000ungm")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(success).To(BeTrue(), txid)
		}

		{
			// No white-listed accounts involved.
			txid, success, err := emcli.Send(Key3, Key2, "5000ungm")
			Expect(err).Should(HaveOccurred())
			Expect(success).To(BeFalse(), txid)
		}
		{
			// Key1 is whitelisted
			txid, success, err := emcli.Send(Key3, Key1, "5000ungm")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(success).To(BeTrue(), txid)
		}
	})
})
