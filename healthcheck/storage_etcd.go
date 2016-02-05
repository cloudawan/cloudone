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

package healthcheck

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"golang.org/x/net/context"
	"time"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/test"); err != nil {
		log.Error("Create if not existing test directory error: %s", err)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) saveTest(name string, updatedTime time.Time) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}
	jsonMap := make(map[string]interface{})
	jsonMap["name"] = name
	jsonMap["updatedTime"] = updatedTime

	byteSlice, err := json.Marshal(jsonMap)
	if err != nil {
		log.Error("Marshal test %v error %s", jsonMap, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/test/"+name, string(byteSlice), nil)
	if err != nil {
		log.Error("Save test %v error: %s", jsonMap, err)
		log.Error(response)
		return err
	}

	return nil
}
