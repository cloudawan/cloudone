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
	"errors"
	"github.com/cloudawan/cloudone/utility/configuration"
)

const (
	StorageTypeDefault   = 0
	StorageTypeDummy     = 1
	StorageTypeCassandra = 2
)

var storage Storage = nil

func GetStorage() Storage {
	switch storage.(type) {
	case nil:
		if err := ReloadStorage(StorageTypeDefault); err != nil {
			log.Error(err)
			log.Critical("Fail to load storage and use dummy")
			if err := ReloadStorage(StorageTypeDummy); err != nil {
				log.Error(err)
			}
		}
	case *StorageDummy:
		// If dummy, will retry to use default storage every configured interval
		if storage.(*StorageDummy).ShouldCheck() {
			// If fail to reload, it will use the previous one
			if err := ReloadStorage(StorageTypeDefault); err != nil {
				log.Error(err)
			}
		}
	}

	return storage
}

func ReloadStorage(storageType int) error {
	switch storageType {
	default:
		return errors.New("Not supported type")
	case StorageTypeDefault:
		// If not indicated, use default
		storageTypeDefault, err := getStorageTypeDefault()
		if err != nil {
			log.Error(err)
			return ReloadStorage(StorageTypeDummy)
		} else {
			return ReloadStorage(storageTypeDefault)
		}
	case StorageTypeDummy:
		newStorage := &StorageDummy{}
		err := newStorage.initialize()
		if err == nil {
			storage = newStorage
		}
		return err
	case StorageTypeCassandra:
		newStorage := &StorageCassandra{}
		err := newStorage.initialize()
		if err == nil {
			storage = newStorage
		}
		return err
	}
}

func getStorageTypeDefault() (int, error) {
	value, ok := configuration.LocalConfiguration.GetInt("storageTypeDefault")
	if ok == false {
		log.Critical("Can't load storageTypeDefault")
		return 0, errors.New("Can't load storageTypeDefault")
	}
	return value, nil
}

type Storage interface {
	initialize() error
	DeleteImageInformationAndRelatedRecord(name string) error
	saveImageInformation(imageInformation *ImageInformation) error
	LoadImageInformation(name string) (*ImageInformation, error)
	LoadAllImageInformation() ([]ImageInformation, error)
	saveImageRecord(imageRecord *ImageRecord) error
	DeleteImageRecord(imageInformationName string, version string) error
	deleteImageRecordWithImageInformationName(imageInformationName string) error
	LoadImageRecord(imageInformationName string, version string) (*ImageRecord, error)
	LoadImageRecordWithImageInformationName(imageInformationName string) ([]ImageRecord, error)
}
