package fabproxy

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

const EVMSCC = "evmscc"
const QSCC = "qscc"

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
	return &EthService{sdk: sdk}
}

func (s *EthService) GetCode(r *http.Request, arg *string, reply *string) error {
	chClient, err := s.sdk.GetChannelClient()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to generate channel client: %s", err.Error()))
	}

	response, err := s.query(chClient, EVMSCC, [][]byte{[]byte("getCode"), []byte(strip0xFromHex(*arg))})

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

	response, err := s.query(chClient, EVMSCC, [][]byte{[]byte(args.To), []byte(args.Data)})

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

	response, err := chClient.Execute(channel.Request{
		ChaincodeID: EVMSCC,
		Fcn:         "invoke",
		Args:        [][]byte{[]byte(args.To), []byte(args.Data)},
	})

	if err != nil {
		return errors.New(fmt.Sprintf("Failed to execute transaction: %s", err.Error()))
	}

	*reply = string(response.TransactionID)
	return nil
}

func (s *EthService) GetTransactionReceipt(r *http.Request, arg *string, reply *TxReceipt) error {
	return nil
}

func strip0xFromHex(addr string) string {
	stripped := strings.Split(addr, "0x")
	// if len(stripped) != 1 {
	// 	panic("Had more then 1 0x in address")
	// }
	return stripped[len(stripped)-1]
}

func (s *EthService) query(chClient ChannelClient, ccid string, queryArgs [][]byte) (channel.Response, error) {

	return chClient.Query(channel.Request{
		ChaincodeID: ccid,
		Fcn:         "invoke",
		Args:        queryArgs,
	})
}
