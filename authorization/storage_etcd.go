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
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"github.com/cloudawan/cloudone_utility/rbac"
	"golang.org/x/net/context"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/user"); err != nil {
		log.Error("Create if not existing user directory error: %s", err)
		return err
	}

	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/role"); err != nil {
		log.Error("Create if not existing role directory error: %s", err)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) DeleteUser(name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/user/"+name, nil)
	if err != nil {
		log.Error("Delete user with name %s error: %s", name, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) SaveUser(user *rbac.User) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	// User doesn't save role data but only role name
	savedRoleSlice := make([]*rbac.Role, 0)
	for _, role := range user.RoleSlice {
		savedRole := rbac.Role{}
		savedRole.Name = role.Name
		savedRoleSlice = append(savedRoleSlice, &savedRole)
	}
	savedUser := rbac.User{
		user.Name,
		user.EncodedPassword,
		savedRoleSlice,
		user.ResourceSlice,
		user.Description,
		user.MetaDataMap,
		user.ExpiredTime,
		user.Disabled,
	}

	byteSlice, err := json.Marshal(savedUser)
	if err != nil {
		log.Error("Marshal user %v error %s", user, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/user/"+user.Name, string(byteSlice), nil)
	if err != nil {
		log.Error("Save user %v error: %s", user, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadUser(name string) (*rbac.User, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/user/"+name, nil)
	if err != nil {
		log.Error("Load name with name %s error: %s", name, err)
		log.Error(response)
		return nil, err
	}

	user := new(rbac.User)
	err = json.Unmarshal([]byte(response.Node.Value), &user)
	if err != nil {
		log.Error("Unmarshal user %v error %s", response.Node.Value, err)
		return nil, err
	}

	// User doesn't have role data but only role name so load role
	roleSlice := make([]*rbac.Role, 0)
	for _, savedRole := range user.RoleSlice {
		role, err := storageEtcd.LoadRole(savedRole.Name)
		if err != nil {
			log.Error("Fail to load role with name %s error %s", savedRole.Name, err)
		} else {
			roleSlice = append(roleSlice, role)
		}
	}
	user.RoleSlice = roleSlice

	return user, nil
}

func (storageEtcd *StorageEtcd) LoadAllUser() ([]rbac.User, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/user", nil)
	if err != nil {
		log.Error("Load all user error: %s", err)
		log.Error(response)
		return nil, err
	}

	userSlice := make([]rbac.User, 0)
	for _, node := range response.Node.Nodes {
		user := rbac.User{}
		err := json.Unmarshal([]byte(node.Value), &user)
		if err != nil {
			log.Error("Unmarshal user %v error %s", node.Value, err)
			return nil, err
		}

		// User doesn't have role data but only role name so load role
		roleSlice := make([]*rbac.Role, 0)
		for _, savedRole := range user.RoleSlice {
			role, err := storageEtcd.LoadRole(savedRole.Name)
			if err != nil {
				log.Error("Fail to load role with name %s error %s", savedRole.Name, err)
			} else {
				roleSlice = append(roleSlice, role)
			}
		}
		user.RoleSlice = roleSlice

		userSlice = append(userSlice, user)
	}

	return userSlice, nil
}

func (storageEtcd *StorageEtcd) DeleteRole(name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/role/"+name, nil)
	if err != nil {
		log.Error("Delete role with name %s error: %s", name, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) SaveRole(role *rbac.Role) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(role)
	if err != nil {
		log.Error("Marshal role %v error %s", role, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/role/"+role.Name, string(byteSlice), nil)
	if err != nil {
		log.Error("Save role %v error: %s", role, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadRole(name string) (*rbac.Role, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/role/"+name, nil)
	if err != nil {
		log.Error("Load name with name %s error: %s", name, err)
		log.Error(response)
		return nil, err
	}

	role := new(rbac.Role)
	err = json.Unmarshal([]byte(response.Node.Value), &role)
	if err != nil {
		log.Error("Unmarshal role %v error %s", response.Node.Value, err)
		return nil, err
	}

	return role, nil
}

func (storageEtcd *StorageEtcd) LoadAllRole() ([]rbac.Role, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/role", nil)
	if err != nil {
		log.Error("Load all role error: %s", err)
		log.Error(response)
		return nil, err
	}

	roleSlice := make([]rbac.Role, 0)
	for _, node := range response.Node.Nodes {
		role := rbac.Role{}
		err := json.Unmarshal([]byte(node.Value), &role)
		if err != nil {
			log.Error("Unmarshal role %v error %s", node.Value, err)
			return nil, err
		}
		roleSlice = append(roleSlice, role)
	}

	return roleSlice, nil
}
