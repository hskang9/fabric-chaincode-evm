/*
Copyright IBM Corp All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric/integration/world"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestEndToEnd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EndToEnd Suite")
}

var (
	components        *world.Components
	testDir           string
	newGopath         string
	evmPluginFilePath string
	peerFilePath      string
)

var _ = SynchronizedBeforeSuite(func() []byte {
	components = &world.Components{}
	components.Build()
	// build(components)

	payload, err := json.Marshal(components)
	Expect(err).NotTo(HaveOccurred())

	return payload
}, func(payload []byte) {
	err := json.Unmarshal(payload, &components)
	Expect(err).NotTo(HaveOccurred())

	newGopath, err = ioutil.TempDir("", "newGopath")
	Expect(err).NotTo(HaveOccurred())

	err = os.Mkdir(filepath.Join(newGopath, "src"), 0755)
	Expect(err).ToNot(HaveOccurred())
	evmPluginFilePath, peerFilePath = compilePlugin("evmscc")

	components.Paths["peer"] = peerFilePath
})

var _ = BeforeEach(func() {
	var err error
	testDir, err = ioutil.TempDir("", "e2e-suite")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	os.RemoveAll(testDir)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	components.Cleanup()
	os.Remove(evmPluginFilePath)
	os.Remove(peerFilePath)
	os.RemoveAll(newGopath)
})

func compilePlugin(pluginType string) (string, string) {
	fabricPath := filepath.Join(newGopath, "src", "fabric")
	cmd := exec.Command("git", "clone", "https://github.com/hyperledger/fabric", fabricPath)
	err := cmd.Run()
	Expect(err).ToNot(HaveOccurred())

	cmd = exec.Command("git", "checkout", "master")
	cmd.Dir = fabricPath
	err = cmd.Run()
	Expect(err).ToNot(HaveOccurred())

	cmd = exec.Command("go", "list", "-f", "'{{ .Deps }}'", "./peer/")
	cmd.Dir = fabricPath
	output, err := cmd.Output()
	fabDeps := getDeps(output)

	cmd = exec.Command("go", "list", "-f", "'{{ .Deps }}'", "../plugin")
	output, err = cmd.Output()
	pluginDeps := getDeps(output)

	pluginOnlyDeps := diffDeps(fabDeps, pluginDeps)
	fmt.Println("PLUGIN ONLY DEPS ARE: ", pluginOnlyDeps)

	os.Mkdir(filepath.Join(fabricPath, "plugin"), 0755)

	files, err := ioutil.ReadDir("../plugin")
	Expect(err).ToNot(HaveOccurred())
	for _, f := range files {
		if strings.Contains(f.Name(), "test") {
			copyFile(f.Name(), filepath.Join(fabricPath, "plugin", f.Name()))
		}
	}

	extraConstraints := `
	[[constraint]]
	  name = "github.com/hyperledger/burrow"
		version - "0.18.0"
	`

	gopkg, err := os.OpenFile(filepath.Join(fabricPath, "Gopkg.toml"), os.O_APPEND|os.O_WRONLY, 0600)
	Expect(err).ToNot(HaveOccurred())

	defer gopkg.Close()

	_, err = gopkg.WriteString(extraConstraints)
	Expect(err).ToNot(HaveOccurred())
	gopkg.Close()

	pluginFilePath := filepath.Join("testdata", "plugin.so")

	cmd = exec.Command("dep", "ensure")
	cmd.Dir = fabricPath
	err = cmd.Run()
	Expect(err).ToNot(HaveOccurred())

	cmd = exec.Command(
		"go", "build", "-buildmode=plugin", "plugin",
	)
	cmd.Dir = fabricPath
	cmd.Run()

	Expect(pluginFilePath).To(BeARegularFile())

	peerFilePath := filepath.Join("testdata", "peer")
	cmd = exec.Command("go", "build", "-o", peerFilePath, "peer")
	cmd.Dir = fabricPath

	return pluginFilePath, peerFilePath
}

func getDeps(output []byte) []string {
	strippedOutput := strings.Replace(string(output), "[", "", 1)
	strippedOutput = strings.Replace(strippedOutput, "]", "", 1)
	return strings.Split(strippedOutput, " ")
}

func diffDeps(fabDeps []string, pluginDeps []string) []string {
	depsMap := map[string]byte{}

	for _, d := range pluginDeps {
		strippedDep := strings.Split(d, "github.com/hyperledger/fabric-chaincode-evm/")
		dep := strippedDep[len(strippedDep)-1]

		if strings.Contains(dep, "vendor") && !strings.Contains(dep, "github.com/hyperledger/fabric") {
			depsMap[dep] = '0'
			fmt.Println(dep)
		}

	}

	for _, d := range fabDeps {
		strippedDep := strings.Split(d, "github.com/hyperledger/fabric/")
		dep := strippedDep[len(strippedDep)-1]

		if _, ok := depsMap[dep]; ok {
			delete(depsMap, dep)
		}
	}

	pluginOnlyDeps := []string{}
	for k, _ := range depsMap {
		pluginOnlyDeps = append(pluginOnlyDeps, k)
	}
	return pluginOnlyDeps
}

func build(c *world.Components) {
	c.Paths = map[string]string{}

	cryptogen, err := gexec.Build("github.com/hyperledger/fabric/common/tools/cryptogen")
	Expect(err).NotTo(HaveOccurred())
	c.Paths["cryptogen"] = cryptogen

	idemixgen, err := gexec.Build("github.com/hyperledger/fabric/common/tools/idemixgen")
	Expect(err).NotTo(HaveOccurred())
	c.Paths["idemixgen"] = idemixgen

	configtxgen, err := gexec.Build("github.com/hyperledger/fabric/common/tools/configtxgen")
	Expect(err).NotTo(HaveOccurred())
	c.Paths["configtxgen"] = configtxgen

	orderer, err := gexec.Build("github.com/hyperledger/fabric/orderer")
	Expect(err).NotTo(HaveOccurred())
	c.Paths["orderer"] = orderer

	args := []string{"-tags", "pluginsenabled"}
	peer, err := gexec.Build("github.com/hyperledger/fabric/peer", args...)
	Expect(err).NotTo(HaveOccurred())
	c.Paths["peer"] = peer
}

func copyFile(src, dest string) {
	data, err := ioutil.ReadFile(src)
	Expect(err).NotTo(HaveOccurred())
	err = ioutil.WriteFile(dest, data, 0775)
	Expect(err).NotTo(HaveOccurred())
}

func copyPeerConfigs(peerOrgs []world.PeerOrgConfig, rootPath string) {
	for _, peerOrg := range peerOrgs {
		for peer := 0; peer < peerOrg.PeerCount; peer++ {
			peerDir := fmt.Sprintf("peer%d.%s", peer, peerOrg.Domain)
			if _, err := os.Stat(filepath.Join(rootPath, peerDir)); os.IsNotExist(err) {
				err := os.Mkdir(filepath.Join(rootPath, peerDir), 0755)
				Expect(err).NotTo(HaveOccurred())
			}
			copyFile(
				filepath.Join("testdata", fmt.Sprintf("%s_%d-core.yaml", peerOrg.Domain, peer)),
				filepath.Join(rootPath, peerDir, "core.yaml"),
			)
		}
	}
}
