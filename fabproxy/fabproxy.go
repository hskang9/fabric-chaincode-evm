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

func NewFabProxy() *FabProxy {
	server := rpc.NewServer()

	proxy := &FabProxy{
		server: server,
	}

	server.RegisterCodec(NewRPCCodec(), "application/json")
	server.RegisterService(EthService{}, "eth")

	return proxy
}

func (p *FabProxy) Start(port int) {
	r := mux.NewRouter()
	r.Handle("/", p.server)

	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
