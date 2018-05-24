package fabproxy

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type FabProxy struct {
	server *rpc.Server
}

func NewFabProxy() *FabProxy {
	server := rpc.NewServer()

	proxy := &FabProxy{
		server: server,
	}

	sdk, err := fabsdk.New(config.FromFile("/Users/Repakula/workspace/swe-cluster/swe-cluster.yml"))
	if err != nil {
		fmt.Println("SDK FAILED: ", err.Error())
	}
	ethService := NewEthService(&fabSDK{sdk: sdk})

	server.RegisterCodec(NewRPCCodec(), "application/json")
	server.RegisterService(ethService, "eth")

	return proxy
}

func (p *FabProxy) Start(port int) {
	r := mux.NewRouter()
	r.Handle("/", p.server)

	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}

type fabSDK struct {
	sdk *fabsdk.FabricSDK
}

func (s *fabSDK) GetChannelClient() (ChannelClient, error) {
	clientChannelContext := s.sdk.ChannelContext("channel1", fabsdk.WithUser("User1"), fabsdk.WithOrg("Org1"))

	return channel.New(clientChannelContext)
}
