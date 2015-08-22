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
	"code.google.com/p/log4go"
	"github.com/cloudawan/kubernetes_management/utility/logger"
	"github.com/cloudawan/kubernetes_management_utility/configuration"
)

var configurationContent = `
{
	"certificate": "/etc/kubernetes_management/development_cert.pem",
	"key": "/etc/kubernetes_management/development_key.pem",
	"cassandraClusterIp": ["127.0.0.1"],
	"cassandraClusterPort": 9042,
	"cassandraKeyspace": "kubernetes_management",
	"cassandraReplicationStrategy": "{'class': 'SimpleStrategy', 'replication_factor' : 1}",
	"kubeapiHost": "127.0.0.1",
	"kubeapiPort": 8080,
	"emailSenderAccount": "cloudawanemailtest@gmail.com",
	"emailSenderPassword": "cloudawan4test",
	"emailSenderHost": "smtp.gmail.com",
	"emailSenderPort": 587,
	"smsNexmoURL": "https://rest.nexmo.com/sms/json",
	"smsNexmoAPIKey": "2045d69e",
	"smsNexmoAPISecret": "fcaf0b59"
}
`

var log log4go.Logger = logger.GetLogger("utility")

var LocalConfiguration *configuration.Configuration

func init() {
	var err error
	LocalConfiguration, err = configuration.CreateConfiguration("kubernetes_management", configurationContent)
	if err != nil {
		log.Critical(err)
		panic(err)
	}
}
