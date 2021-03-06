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

package notification

import (
	"errors"
	"github.com/cloudawan/cloudone/utility/configuration"
)

var storage Storage = nil

func GetStorage() Storage {
	switch storage.(type) {
	case nil:
		if err := ReloadStorage(configuration.StorageTypeDefault); err != nil {
			log.Error(err)
			log.Critical("Fail to load storage and use dummy")
			if err := ReloadStorage(configuration.StorageTypeDummy); err != nil {
				log.Error(err)
			}
		}
	case *StorageDummy:
		// If dummy, will retry to use default storage every configured interval
		if storage.(*StorageDummy).ShouldCheck() {
			// If fail to reload, it will use the previous one
			if err := ReloadStorage(configuration.StorageTypeDefault); err != nil {
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
	case configuration.StorageTypeDefault:
		// If not indicated, use default
		storageTypeDefault, err := configuration.GetStorageTypeDefault()
		if err != nil {
			log.Error(err)
			return ReloadStorage(configuration.StorageTypeDummy)
		} else {
			return ReloadStorage(storageTypeDefault)
		}
	case configuration.StorageTypeDummy:
		newStorage := &StorageDummy{}
		err := newStorage.initialize()
		if err == nil {
			storage = newStorage
		}
		return err
	case configuration.StorageTypeCassandra:
		newStorage := &StorageCassandra{}
		err := newStorage.initialize()
		if err == nil {
			storage = newStorage
		}
		return err
	case configuration.StorageTypeEtcd:
		newStorage := &StorageEtcd{}
		err := newStorage.initialize()
		if err == nil {
			storage = newStorage
		}
		return err
	}
}

type Storage interface {
	initialize() error
	DeleteReplicationControllerNotifierSerializable(namespace string, kind string, name string) error
	SaveReplicationControllerNotifierSerializable(replicationControllerNotifierSerializable *ReplicationControllerNotifierSerializable) error
	LoadReplicationControllerNotifierSerializable(namespace string, kind string, name string) (*ReplicationControllerNotifierSerializable, error)
	LoadAllReplicationControllerNotifierSerializable() ([]ReplicationControllerNotifierSerializable, error)
	DeleteEmailServerSMTP(name string) error
	SaveEmailServerSMTP(emailServerSMTP *EmailServerSMTP) error
	LoadEmailServerSMTP(name string) (*EmailServerSMTP, error)
	LoadAllEmailServerSMTP() ([]EmailServerSMTP, error)
	DeleteSMSNexmo(name string) error
	SaveSMSNexmo(sMSNexmo *SMSNexmo) error
	LoadSMSNexmo(name string) (*SMSNexmo, error)
	LoadAllSMSNexmo() ([]SMSNexmo, error)
}
