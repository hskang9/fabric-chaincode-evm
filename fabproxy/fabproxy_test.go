package fabproxy_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hyperledger/fabric-chaincode-evm/fabproxy"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fabproxy", func() {

	var (
		proxy     *fabproxy.FabProxy
		proxyAddr string
	)

	BeforeEach(func() {
		port := config.GinkgoConfig.ParallelNode + 5000
		proxy = fabproxy.NewFabProxy()

		go func() {
			proxy.Start(port)
		}()

		proxyAddr = fmt.Sprintf("http://localhost:%d", port)
	})

	Describe("eth_getCode", func() {
		var (
			req  *http.Request
			resp *http.Response
		)

		BeforeEach(func() {
			var err error
			//curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getCode","params":["0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b", "0x2"],"id":1}'

			body := strings.NewReader(`{"jsonrpc":"2.0","method":"eth_getCode","params":["0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b"],"id":1}`)
			req, err = http.NewRequest("POST", proxyAddr, body)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
		})

		JustBeforeEach(func() {
			var err error

			client := &http.Client{}
			resp, err = client.Do(req)
			Expect(err).ToNot(HaveOccurred())
		})

		It("can get the code of a previously deployed contract", func() {
			type responseBody struct {
				JsonRPC string `json:"jsonrpc"`
				Id      int    `json:"id"`
				Result  string `json:"result"`
			}
			expectedBody := responseBody{JsonRPC: "2.0", Id: 1, Result: "0x11110"}

			rBody, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			var respBody responseBody
			err = json.Unmarshal(rBody, &respBody)
			Expect(err).ToNot(HaveOccurred())

			Expect(respBody).To(Equal(expectedBody))
		})
	})

})
