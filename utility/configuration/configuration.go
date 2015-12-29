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
	"github.com/cloudawan/cloudone/utility/logger"
	"github.com/cloudawan/cloudone_utility/configuration"
)

var log = logger.GetLogManager().GetLogger("utility")

var configurationContent = `
{
	"certificate": "/etc/cloudone/development_cert.pem",
	"key": "/etc/cloudone/development_key.pem",
	"cassandraClusterIp": ["127.0.0.1"],
	"cassandraClusterPort": 9042,
	"cassandraKeyspace": "cloudone",
	"cassandraReplicationStrategy": "{'class': 'SimpleStrategy', 'replication_factor' : 1}",
	"cassandraTimeoutInMilliSecond": 3000,
	"emailSenderAccount": "cloudawanemailtest@gmail.com",
	"emailSenderPassword": "cloudawan4test",
	"emailSenderHost": "smtp.gmail.com",
	"emailSenderPort": 587,
	"smsNexmoURL": "https://rest.nexmo.com/sms/json",
	"smsNexmoAPIKey": "2045d69e",
	"smsNexmoAPISecret": "fcaf0b59",
	"glusterfsHost": ["127.0.0.1"],
	"glusterfsPath": "/data/glusterfs",
	"glusterfsSSHDialTimeoutInMilliSecond": 1000,
	"glusterfsSSHSessionTimeoutInMilliSecond": 10000,
	"glusterfsSSHHost": ["127.0.0.1"],
	"glusterfsSSHPort": 22,
	"glusterfsSSHUser": "user",
	"glusterfsSSHPassword": "password",
	"etcdHostAndPort": ["127.0.0.1:4001"],
	"etcdCheckTimeoutInMilliSecond": 300
}
`

var LocalConfiguration *configuration.Configuration

func init() {
	var err error
	LocalConfiguration, err = configuration.CreateConfiguration("cloudone", configurationContent)
	if err != nil {
		log.Critical(err)
		panic(err)
	}
}
