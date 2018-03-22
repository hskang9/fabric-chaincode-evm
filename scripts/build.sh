#!/usr/bin/env bash

set -e

FABRIC_PATH=$1
EVMSCC_PATH=$2

echo "installing govendor"
go get -u github.com/kardianos/govendor

cd $EVMSCC_PATH

echo "removing packages already vendored in fabric from evmscc"
comm -12 <(cd $EVMSCC_PATH && govendor list +v +l | awk '{print $2}' | sort) <(cd $FABRIC_PATH && govendor list +v +l | awk '{print $2}' | sort) | xargs govendor remove

echo "copying packages needed by evmscc to fabric vendor"
cp -r $EVMSCC_PATH/vendor/* $FABRIC_PATH/vendor/ && rm -rf $EVMSCC_PATH/vendor

echo "copying evmscc to fabric plugin"
mkdir $FABRIC_PATH/plugin && cp -r $EVMSCC_PATH/* $FABRIC_PATH/plugin/

echo "building evmscc plugin shared object to $LIB_DIR"
cd $FABRIC_PATH/plugin && go build -o $LIB_DIR/evmscc.so -buildmode=plugin ./evmscc
