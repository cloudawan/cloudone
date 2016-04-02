// Copyright 2015 CloudAwan LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authorization

import (
	"github.com/cloudawan/cloudone_utility/rbac"
	"time"
)

type DummyError struct {
	text string
}

func (dummyError *DummyError) Error() string {
	return dummyError.text
}

var defaultCheckInterval time.Duration = time.Minute

type StorageDummy struct {
	dummyError    DummyError
	lastCheckTime time.Time
	checkInterval time.Duration
}

func (storageDummy *StorageDummy) ShouldCheck() bool {
	if time.Now().Sub(storageDummy.lastCheckTime) > storageDummy.checkInterval {
		storageDummy.lastCheckTime = time.Now()
		return true
	} else {
		return false
	}
}

func (storageDummy *StorageDummy) initialize() error {
	storageDummy.dummyError = DummyError{"Dummy support nothing"}
	storageDummy.lastCheckTime = time.Now()
	storageDummy.checkInterval = defaultCheckInterval
	return nil
}

func (storageDummy *StorageDummy) DeleteUser(name string) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) SaveUser(user *rbac.User) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadUser(name string) (*rbac.User, error) {
	return nil, &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadAllUser() ([]rbac.User, error) {
	return nil, &storageDummy.dummyError
}

func (storageDummy *StorageDummy) DeleteRole(name string) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) SaveRole(role *rbac.Role) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadRole(name string) (*rbac.Role, error) {
	return nil, &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadAllRole() ([]rbac.Role, error) {
	return nil, &storageDummy.dummyError
}
