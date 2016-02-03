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

package cassandra

import (
	"errors"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/cloudawan/cloudone/utility/logger"
	"github.com/cloudawan/cloudone_utility/database/cassandra"
	"time"
)

var log = logger.GetLogManager().GetLogger("utility")

var CassandraClient *cassandra.CassandraClient

func init() {
	err := Reload()
	if err != nil {
		panic(err)
	}
}

func Reload() error {
	cassandraClusterIp, ok := configuration.LocalConfiguration.GetStringSlice("cassandraClusterIp")
	if ok == false {
		log.Critical("Can't load cassandraClusterIp")
		return errors.New("Can't load cassandraClusterIp")
	}

	cassandraClusterPort, ok := configuration.LocalConfiguration.GetInt("cassandraClusterPort")
	if ok == false {
		log.Critical("Can't load cassandraClusterPort")
		return errors.New("Can't load cassandraClusterPort")
	}

	cassandraKeyspace, ok := configuration.LocalConfiguration.GetString("cassandraKeyspace")
	if ok == false {
		log.Critical("Can't load cassandraKeyspace")
		return errors.New("Can't load cassandraKeyspace")
	}

	cassandraReplicationStrategy, ok := configuration.LocalConfiguration.GetString("cassandraReplicationStrategy")
	if ok == false {
		log.Critical("Can't load cassandraReplicationStrategy")
		return errors.New("Can't load cassandraReplicationStrategy")
	}

	cassandraTimeoutInMilliSecond, ok := configuration.LocalConfiguration.GetInt("cassandraTimeoutInMilliSecond")
	if ok == false {
		log.Critical("Can't load cassandraTimeoutInMilliSecond")
		return errors.New("Can't load cassandraTimeoutInMilliSecond")
	}

	CassandraClient = cassandra.CreateCassandraClient(
		cassandraClusterIp,
		cassandraClusterPort,
		cassandraKeyspace,
		cassandraReplicationStrategy,
		time.Millisecond*time.Duration(cassandraTimeoutInMilliSecond))

	return nil
}
