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
	"github.com/cloudawan/kubernetes_management/Godeps/_workspace/src/code.google.com/p/log4go"
	"github.com/cloudawan/kubernetes_management/Godeps/_workspace/src/github.com/gocql/gocql"
	"github.com/cloudawan/kubernetes_management/utility/logger"
	"time"
)

var log log4go.Logger = logger.GetLogger("utility")

type CassandraClient struct {
	session             *gocql.Session
	clusterIp           []string
	clusterPort         int
	keyspace            string
	replicationStrategy string
}

func CreateCassandraClient(clusterIp []string, clusterPort int,
	keyspace string, replicationStrategy string) *CassandraClient {
	cassandraClient := &CassandraClient{nil, clusterIp, clusterPort, keyspace, replicationStrategy}
	cassandraClient.GetSession()
	return cassandraClient
}

func (cassandraClient *CassandraClient) GetSession() *gocql.Session {
	if cassandraClient.session != nil {
		return cassandraClient.session
	} else {

		cluster := gocql.NewCluster(cassandraClient.clusterIp...)
		cluster.Port = cassandraClient.clusterPort
		session, err := cluster.CreateSession()
		if err != nil {
			log.Critical("Fail to create Cassandra session: %s", err)
			session = nil
			return nil
		} else {
			if err := session.Query("CREATE KEYSPACE IF NOT EXISTS " + cassandraClient.keyspace + " WITH replication = " + cassandraClient.replicationStrategy).Exec(); err != nil {
				log.Critical("Fail to check if not exist then create keyspace error: %s", err)
				session = nil
				return nil
			} else {
				session.Close()
				cluster.Keyspace = cassandraClient.keyspace
				session, err := cluster.CreateSession()
				if err != nil {
					log.Critical("Fail to create Cassandra session: %s", err)
					session = nil
					return nil
				} else {
					cassandraClient.session = session
					return cassandraClient.session
				}
			}
		}
	}
}

func (cassandraClient *CassandraClient) CloseSession() {
	cassandraClient.session.Close()
	cassandraClient.session = nil
}

func (cassandraClient *CassandraClient) ReloadConfiguration() {
	cassandraClient.CloseSession()
	cassandraClient.GetSession()
}

func (cassandraClient *CassandraClient) CreateTableIfNotExist(tableSchema string, retryAmount int, retryInterval time.Duration) error {
	session := cassandraClient.GetSession()
	var returnedError error = nil
	for i := 0; i < retryAmount; i++ {
		if err := session.Query(tableSchema).Exec(); err == nil {
			return nil
		} else {
			log.Error("Check if not exist then create table schema %s error: %s", tableSchema, err)
			returnedError = err
		}
		time.Sleep(retryInterval)
	}

	return returnedError
}
