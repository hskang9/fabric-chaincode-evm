package fabproxy_test

import (
	"errors"
	"net/http"

	"github.com/hyperledger/fabric-chaincode-evm/fabproxy"
	"github.com/hyperledger/fabric-chaincode-evm/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var channelID = "mychannel"
var _ = FDescribe("Ethservice", func() {
	var (
		ethservice *fabproxy.EthService

		fabSDK       *mocks.MockSDK
		mockChClient *mocks.MockChannelClient
	)

	BeforeEach(func() {
		fabSDK = &mocks.MockSDK{}
		mockChClient = &mocks.MockChannelClient{}

		fabSDK.GetChannelClientReturns(mockChClient, nil)
		ethservice = fabproxy.NewEthService(fabSDK, channelID)

	})

	//TODO: Fix the query args. Need to find out if fab sdk uses invoke or the "function" used by the chaincode as the function arg
	//TODO: Fix getTransactionReceipt tests

	Describe("GetCode", func() {
		var sampleCode []byte
		BeforeEach(func() {
			sampleCode = []byte("sample-code")
			mockChClient.QueryReturns(channel.Response{
				Payload: sampleCode,
			}, nil)
		})

		It("returns the code associated to that address", func() {
			sampleAddress := "0x1234567123"
			var reply string

			err := ethservice.GetCode(&http.Request{}, &sampleAddress, &reply)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockChClient.QueryCallCount()).To(Equal(1))
			chReq, reqOpts := mockChClient.QueryArgsForCall(0)
			Expect(chReq).To(Equal(channel.Request{
				ChaincodeID: fabproxy.EVMSCC,
				Fcn:         "getCode",
				Args:        [][]byte{[]byte(sampleAddress)},
			}))

			Expect(reqOpts).To(HaveLen(0))

			Expect(reply).To(Equal(string(sampleCode)))
		})

		Context("when getting the channel client errors ", func() {
			BeforeEach(func() {
				fabSDK.GetChannelClientReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				sampleAddress := "0x1234567123"
				var reply string

				err := ethservice.GetCode(&http.Request{}, &sampleAddress, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to generate channel client"))

				Expect(reply).To(BeEmpty())
			})
		})

		Context("when querying the ledger errors", func() {
			BeforeEach(func() {
				mockChClient.QueryReturns(channel.Response{}, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				sampleAddress := "0x1234567123"
				var reply string

				err := ethservice.GetCode(&http.Request{}, &sampleAddress, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to query the ledger"))

				Expect(reply).To(BeEmpty())
			})
		})
	})

	Describe("Call", func() {
		var sampleResponse []byte
		BeforeEach(func() {
			sampleResponse = []byte("sample response")
			mockChClient.QueryReturns(channel.Response{
				Payload: sampleResponse,
			}, nil)
		})

		It("returns the value of the simulation of executing a smart contract", func() {
			sampleArgs := &fabproxy.EthArgs{
				To:   "0x1234567123",
				Data: "sample-data",
			}

			var reply string

			err := ethservice.Call(&http.Request{}, sampleArgs, &reply)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockChClient.QueryCallCount()).To(Equal(1))
			chReq, reqOpts := mockChClient.QueryArgsForCall(0)
			Expect(chReq).To(Equal(channel.Request{
				ChaincodeID: fabproxy.EVMSCC,
				Fcn:         "invoke",
				Args:        [][]byte{[]byte(sampleArgs.To), []byte(sampleArgs.Data)},
			}))

			Expect(reqOpts).To(HaveLen(0))

			Expect(reply).To(Equal(string(sampleResponse)))
		})

		Context("when getting the channel client errors ", func() {
			BeforeEach(func() {
				fabSDK.GetChannelClientReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply string

				err := ethservice.Call(&http.Request{}, &fabproxy.EthArgs{}, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to generate channel client"))

				Expect(reply).To(BeEmpty())
			})
		})

		Context("when querying the ledger errors", func() {
			BeforeEach(func() {
				mockChClient.QueryReturns(channel.Response{}, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply string

				err := ethservice.Call(&http.Request{}, &fabproxy.EthArgs{}, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to query the ledger"))

				Expect(reply).To(BeEmpty())
			})
		})
	})

	Describe("SendTransaction", func() {
		var (
			sampleResponse channel.Response
		)

		BeforeEach(func() {
			sampleResponse = channel.Response{
				Payload:       []byte("sample-response"),
				TransactionID: "1",
			}
			mockChClient.ExecuteReturns(sampleResponse, nil)
		})

		It("returns the value of the simulation of executing a smart contract", func() {
			sampleArgs := &fabproxy.EthArgs{
				To:   "0x1234567123",
				Data: "sample-data",
			}

			var reply string

			err := ethservice.SendTransaction(&http.Request{}, sampleArgs, &reply)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockChClient.ExecuteCallCount()).To(Equal(1))
			chReq, reqOpts := mockChClient.ExecuteArgsForCall(0)
			Expect(chReq).To(Equal(channel.Request{
				ChaincodeID: fabproxy.EVMSCC,
				Fcn:         "invoke",
				Args:        [][]byte{[]byte(sampleArgs.To), []byte(sampleArgs.Data)},
			}))

			Expect(reqOpts).To(HaveLen(0))

			Expect(reply).To(Equal(string(sampleResponse.TransactionID)))
		})

		Context("when getting the channel client errors ", func() {
			BeforeEach(func() {
				fabSDK.GetChannelClientReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply string

				err := ethservice.SendTransaction(&http.Request{}, &fabproxy.EthArgs{}, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to generate channel client"))

				Expect(reply).To(BeEmpty())
			})
		})

		Context("when querying the ledger errors", func() {
			BeforeEach(func() {
				mockChClient.ExecuteReturns(channel.Response{}, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply string

				err := ethservice.SendTransaction(&http.Request{}, &fabproxy.EthArgs{}, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to execute transaction"))

				Expect(reply).To(BeEmpty())
			})
		})
	})

	FDescribe("GetTransactionReceipt", func() {
		var sampleCode []byte
		BeforeEach(func() {
			sampleCode = []byte("sample-code")
			mockChClient.QueryReturns(channel.Response{
				Payload: sampleCode,
			}, nil)
		})

		It("returns the transaction receipt associated to that transaction address", func() {
			sampleTransactionID := "0x1234567123"
			var reply string

			err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockChClient.QueryCallCount()).To(Equal(1))
			chReq, reqOpts := mockChClient.QueryArgsForCall(0)
			Expect(chReq).To(Equal(channel.Request{
				ChaincodeID: fabproxy.QSCC,
				Fcn:         "GetTransactionByID",
				Args:        [][]byte{[]byte("getCode"), []byte(sampleAddress)},
			}))

			Expect(reqOpts).To(HaveLen(0))

			Expect(reply).To(Equal(string(sampleCode)))
		})

		Context("when getting the channel client errors ", func() {
			BeforeEach(func() {
				fabSDK.GetChannelClientReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				sampleAddress := "0x1234567123"
				var reply string

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleAddress, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to generate channel client"))

				Expect(reply).To(BeEmpty())
			})
		})

		// Need to test multiple places the querying can fail
		XContext("when querying the ledger errors", func() {
			BeforeEach(func() {
				mockChClient.QueryReturns(channel.Response{}, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				sampleAddress := "0x1234567123"
				var reply string

				err := ethservice.GetCode(&http.Request{}, &sampleAddress, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to query the ledger"))

				Expect(reply).To(BeEmpty())
			})
		})
	})
})
