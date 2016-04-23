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

package deploy

import (
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

func (storageDummy *StorageDummy) DeleteDeployInformation(namespace string, imageInformation string) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) saveDeployInformation(deployInformation *DeployInformation) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadDeployInformation(namespace string, imageInformation string) (*DeployInformation, error) {
	return nil, &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadAllDeployInformation() ([]DeployInformation, error) {
	return nil, &storageDummy.dummyError
}

func (storageDummy *StorageDummy) DeleteDeployBlueGreen(imageInformation string) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) saveDeployBlueGreen(deployBlueGreen *DeployBlueGreen) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadDeployBlueGreen(imageInformation string) (*DeployBlueGreen, error) {
	return nil, &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadAllDeployBlueGreen() ([]DeployBlueGreen, error) {
	return nil, &storageDummy.dummyError
}

func (storageDummy *StorageDummy) DeleteDeployClusterApplication(namespace string, name string) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) SaveDeployClusterApplication(deployClusterApplication *DeployClusterApplication) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadDeployClusterApplication(namespace string, name string) (*DeployClusterApplication, error) {
	return nil, &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadAllDeployClusterApplication() ([]DeployClusterApplication, error) {
	return nil, &storageDummy.dummyError
}
