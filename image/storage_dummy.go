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

package image

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

func (storageDummy *StorageDummy) DeleteImageInformationAndRelatedRecord(name string) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) saveImageInformation(imageInformation *ImageInformation) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadImageInformation(name string) (*ImageInformation, error) {
	return nil, &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadAllImageInformation() ([]ImageInformation, error) {
	return nil, &storageDummy.dummyError
}

func (storageDummy *StorageDummy) saveImageRecord(imageRecord *ImageRecord) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) DeleteImageRecord(imageInformationName string, version string) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) deleteImageRecordWithImageInformationName(imageInformationName string) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadImageRecord(imageInformationName string, version string) (*ImageRecord, error) {
	return nil, &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadImageRecordWithImageInformationName(imageInformationName string) ([]ImageRecord, error) {
	return nil, &storageDummy.dummyError
}

func (storageDummy *StorageDummy) DeleteImageInformationBuildLock(name string) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) saveImageInformationBuildLock(imageInformationBuildLock *ImageInformationBuildLock) error {
	return &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadImageInformationBuildLock(imageInformation string) (*ImageInformationBuildLock, error) {
	return nil, &storageDummy.dummyError
}

func (storageDummy *StorageDummy) LoadAllImageInformationBuildLock() ([]ImageInformationBuildLock, error) {
	return nil, &storageDummy.dummyError
}
