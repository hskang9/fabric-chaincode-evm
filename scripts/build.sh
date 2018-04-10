#!/usr/bin/env bash

set -e

FABRIC_PATH=$1
EVMSCC_PATH=$2

echo "installing govendor"
mkdir -p $GOPATH/bin
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh


cd $EVMSCC_PATH

echo "removing packages already vendored in fabric from evmscc"
comm -12 <(cd $EVMSCC_PATH && dep status| awk '{if(NR>1)print $1}' | sort) <(cd $FABRIC_PATH && dep status| awk '{if(NR>1)print $1}' | sort) | xargs -n1 -I{} rm -rf vendor/{} 

echo "copying packages needed by evmscc to fabric vendor"
cp -r $EVMSCC_PATH/vendor/* $FABRIC_PATH/vendor/ #&& rm -rf $EVMSCC_PATH/vendor

echo "copying evmscc to fabric plugin"
mkdir -p $FABRIC_PATH/plugin/evmscc && cp -r $EVMSCC_PATH/evmscc/evmscc.go $FABRIC_PATH/plugin/evmscc
mkdir -p $FABRIC_PATH/vendor/github.com/hyperledger/fabric-chaincode-evm
cp -r $EVMSCC_PATH/evmscc/mocks $FABRIC_PATH/vendor/github.com/hyperledger/fabric-chaincode-evm/.
cp -r $EVMSCC_PATH/evmscc/statemanager $FABRIC_PATH/vendor/github.com/hyperledger/fabric-chaincode-evm/.

echo "building evmscc plugin shared object to $LIB_DIR"
cd $FABRIC_PATH/plugin && go build -o $LIB_DIR/evmscc.so -buildmode=plugin ./evmscc

