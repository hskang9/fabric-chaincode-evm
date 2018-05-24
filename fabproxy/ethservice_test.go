package fabproxy_test

import (
	"errors"
	"net/http"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-evm/fabproxy"
	"github.com/hyperledger/fabric-chaincode-evm/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
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
		ethservice = fabproxy.NewEthService(fabSDK)
	})

	//TODO: Fix the query args. Need to find out if fab sdk uses invoke or the "function" used by the chaincode as the function arg
	//TODO: Fix getTransactionReceipt tests

	Describe("GetCode", func() {
		var (
			sampleCode                          []byte
			sampleAddress, addressWithoutPrefix string
		)

		BeforeEach(func() {
			sampleCode = []byte("sample-code")
			mockChClient.QueryReturns(channel.Response{
				Payload: sampleCode,
			}, nil)

			sampleAddress = "0x1234567123"
			addressWithoutPrefix = "1234567123"
		})

		It("returns the code associated to that address", func() {
			var reply string

			err := ethservice.GetCode(&http.Request{}, &sampleAddress, &reply)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockChClient.QueryCallCount()).To(Equal(1))
			chReq, reqOpts := mockChClient.QueryArgsForCall(0)
			Expect(chReq).To(Equal(channel.Request{
				ChaincodeID: fabproxy.EVMSCC,
				Fcn:         "invoke",
				Args:        [][]byte{[]byte("getCode"), []byte(addressWithoutPrefix)},
			}))

			Expect(reqOpts).To(HaveLen(0))

			Expect(reply).To(Equal(string(sampleCode)))
		})

		Context("when the address is malformed", func() {
			BeforeEach(func() {
				sampleAddress = "0x0x12345"
			})

			It("returns an error for a malformed address", func() {
				var reply string

				err := ethservice.GetCode(&http.Request{}, &sampleAddress, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Recieved malformed address"))

				Expect(reply).To(BeEmpty())
			})
		})

		Context("when the address does not have `0x` prefix", func() {
			BeforeEach(func() {
				sampleAddress = "123456"
			})
			It("returns the code associated with that address", func() {
				var reply string

				err := ethservice.GetCode(&http.Request{}, &sampleAddress, &reply)
				Expect(err).ToNot(HaveOccurred())

				Expect(mockChClient.QueryCallCount()).To(Equal(1))
				chReq, reqOpts := mockChClient.QueryArgsForCall(0)
				Expect(chReq).To(Equal(channel.Request{
					ChaincodeID: fabproxy.EVMSCC,
					Fcn:         "invoke",
					Args:        [][]byte{[]byte("getCode"), []byte(sampleAddress)},
				}))

				Expect(reqOpts).To(HaveLen(0))

				Expect(reply).To(Equal(string(sampleCode)))
			})
		})

		Context("when getting the channel client errors ", func() {
			BeforeEach(func() {
				fabSDK.GetChannelClientReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
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
				var reply string

				err := ethservice.GetCode(&http.Request{}, &sampleAddress, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to query the ledger"))

				Expect(reply).To(BeEmpty())
			})
		})
	})

	//TODO: Add tests for Prefix
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

	//TODO: Add tests for Prefix
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

	// TODO: Add tests for Prefix
	FDescribe("GetTransactionReceipt", func() {
		var (
			sampleResponse      channel.Response
			sampleTransactionID string
		)

		BeforeEach(func() {
			sampleResponse = channel.Response{}
			mockChClient.QueryStub = func(request channel.Request, options ...channel.RequestOption) (channel.Response, error) {

				if string(request.Args[0]) == "GetTransactionById" {
					sampleTx, err := GetSampleTransaction([][]byte{[]byte("sample arg 1"), []byte("sample arg 2")}, []byte("sample-response"))
					Expect(err).ToNot(HaveOccurred())

					txBytes, err := proto.Marshal(&sampleTx)
					Expect(err).ToNot(HaveOccurred())
					sampleResponse.Payload = txBytes

				} else if string(request.Args[0]) == "GetBlockByTxID" {

					sampleTx, err := GetSampleBlock(1, []byte("12345abcd"))
					Expect(err).ToNot(HaveOccurred())

					txBytes, err := proto.Marshal(&sampleTx)
					Expect(err).ToNot(HaveOccurred())
					sampleResponse.Payload = txBytes
				}

				return sampleResponse, errors.New("boom!")
			}

			sampleTransactionID = "0x1234567123"
		})

		It("returns the transaction receipt associated to that transaction address", func() {
			var reply fabproxy.TxReceipt

			err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockChClient.QueryCallCount()).To(Equal(1))
			chReq, reqOpts := mockChClient.QueryArgsForCall(0)
			Expect(chReq).To(Equal(channel.Request{
				ChaincodeID: fabproxy.QSCC,
				Fcn:         "invoke",
				Args:        [][]byte{[]byte("GetTransactionByID"), []byte(sampleTransactionID)},
			}))

			Expect(reqOpts).To(HaveLen(0))

			Expect(reply).To(Equal(fabproxy.TxReceipt{
				TransactionHash:   sampleTransactionID,
				BlockHash:         "",
				BlockNumber:       "",
				GasUsed:           0,
				CumulativeGasUsed: 0,
			}))

		})

		Context("when getting the channel client errors ", func() {
			BeforeEach(func() {
				fabSDK.GetChannelClientReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply fabproxy.TxReceipt

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
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
				var reply fabproxy.TxReceipt

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to query the ledger"))

				Expect(reply).To(BeEmpty())
			})
		})
	})
})

func GetSampleBlock(blkNumber uint64, blkHash []byte) (common.Block, error) {
	blk := common.Block{
		Header: &common.BlockHeader{Number: blkNumber, DataHash: blkHash},
	}

	return blk, nil
}

func GetSampleTransaction(inputArgs [][]byte, txResponse []byte) (peer.ProcessedTransaction, error) {

	respPayload := &peer.ChaincodeAction{
		Response: &peer.Response{
			Payload: txResponse,
		},
	}

	ext, err := proto.Marshal(respPayload)
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	pRespPayload := &peer.ProposalResponsePayload{
		Extension: ext,
	}

	ccProposalPayload, err := proto.Marshal(pRespPayload)
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	invokeSpec := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			ChaincodeId: &peer.ChaincodeID{
				Name: fabproxy.EVMSCC,
			},
			Input: &peer.ChaincodeInput{
				Args: inputArgs,
			},
		},
	}

	invokeSpecBytes, err := proto.Marshal(invokeSpec)
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	ccPropPayload, err := proto.Marshal(&peer.ChaincodeProposalPayload{
		Input: invokeSpecBytes,
	})
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	ccPayload := &peer.ChaincodeActionPayload{
		Action: &peer.ChaincodeEndorsedAction{
			ProposalResponsePayload: ccProposalPayload,
		},
		ChaincodeProposalPayload: ccPropPayload,
	}

	actionPayload, err := proto.Marshal(ccPayload)
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	txAction := &peer.TransactionAction{
		Payload: actionPayload,
	}

	txActions := &peer.Transaction{
		Actions: []*peer.TransactionAction{txAction},
	}

	actionsPayload, err := proto.Marshal(txActions)
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	payload := &common.Payload{
		Data: actionsPayload,
	}

	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	tx := peer.ProcessedTransaction{
		TransactionEnvelope: &common.Envelope{
			Payload: payloadBytes,
		},
	}

	return tx, nil
}
