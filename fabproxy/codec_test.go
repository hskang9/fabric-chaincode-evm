package fabproxy_test

import (
	"net/http"
	"strings"

	"github.com/gorilla/rpc/v2"
	"github.com/hyperledger/fabric-chaincode-evm/fabproxy"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Codec", func() {
	var (
		codec rpc.Codec
		req   *http.Request
	)

	Describe("NewRequest", func() {
		var codecRequest rpc.CodecRequest

		BeforeEach(func() {
			codec = fabproxy.NewRPCCodec()

			var err error
			body := strings.NewReader(`{"jsonrpc":"2.0","method":"someService_someMethod"}`)
			req, err = http.NewRequest("POST", "http://some-url", body)

			Expect(err).ToNot(HaveOccurred())
			codecRequest = codec.NewRequest(req)
		})

		Describe("Method", func() {
			It("returns the method being called on the fab proxy", func() {
				method, err := codecRequest.Method()
				Expect(err).ToNot(HaveOccurred())

				Expect(method).To(Equal("someService.SomeMethod"))
			})
		})
	})
})
