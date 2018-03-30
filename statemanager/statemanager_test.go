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

package statemanager_test

import (
	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	. "github.com/hyperledger/fabric-chaincode-evm/statemanager"
	"github.com/hyperledger/fabric/core/chaincode/shim"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Statemanager", func() {

	var (
		sm       StateManager
		mockStub *shim.MockStub
		addr     account.Address
	)

	BeforeEach(func() {
		mockStub = shim.NewMockStub("mock", nil)
		sm = NewStateManager(mockStub)

		var err error
		addr, err = account.AddressFromBytes([]byte("0000000000000address"))
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("GetAccount", func() {
		It("returns the account associated with the address", func() {
			mockStub.MockTransactionStart("tx1")
			err := mockStub.PutState(addr.String(), []byte("account code"))
			mockStub.MockTransactionEnd("tx1")

			Expect(err).ToNot(HaveOccurred())

			expectedAcct := account.ConcreteAccount{
				Address: addr,
				Code:    []byte("account code"),
			}.Account()

			acct, err := sm.GetAccount(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(acct).To(Equal(expectedAcct))

		})

		Context("when no account exists", func() {
			It("returns an empty account", func() {
				acct, err := sm.GetAccount(addr)
				Expect(err).ToNot(HaveOccurred())

				Expect(acct).To(Equal(account.ConcreteAccount{}.Account()))
			})
		})

		//TODO: when mockStub errors out
		XContext("when GetState errors out", func() {
			BeforeEach(func() {
				//TODO: cause mockStub to error when doing GetState
			})
			It("returns an empty account and an error", func() {
				acct, err := sm.GetAccount(addr)
				Expect(err).To(HaveOccurred())

				Expect(acct).To(Equal(account.ConcreteAccount{}.Account()))
			})
		})
	})

	Describe("GetStorage", func() {
		var expectedVal, key binary.Word256
		BeforeEach(func() {

			expectedVal = binary.LeftPadWord256([]byte("storage-value"))
			key = binary.LeftPadWord256([]byte("key"))
		})

		It("returns the value associated with the key", func() {
			mockStub.MockTransactionStart("tx1")
			err := mockStub.PutState(addr.String()+key.String(), expectedVal.Bytes())
			mockStub.MockTransactionEnd("tx1")

			Expect(err).ToNot(HaveOccurred())

			val, err := sm.GetStorage(addr, key)
			Expect(err).ToNot(HaveOccurred())

			Expect(val).To(Equal(expectedVal))
		})

		XContext("when GetState returns an error", func() {
			BeforeEach(func() {
				//TODO: cause mockStub to error when doing GetState
			})

			It("returns an error", func() {
				val, err := sm.GetStorage(addr, key)
				Expect(err).ToNot(HaveOccurred())

				Expect(val).To(Equal(expectedVal))
			})
		})
	})

	Describe("UpdateAccount", func() {
		var initialCode []byte
		BeforeEach(func() {
			initialCode = []byte("account code")
		})

		Context("when the account didn't exist", func() {
			It("creates the account", func() {

				expectedAcct := account.ConcreteAccount{
					Address: addr,
					Code:    initialCode,
				}.Account()

				mockStub.MockTransactionStart("tx1")
				err := sm.UpdateAccount(expectedAcct)
				mockStub.MockTransactionStart("tx1")
				Expect(err).ToNot(HaveOccurred())

				code, err := mockStub.GetState(addr.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(code).To(Equal(initialCode))
			})
		})

		Context("when the account exists", func() {
			It("updates the account", func() {

				mockStub.MockTransactionStart("tx1")
				err := mockStub.PutState(addr.String(), initialCode)
				mockStub.MockTransactionEnd("tx1")
				Expect(err).ToNot(HaveOccurred())

				updatedCode := []byte("updated account code")
				updatedAccount := account.ConcreteAccount{
					Address: addr,
					Code:    updatedCode,
				}.Account()

				mockStub.MockTransactionStart("tx1")
				err = sm.UpdateAccount(updatedAccount)
				mockStub.MockTransactionStart("tx1")
				Expect(err).ToNot(HaveOccurred())

				code, err := mockStub.GetState(addr.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(code).To(Equal(updatedCode))
			})
		})

		XContext("when stub throws an error", func() {
			BeforeEach(func() {
				//TODO: cause mockStub to error when doing PutState
			})

			It("returns an error", func() {
				expectedAcct := account.ConcreteAccount{
					Address: addr,
					Code:    initialCode,
				}.Account()

				mockStub.MockTransactionStart("tx1")
				err := sm.UpdateAccount(expectedAcct)
				mockStub.MockTransactionStart("tx1")
				Expect(err).To(HaveOccurred())

				code, err := mockStub.GetState(addr.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(code).To(BeEmpty())
			})
		})
	})

	Describe("RemoveAccount", func() {
		Context("when the account existed previously", func() {
			It("removes the account", func() {
				mockStub.MockTransactionStart("tx1")
				err := mockStub.PutState(addr.String(), []byte("account code"))
				mockStub.MockTransactionEnd("tx1")
				Expect(err).ToNot(HaveOccurred())

				err = sm.RemoveAccount(addr)
				Expect(err).ToNot(HaveOccurred())

				code, err := mockStub.GetState(addr.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(code).To(BeEmpty())
			})
		})

		Context("when the accound did not exists previously", func() {
			It("does not return an error", func() {
				err := sm.RemoveAccount(addr)
				Expect(err).ToNot(HaveOccurred())

				code, err := mockStub.GetState(addr.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(code).To(BeEmpty())
			})
		})

		XContext("when stub throws an error", func() {
			BeforeEach(func() {
				//TODO: cause mockStub to error when doing DelState
			})

			It("returns an error", func() {
				initialCode := []byte("account code")
				expectedAcct := account.ConcreteAccount{
					Address: addr,
					Code:    initialCode,
				}.Account()

				mockStub.MockTransactionStart("tx1")
				err := sm.UpdateAccount(expectedAcct)
				mockStub.MockTransactionStart("tx1")
				Expect(err).To(HaveOccurred())

				code, err := mockStub.GetState(addr.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(code).To(Equal(initialCode))
			})
		})
	})

	Describe("SetStorage", func() {
		var (
			key, initialVal binary.Word256
			compKey         string
		)

		BeforeEach(func() {

			initialVal = binary.LeftPadWord256([]byte("storage-value"))
			key = binary.LeftPadWord256([]byte("key"))
			compKey = addr.String() + key.String()
		})

		Context("when key already exists", func() {
			It("updates the key value pair", func() {
				mockStub.MockTransactionStart("tx1")
				err := mockStub.PutState(compKey, initialVal.Bytes())
				mockStub.MockTransactionEnd("tx1")
				Expect(err).ToNot(HaveOccurred())

				updatedVal := binary.LeftPadWord256([]byte("updated-storage-value"))

				mockStub.MockTransactionStart("tx2")
				err = sm.SetStorage(addr, key, updatedVal)
				mockStub.MockTransactionEnd("tx2")
				Expect(err).ToNot(HaveOccurred())

				val, err := mockStub.GetState(compKey)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(updatedVal.Bytes()))
			})
		})

		Context("when the key does not exist", func() {
			It("creates the key value pair", func() {
				mockStub.MockTransactionStart("tx1")
				err := sm.SetStorage(addr, key, initialVal)
				mockStub.MockTransactionEnd("tx1")
				Expect(err).ToNot(HaveOccurred())

				val, err := mockStub.GetState(compKey)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(initialVal.Bytes()))
			})
		})

		XContext("when stub throws an error", func() {
			BeforeEach(func() {
				//TODO: cause mockStub to error when doing DelState
			})

			It("returns an error", func() {
				mockStub.MockTransactionStart("tx1")
				err := sm.SetStorage(addr, key, initialVal)
				mockStub.MockTransactionStart("tx1")
				Expect(err).To(HaveOccurred())

				val, err := mockStub.GetState(compKey)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(initialVal.Bytes()))
			})
		})
	})
})
