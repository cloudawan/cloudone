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

package etcd

import (
	"errors"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/cloudawan/cloudone/utility/logger"
	"github.com/cloudawan/cloudone_utility/database/etcd"
	"time"
)

var log = logger.GetLogManager().GetLogger("utility")

var EtcdClient *etcd.EtcdClient

func init() {
	storageTypeDefault, err := configuration.GetStorageTypeDefault()
	if err != nil {
		log.Critical("Unable to load the storage type so it can't decide to use this or not with error %s", err)
		panic(err)
	}
	if storageTypeDefault == configuration.StorageTypeEtcd {
		err := Reload()
		if err != nil {
			panic(err)
		}
	}
}

func Reload() error {
	etcdEndpointSlice, ok := configuration.LocalConfiguration.GetStringSlice("etcdEndpoints")
	if ok == false {
		log.Critical("Can't load etcdEndpoints")
		return errors.New("Can't load etcdEndpoints")
	}

	etcdHeaderTimeoutPerRequestInMilliSecond, ok := configuration.LocalConfiguration.GetInt("etcdHeaderTimeoutPerRequestInMilliSecond")
	if ok == false {
		log.Critical("Can't load etcdHeaderTimeoutPerRequestInMilliSecond")
		return errors.New("Can't load etcdHeaderTimeoutPerRequestInMilliSecond")
	}

	etcdBasePath, ok := configuration.LocalConfiguration.GetString("etcdBasePath")
	if ok == false {
		log.Critical("Can't load etcdBasePath")
		return errors.New("Can't load etcdBasePath")
	}

	EtcdClient = etcd.CreateEtcdClient(
		etcdEndpointSlice,
		time.Millisecond*time.Duration(etcdHeaderTimeoutPerRequestInMilliSecond),
		etcdBasePath)

	return nil
}
