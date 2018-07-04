package fabproxy

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
)

type FabProxy struct {
	server *rpc.Server
}

func NewFabProxy(service EthService) *FabProxy {
	server := rpc.NewServer()

	proxy := &FabProxy{
		server: server,
	}

	// sdk, err := fabsdk.New(config.FromFile("/Users/Repakula/go/src/github.com/hyperledger/fabric-chaincode-evm/fabproxy/cmd/config_test.yaml"))
	// if err != nil {
	// 	fmt.Println("SDK FAILED: ", err.Error())
	// }
	// ethService := NewEthService(&fabSDK{sdk: sdk}, "CHANGE ME")

	server.RegisterCodec(NewRPCCodec(), "application/json")
	server.RegisterService(service, "eth")

	return proxy
}

func (p *FabProxy) Start(port int) {
	r := mux.NewRouter()
	r.Handle("/", p.server)

	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
