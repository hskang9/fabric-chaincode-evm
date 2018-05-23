package fabproxy

import (
	"errors"
	"fmt"
	"net/http"

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

func NewEthService(sdk SDK, channelID string) *EthService {
	return &EthService{sdk: sdk}
}

func (s *EthService) GetCode(r *http.Request, arg *string, reply *string) error {
	chClient, err := s.sdk.GetChannelClient()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to generate channel client: %s", err.Error()))
	}

	response, err := s.query(EVMSCC, chClient, "getCode", [][]byte{[]byte(*arg)})

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

	response, err := s.query(EVMSCC, chClient, args.To, [][]byte{[]byte(args.Data)})

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

func (s *EthService) query(chClient ChannelClient, ccid string, function string, queryArgs [][]byte) (channel.Response, error) {

	return chClient.Query(channel.Request{
		ChaincodeID: ccid,
		Fcn:         function,
		Args:        queryArgs,
	})
}
