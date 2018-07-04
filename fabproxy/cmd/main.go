package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-evm/fabproxy"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type fabSDK struct {
	sdk     *fabsdk.FabricSDK
	channel string
	user    string
	org     string
}

func (s *fabSDK) GetChannelClient() (fabproxy.ChannelClient, error) {
	clientChannelContext := s.sdk.ChannelContext(s.channel, fabsdk.WithUser(s.user), fabsdk.WithOrg(s.org))

	return channel.New(clientChannelContext)
}

func main() {
	cfg := os.Getenv("ETHSERVER_CONFIG")
	org := os.Getenv("ETHSERVER_ORG")
	user := os.Getenv("ETHSERVER_USER")
	channel := os.Getenv("ETHSERVER_CHANNEL")
	port := os.Getenv("PORT")

	var portNumber int
	if port != "" {
		var err error
		portNumber, err = strconv.Atoi(port)
		if err != nil {
			panic("Error converting value of environment variable PORT to int")
		}
	} else {
		portNumber = 5000
	}

	sdk, err := fabsdk.New(config.FromFile(cfg))
	if err != nil {
		panic(fmt.Sprintf("Failed to create Fabric SDK Clienti: %s", err.Error()))
		return
	}
	ethService := fabproxy.NewEthService(&fabSDK{sdk: sdk, channel: channel, user: user, org: org}, channel)

	proxy := fabproxy.NewFabProxy(ethService)
	proxy.Start(portNumber)
}
