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

package configuration

import (
	"errors"
	"github.com/cloudawan/cloudone/utility/logger"
	"github.com/cloudawan/cloudone_utility/configuration"
)

var log = logger.GetLogManager().GetLogger("utility")

var configurationContent = `
{
	"certificate": "/etc/cloudone/development_cert.pem",
	"key": "/etc/cloudone/development_key.pem",
	"restapiPort": 8081,
	"emailSenderAccount": "cloudawanemailtest@gmail.com",
	"emailSenderPassword": "cloudawan4test",
	"emailSenderHost": "smtp.gmail.com",
	"emailSenderPort": 587,
	"smsNexmoURL": "https://rest.nexmo.com/sms/json",
	"smsNexmoAPIKey": "2045d69e",
	"smsNexmoAPISecret": "fcaf0b59",
	"etcdEndpoints": ["http://127.0.0.1:4001"],
	"etcdHeaderTimeoutPerRequestInMilliSecond": 2000,
	"etcdBasePath": "/cloudawan/cloudone",
	"storageTypeDefault": 3
}
`

var LocalConfiguration *configuration.Configuration

func init() {
	err := Reload()
	if err != nil {
		log.Critical(err)
		panic(err)
	}
}

func Reload() error {
	localConfiguration, err := configuration.CreateConfiguration("cloudone", configurationContent)
	if err == nil {
		LocalConfiguration = localConfiguration
	}

	return err
}

const (
	StorageTypeDefault   = 0
	StorageTypeDummy     = 1
	StorageTypeCassandra = 2
	StorageTypeEtcd      = 3
)

func GetStorageTypeDefault() (int, error) {
	value, ok := LocalConfiguration.GetInt("storageTypeDefault")
	if ok == false {
		log.Critical("Can't load storageTypeDefault")
		return 0, errors.New("Can't load storageTypeDefault")
	}
	return value, nil
}
