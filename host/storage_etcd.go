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

package host

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"golang.org/x/net/context"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/host_credential"); err != nil {
		log.Error("Create if not existing host credential directory error: %s", err)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) DeleteCredential(name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/host_credential/"+name, nil)
	if err != nil {
		log.Error("Delete host credential with name %s error: %s", name, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) SaveCredential(credential *Credential) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(credential)
	if err != nil {
		log.Error("Marshal host credential %v error %s", credential, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/host_credential/"+credential.IP, string(byteSlice), nil)
	if err != nil {
		log.Error("Save host credential %v error: %s", credential, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadCredential(ip string) (*Credential, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/host_credential/"+ip, nil)
	if err != nil {
		log.Error("Load host credential with ip %s error: %s", ip, err)
		log.Error(response)
		return nil, err
	}

	credential := new(Credential)
	err = json.Unmarshal([]byte(response.Node.Value), &credential)
	if err != nil {
		log.Error("Unmarshal host credential %v error %s", response.Node.Value, err)
		return nil, err
	}

	return credential, nil
}

func (storageEtcd *StorageEtcd) LoadAllCredential() ([]Credential, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/host_credential", nil)
	if err != nil {
		log.Error("Load all host credential error: %s", err)
		log.Error(response)
		return nil, err
	}

	credentialSlice := make([]Credential, 0)
	for _, node := range response.Node.Nodes {
		credential := Credential{}
		err := json.Unmarshal([]byte(node.Value), &credential)
		if err != nil {
			log.Error("Unmarshal host credential %v error %s", node.Value, err)
			return nil, err
		}
		credentialSlice = append(credentialSlice, credential)
	}

	return credentialSlice, nil
}
