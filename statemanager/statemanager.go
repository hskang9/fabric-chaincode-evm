/*
Copyright IBM Corp. 2016 All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
		 http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package statemanager

import (
	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type StateManager interface {
	GetAccount(address account.Address) (account.Account, error)
	GetStorage(address account.Address, key binary.Word256) (binary.Word256, error)
	UpdateAccount(updatedAccount account.Account) error
	RemoveAccount(address account.Address) error
	SetStorage(address account.Address, key, value binary.Word256) error
}

type stateManager struct {
	stub         shim.ChaincodeStubInterface
	cache        map[string]binary.Word256
	acctCache    map[string][]byte
	deletedAccts map[string]byte
}

func NewStateManager(stub shim.ChaincodeStubInterface) StateManager {
	return &stateManager{
		stub:         stub,
		cache:        make(map[string]binary.Word256),
		acctCache:    make(map[string][]byte),
		deletedAccts: make(map[string]byte),
	}
}

func (s *stateManager) GetAccount(address account.Address) (account.Account, error) {
	var code []byte

	var ok bool
	if _, ok = s.deletedAccts[address.String()]; ok {
		return account.ConcreteAccount{}.Account(), nil
	}

	if code, ok = s.acctCache[address.String()]; !ok {
		var err error
		code, err = s.stub.GetState(address.String())
		if err != nil {
			return account.ConcreteAccount{}.Account(), err
		}
	}

	if len(code) == 0 {
		return account.ConcreteAccount{}.Account(), nil
	}

	return account.ConcreteAccount{
		Address: address,
		Code:    code,
	}.Account(), nil
}

func (s *stateManager) GetStorage(address account.Address, key binary.Word256) (binary.Word256, error) {
	compKey := address.String() + key.String()

	if val, ok := s.cache[compKey]; ok {
		return val, nil
	}

	val, err := s.stub.GetState(compKey)
	if err != nil {
		return binary.Word256{}, err
	}

	return binary.LeftPadWord256(val), nil
}

func (s *stateManager) UpdateAccount(updatedAccount account.Account) error {
	var err error
	if err = s.stub.PutState(updatedAccount.Address().String(), updatedAccount.Code().Bytes()); err == nil {
		s.acctCache[updatedAccount.Address().String()] = updatedAccount.Code().Bytes()
	}

	if _, ok := s.deletedAccts[updatedAccount.Address().String()]; ok {
		delete(s.deletedAccts, updatedAccount.Address().String())
	}

	return err

}

//What happens you delete account and you try to reference it afterwards?
func (s *stateManager) RemoveAccount(address account.Address) error {
	var err error
	if err = s.stub.DelState(address.String()); err == nil {
		s.deletedAccts[address.String()] = byte('1')
	}
	return err
}

func (s *stateManager) SetStorage(address account.Address, key, value binary.Word256) error {

	var err error
	if err = s.stub.PutState(address.String()+key.String(), value.Bytes()); err == nil {

		s.cache[address.String()+key.String()] = value
	}

	return err
}
