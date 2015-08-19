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

package application

import (
	"github.com/cloudawan/kubernetes_management/utility/database/cassandra"
	"time"
)

var tableSchemaStatelessApplication = `
	CREATE TABLE IF NOT EXISTS stateless_application (
	name varchar,
	description varchar,
	replication_controller_json blob,
	service_json blob,
	environment map<varchar, varchar>,
	PRIMARY KEY (name));
	`

func init() {
	err := cassandra.CassandraClient.CreateTableIfNotExist(tableSchemaStatelessApplication, 3, time.Second*3)
	if err != nil {
		panic(err)
	}
}

func DeleteStatelessApplication(name string) error {
	session := cassandra.CassandraClient.GetSession()
	if err := session.Query("DELETE FROM stateless_application WHERE name = ?", name).Exec(); err != nil {
		log.Error("Delete stateless application with name %s error: %s", name, err)
		return err
	}
	return nil
}

func saveStatelessApplication(stateless *Stateless) error {
	session := cassandra.CassandraClient.GetSession()
	if err := session.Query("INSERT INTO stateless_application (name, description, replication_controller_json, service_json, environment) VALUES (?, ?, ?, ?, ?)",
		stateless.Name,
		stateless.Description,
		stateless.replicationControllerJson,
		stateless.serviceJson,
		stateless.Environment,
	).Exec(); err != nil {
		log.Error("Save stateless application %s error: %s", stateless, err)
		return err
	}
	return nil
}

func LoadStatelessApplication(name string) (*Stateless, error) {
	stateless := new(Stateless)

	session := cassandra.CassandraClient.GetSession()
	err := session.Query("SELECT name, description, replication_controller_json, service_json, environment FROM stateless_application WHERE name = ?", name).Scan(
		&stateless.Name,
		&stateless.Description,
		&stateless.replicationControllerJson,
		&stateless.serviceJson,
		&stateless.Environment,
	)
	if err != nil {
		return nil, err
	}

	return stateless, nil
}

func LoadAllStatelessApplication() ([]Stateless, error) {
	session := cassandra.CassandraClient.GetSession()
	iter := session.Query("SELECT name, description, replication_controller_json, service_json, environment FROM stateless_application").Iter()

	statelessSlice := make([]Stateless, 0)
	stateless := new(Stateless)

	for iter.Scan(
		&stateless.Name,
		&stateless.Description,
		&stateless.replicationControllerJson,
		&stateless.serviceJson,
		&stateless.Environment,
	) {
		statelessSlice = append(statelessSlice, *stateless)
		stateless = new(Stateless)
	}

	err := iter.Close()
	if err != nil {
		return nil, err
	}

	return statelessSlice, nil
}
