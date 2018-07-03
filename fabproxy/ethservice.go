package fabproxy

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
)

const EVMSCC = "evmscc"
const QSCC = "qscc"

var ZeroAddress = make([]byte, 20)

type SDK interface {
	// ChannelContext(channelID string, options ...fabsdk.Option) contextApi.ChannelProvider
	GetChannelClient() (ChannelClient, error)
}

type ChannelClient interface {
	Query(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
	Execute(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
}

type EthService struct {
	sdk       SDK
	channelID string
}

type EthArgs struct {
	To       string
	From     string
	Gas      string
	GasPrice string
	Value    string
	Data     string
	Nonce    string
}

type TxReceipt struct {
	TransactionHash   string
	BlockHash         string
	BlockNumber       string
	ContractAddress   string
	GasUsed           int
	CumulativeGasUsed int
}

func NewEthService(sdk SDK, channelID string) *EthService {
	return &EthService{sdk: sdk, channelID: channelID}
}

func (s *EthService) GetCode(r *http.Request, arg *string, reply *string) error {
	chClient, err := s.sdk.GetChannelClient()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to generate channel client: %s", err.Error()))
	}

	strippedAddr := strip0xFromHex(*arg)

	response, err := s.query(chClient, EVMSCC, "getCode", [][]byte{[]byte(strippedAddr)})

	if err != nil {
		return errors.New(fmt.Sprintf("Failed to query the ledger: %s", err.Error()))
	}

	*reply = string(response.Payload)

	return nil
}

func (s *EthService) Call(r *http.Request, args *EthArgs, reply *string) error {
	chClient, err := s.sdk.GetChannelClient()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to generate channel client: %s", err.Error()))
	}

	strippedAddr := strip0xFromHex(args.To)

	response, err := s.query(chClient, EVMSCC, strippedAddr, [][]byte{[]byte(args.Data)})

	if err != nil {
		return errors.New(fmt.Sprintf("Failed to query the ledger: %s", err.Error()))
	}

	*reply = string(response.Payload)

	return nil
}

func (s *EthService) SendTransaction(r *http.Request, args *EthArgs, reply *string) error {
	chClient, err := s.sdk.GetChannelClient()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to generate channel client: %s", err.Error()))
	}

	strippedAddr := strip0xFromHex(args.To)
	response, err := chClient.Execute(channel.Request{
		ChaincodeID: EVMSCC,
		Fcn:         strippedAddr,
		Args:        [][]byte{[]byte(args.Data)},
	})

	if err != nil {
		return errors.New(fmt.Sprintf("Failed to execute transaction: %s", err.Error()))
	}

	*reply = string(response.TransactionID)
	return nil
}

func (s *EthService) GetTransactionReceipt(r *http.Request, arg *string, reply *TxReceipt) error {
	chClient, err := s.sdk.GetChannelClient()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to generate channel client: %s", err.Error()))
	}

	strippedTxId := strip0xFromHex(*arg)

	args := [][]byte{[]byte(s.channelID), []byte(strippedTxId)}

	t, err := s.query(chClient, "qscc", "GetTransactionByID", args)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to query the ledger: %s", err.Error()))
	}

	tx := &peer.ProcessedTransaction{}
	err = proto.Unmarshal(t.Payload, tx)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal transaction: %s", err.Error()))
	}

	p := tx.GetTransactionEnvelope().GetPayload()
	payload := &common.Payload{}
	err = proto.Unmarshal(p, payload)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal transaction: %s", err.Error()))
	}

	txActions := &peer.Transaction{}
	err = proto.Unmarshal(payload.GetData(), txActions)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal transaction: %s", err.Error()))
	}

	ccPropPayload, respPayload, err := getPayloads(txActions.GetActions()[0])
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal transaction: %s", err.Error()))
	}

	invokeSpec := &peer.ChaincodeInvocationSpec{}
	err = proto.Unmarshal(ccPropPayload.GetInput(), invokeSpec)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal transaction: %s", err.Error()))
	}

	b, err := s.query(chClient, "qscc", "GetBlockByTxID", args)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to query the ledger: %s", err.Error()))
	}
	block := &common.Block{}
	err = proto.Unmarshal(b.Payload, block)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal block: %s", err.Error()))
	}

	blkHeader := block.GetHeader()

	receipt := TxReceipt{
		TransactionHash:   *arg,
		BlockHash:         hex.EncodeToString(blkHeader.Hash()),
		BlockNumber:       strconv.FormatUint(blkHeader.GetNumber(), 10),
		GasUsed:           0,
		CumulativeGasUsed: 0,
	}

	args = invokeSpec.GetChaincodeSpec().GetInput().Args
	// First arg is the callee address. If it is zero address, tx was a contract creation
	callee, err := hex.DecodeString(string(args[0]))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to decode transaction arguments: %s", err.Error()))
	}

	if bytes.Equal(callee, ZeroAddress) {
		receipt.ContractAddress = string(respPayload.GetResponse().GetPayload())
	}
	*reply = receipt

	return nil
}

func (s *EthService) Accounts(r *http.Request, reply *[]string) error {
	chClient, err := s.sdk.GetChannelClient()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to generate channel client: %s", err.Error()))
	}

	response, err := s.query(chClient, EVMSCC, "account", [][]byte{})
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to query the ledger: %s", err.Error()))
	}

	*reply = []string{"0x" + strings.ToLower(string(response.Payload))}

	return nil
}

func strip0xFromHex(addr string) string {
	//Not checking for malformed addresses just stripping `0x` prefix where applicable
	if len(addr) > 2 && addr[0:2] == "0x" {
		return addr[2:]
	}
	return addr
}

func (s *EthService) query(chClient ChannelClient, ccid, function string, queryArgs [][]byte) (channel.Response, error) {

	return chClient.Query(channel.Request{
		ChaincodeID: ccid,
		Fcn:         function,
		Args:        queryArgs,
	})
}

func getPayloads(txActions *peer.TransactionAction) (*peer.ChaincodeProposalPayload, *peer.ChaincodeAction, error) {
	// TODO: pass in the tx type (in what follows we're assuming the type is ENDORSER_TRANSACTION)
	ccPayload := &peer.ChaincodeActionPayload{}
	err := proto.Unmarshal(txActions.Payload, ccPayload)
	if err != nil {
		return nil, nil, err
	}

	if ccPayload.Action == nil || ccPayload.Action.ProposalResponsePayload == nil {
		return nil, nil, fmt.Errorf("no payload in ChaincodeActionPayload")
	}

	ccProposalPayload := &peer.ChaincodeProposalPayload{}
	err = proto.Unmarshal(ccPayload.ChaincodeProposalPayload, ccProposalPayload)
	if err != nil {
		return nil, nil, err
	}

	pRespPayload := &peer.ProposalResponsePayload{}
	err = proto.Unmarshal(ccPayload.Action.ProposalResponsePayload, pRespPayload)
	if err != nil {
		return nil, nil, err
	}

	if pRespPayload.Extension == nil {
		return nil, nil, fmt.Errorf("response payload is missing extension")
	}

	respPayload := &peer.ChaincodeAction{}
	err = proto.Unmarshal(pRespPayload.Extension, respPayload)
	if err != nil {
		return ccProposalPayload, nil, err
	}
	return ccProposalPayload, respPayload, nil
}
