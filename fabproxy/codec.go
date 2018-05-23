package fabproxy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
)

type rpcCodec struct {
	codec *json2.Codec
}

type rpcCodecRequest struct {
	rpc.CodecRequest
}

func NewRPCCodec() rpc.Codec {
	return &rpcCodec{codec: json2.NewCodec()}
}

func (c *rpcCodec) NewRequest(r *http.Request) rpc.CodecRequest {
	return &rpcCodecRequest{c.codec.NewRequest(r)}
}

func (r *rpcCodecRequest) Method() (string, error) {
	m, err := r.CodecRequest.Method()
	if err != nil {
		return "", err
	}
	method := strings.Split(m, "_")
	modifiedMethod := fmt.Sprintf("%s.%s", method[0], strings.Title(method[1]))
	return modifiedMethod, nil
}
