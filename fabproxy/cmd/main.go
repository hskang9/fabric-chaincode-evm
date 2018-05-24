package main

import "github.com/hyperledger/fabric-chaincode-evm/fabproxy"

func main() {
	proxy := fabproxy.NewFabProxy()
	proxy.Start(5000)
}
