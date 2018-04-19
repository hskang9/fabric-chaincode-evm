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
	"errors"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/fabric-chaincode-evm/mocks"
	"github.com/hyperledger/fabric-chaincode-evm/statemanager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Statemanager", func() {

	var (
		sm         statemanager.StateManager
		mockStub   *mocks.MockStub
		addr       account.Address
		fakeLedger map[string][]byte
	)

	BeforeEach(func() {
		mockStub = &mocks.MockStub{}
		sm = statemanager.NewStateManager(mockStub)

		var err error
		addr, err = account.AddressFromBytes([]byte("0000000000000address"))
		Expect(err).ToNot(HaveOccurred())
		fakeLedger = make(map[string][]byte)
		mockStub.PutStateStub = func(key string, value []byte) error {
			fakeLedger[key] = value
			return nil
		}

		mockStub.GetStateStub = func(key string) ([]byte, error) {
			return fakeLedger[key], nil
		}
	})

	Describe("GetAccount", func() {
		BeforeEach(func() {

		})
		It("returns the account associated with the address", func() {
			err := mockStub.PutState(addr.String(), []byte("account code"))

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

		Context("when GetState errors out", func() {
			BeforeEach(func() {
				mockStub.GetStateReturns(nil, errors.New("boom!"))
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
			err := mockStub.PutState(addr.String()+key.String(), expectedVal.Bytes())
			Expect(err).ToNot(HaveOccurred())

			val, err := sm.GetStorage(addr, key)
			Expect(err).ToNot(HaveOccurred())

			Expect(val).To(Equal(expectedVal))
		})

		Context("when GetState returns an error", func() {
			BeforeEach(func() {
				mockStub.GetStateReturns(nil, errors.New("boom!"))
			})

			It("returns an error", func() {
				val, err := sm.GetStorage(addr, key)
				Expect(err).To(HaveOccurred())

				Expect(val).To(Equal(binary.Word256{}))
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

				err := sm.UpdateAccount(expectedAcct)
				Expect(err).ToNot(HaveOccurred())

				code, err := mockStub.GetState(addr.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(code).To(Equal(initialCode))
			})
		})

		Context("when the account exists", func() {
			It("updates the account", func() {

				err := mockStub.PutState(addr.String(), initialCode)
				Expect(err).ToNot(HaveOccurred())

				updatedCode := []byte("updated account code")
				updatedAccount := account.ConcreteAccount{
					Address: addr,
					Code:    updatedCode,
				}.Account()

				err = sm.UpdateAccount(updatedAccount)
				Expect(err).ToNot(HaveOccurred())

				code, err := mockStub.GetState(addr.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(code).To(Equal(updatedCode))
			})
		})

		Context("when stub throws an error", func() {
			BeforeEach(func() {
				mockStub.PutStateReturns(errors.New("boom!"))
			})

			It("returns an error", func() {
				expectedAcct := account.ConcreteAccount{
					Address: addr,
					Code:    initialCode,
				}.Account()

				err := sm.UpdateAccount(expectedAcct)
				Expect(err).To(HaveOccurred())

				code, err := mockStub.GetState(addr.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(code).To(BeEmpty())
			})
		})
	})

	Describe("RemoveAccount", func() {
		BeforeEach(func() {
			mockStub.DelStateStub = func(key string) error {
				delete(fakeLedger, key)
				return nil
			}
		})
		Context("when the account existed previously", func() {
			It("removes the account", func() {
				err := mockStub.PutState(addr.String(), []byte("account code"))
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

		Context("when stub throws an error", func() {
			BeforeEach(func() {
				mockStub.DelStateReturns(errors.New("boom!"))
			})

			It("returns an error", func() {
				initialCode := []byte("account code")
				err := mockStub.PutState(addr.String(), initialCode)
				Expect(err).ToNot(HaveOccurred())

				err = sm.RemoveAccount(addr)
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
				err := mockStub.PutState(compKey, initialVal.Bytes())
				Expect(err).ToNot(HaveOccurred())

				updatedVal := binary.LeftPadWord256([]byte("updated-storage-value"))

				err = sm.SetStorage(addr, key, updatedVal)
				Expect(err).ToNot(HaveOccurred())

				val, err := mockStub.GetState(compKey)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(updatedVal.Bytes()))

				// Put state should only have been called from above call in the test.
				Expect(mockStub.PutStateCallCount()).To(Equal(1))
			})
		})

		Context("when the key does not exist", func() {
			It("creates the key value pair", func() {
				err := sm.SetStorage(addr, key, initialVal)
				Expect(err).ToNot(HaveOccurred())

				val, err := mockStub.GetState(compKey)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(initialVal.Bytes()))

				Expect(mockStub.PutStateCallCount()).To(Equal(0))
			})
		})

		Context("when stub throws an error", func() {
			BeforeEach(func() {
				mockStub.PutStateReturns(errors.New("boom!"))
			})

			It("returns an error", func() {
				err := sm.SetStorage(addr, key, initialVal)
				Expect(err).To(HaveOccurred())

				val, err := mockStub.GetState(compKey)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(BeEmpty())
			})
		})
	})

	Describe("Flush", func() {
		var key, key2, val binary.Word256

		Context("when SetStorage has previously been called", func() {
			BeforeEach(func() {
				val = binary.LeftPadWord256([]byte("storage-value"))
				key = binary.LeftPadWord256([]byte("key"))

				key2 = binary.LeftPadWord256([]byte("key2"))

				err := sm.SetStorage(addr, key, val)
				Expect(err).ToNot(HaveOccurred())
				err = sm.SetStorage(addr, key2, val)
				Expect(err).ToNot(HaveOccurred())

				Expect(mockStub.PutStateCallCount()).To(Equal(0))

				Expect(mockStub.GetState(addr.String() + key.String())).To(BeEmpty())
				Expect(mockStub.GetState(addr.String() + key2.String())).To(BeEmpty())
			})

			It("writes all the setStorage calls to the ledger", func() {
				sm.Flush()
				Expect(mockStub.PutStateCallCount()).To(Equal(2))

				Expect(mockStub.GetState(addr.String() + key.String())).To(Equal(val))
				Expect(mockStub.GetState(addr.String() + key2.String())).To(Equal(val))
			})

			Context("when SetStorage is called on the same key more than once", func() {
				var val2 binary.Word256
				BeforeEach(func() {
					val2 = binary.LeftPadWord256([]byte("storage-value"))

					err := sm.SetStorage(addr, key, val)
					Expect(err).ToNot(HaveOccurred())
				})

				It("writes the last value to the ledger", func() {
					sm.Flush()
					Expect(mockStub.PutStateCallCount()).To(Equal(2))

					Expect(mockStub.GetState(addr.String() + key.String())).To(Equal(val2))
					Expect(mockStub.GetState(addr.String() + key2.String())).To(Equal(val))
				})
			})
		})

		Context("when SetStorage has not been previously been called", func() {
			It("does not write anything to the ledger", func() {
				err := sm.Flush()
				Expect(err).ToNot(HaveOccurred())

				Expect(mockStub.PutStateCallCount()).To(Equal(0))
			})
		})
	})
})
