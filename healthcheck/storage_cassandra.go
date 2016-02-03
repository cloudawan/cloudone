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
	"github.com/cloudawan/cloudone/utility/database/cassandra"
	"github.com/gocql/gocql"
	"time"
)

type StorageCassandra struct {
}

func (storageCassandra *StorageCassandra) initialize() error {
	tableSchemaTest := `
	CREATE TABLE IF NOT EXISTS test (
	name varchar,
	updated_time timeuuid,
	PRIMARY KEY (name));
	`

	err := cassandra.CassandraClient.CreateTableIfNotExist(tableSchemaTest, 3, time.Second*5)
	if err != nil {
		log.Critical("Fail to create table with schema %s", tableSchemaTest)
		return err
	}

	return nil
}

func (storageCassandra *StorageCassandra) saveTest(name string, updatedTime time.Time) error {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("INSERT INTO test (name, updated_time) VALUES (?, ?)",
		name,
		gocql.UUIDFromTime(updatedTime),
	).Exec(); err != nil {
		log.Error("Save Test name %s updatedTime %s error: %s", name, updatedTime, err)
		return err
	}
	return nil
}
