## Hyperledger Fabric Shim for node.js chaincodes

This is the project for the fabric chaincode shim for the Burrow EVM.

Please see the draft and evolving design document in [FAB-6590](https://jira.hyperledger.org/browse/FAB-6590).

<a rel="license" href="http://creativecommons.org/licenses/by/4.0/"><img alt="Creative Commons License" style="border-width:0" src="https://i.creativecommons.org/l/by/4.0/88x31.png" /></a><br />This work is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">Creative Commons Attribution 4.0 International License</a>

### Steps to build artifacts and run an e2e example

*prerequisite:*
* clone Fabric to $GOPATH/src/github.com/hyperledger/fabric
* clone this project to $GOPATH/src/github.com/hyperledger/fabric-chaincode-evm
* `hyperledger/fabric-tools` and `hyperledger/fabric-orderer` docker images

#### Build Fabric peer image
First, we need to patch `fabric/images/peer/Dockerfile.in` with one extra line:
```dockerfile
RUN apt-get update && apt-get install -y libltdl-dev
```
This is to install `libltdl` which is needed by evmscc plugin

Then, go to fabric root directory and build peer docker image with following command:
```bash
GO_TAGS=pluginsenabled EXPERIMENTAL=false DOCKER_DYNAMIC_LINK=true make peer-docker
```

#### Build `evmscc` shared object
execute this at fabric-chaincode-evm root directory:
```bash
make evmscc-linux
```
`evmscc.so` for Linux will be produced to `build/linux/lib`

#### Run e2e example
go to `e2e_cli` and run `./network_setup.sh up`