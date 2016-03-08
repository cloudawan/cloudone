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

package deploy

import (
	"github.com/cloudawan/cloudone/utility/database/cassandra"
	"time"
)

type StorageCassandra struct {
}

func (storageCassandra *StorageCassandra) initialize() error {
	tableSchemaDeployInformation := `
	CREATE TABLE IF NOT EXISTS deploy_information (
	namespace  varchar,
	image_information varchar,
	current_version varchar,
	current_version_description varchar,
	description varchar,
	PRIMARY KEY (namespace, image_information));
	`

	tableSchemaDeployBlueGreen := `
	CREATE TABLE IF NOT EXISTS deploy_blue_green (
	image_information varchar,
	current_namespace varchar,
	node_port int,
	description varchar,
	session_affinity varchar,
	PRIMARY KEY (image_information));
	`

	err := cassandra.CassandraClient.CreateTableIfNotExist(tableSchemaDeployInformation, 3, time.Second*5)
	if err != nil {
		log.Critical("Fail to create table with schema %s", tableSchemaDeployInformation)
		return err
	}
	err = cassandra.CassandraClient.CreateTableIfNotExist(tableSchemaDeployBlueGreen, 3, time.Second*5)
	if err != nil {
		log.Critical("Fail to create table with schema %s", tableSchemaDeployBlueGreen)
		return err
	}
	return nil
}

func (storageCassandra *StorageCassandra) DeleteDeployInformation(namespace string, imageInformation string) error {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("DELETE FROM deploy_information WHERE namespace = ? AND image_information = ?",
		namespace, imageInformation).Exec(); err != nil {
		log.Error("Delete DeployInformation with namespace %s imageInformation %s error: %s", namespace, imageInformation, err)
		return err
	}
	return nil
}

func (storageCassandra *StorageCassandra) saveDeployInformation(deployInformation *DeployInformation) error {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("INSERT INTO deploy_information (namespace, image_information, current_version, current_version_description, description) VALUES (?, ?, ?, ?, ?)",
		deployInformation.Namespace,
		deployInformation.ImageInformationName,
		deployInformation.CurrentVersion,
		deployInformation.Description,
	).Exec(); err != nil {
		log.Error("Save DeployInformation %s error: %s", deployInformation, err)
		return err
	}
	return nil
}

func (storageCassandra *StorageCassandra) LoadDeployInformation(namespace string, imageInformation string) (*DeployInformation, error) {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return nil, err
	}
	deployInformation := new(DeployInformation)
	err = session.Query("SELECT namespace, image_information, current_version, current_version_description, description FROM deploy_information WHERE namespace = ? AND image_information = ?", namespace, imageInformation).Scan(
		&deployInformation.Namespace,
		&deployInformation.ImageInformationName,
		&deployInformation.CurrentVersion,
		&deployInformation.CurrentVersionDescription,
		&deployInformation.Description,
	)
	if err != nil {
		log.Error("Load DeployInformation namespace %s imageInformation %s error: %s", namespace, imageInformation, err)
		return nil, err
	} else {
		return deployInformation, nil
	}
}

func (storageCassandra *StorageCassandra) LoadAllDeployInformation() ([]DeployInformation, error) {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return nil, err
	}
	iter := session.Query("SELECT namespace, image_information, current_version, current_version_description, description FROM deploy_information").Iter()

	deployInformationSlice := make([]DeployInformation, 0)
	deployInformation := new(DeployInformation)

	for iter.Scan(
		&deployInformation.Namespace,
		&deployInformation.ImageInformationName,
		&deployInformation.CurrentVersion,
		&deployInformation.CurrentVersionDescription,
		&deployInformation.Description,
	) {
		deployInformationSlice = append(deployInformationSlice, *deployInformation)
		deployInformation = new(DeployInformation)
	}

	err = iter.Close()
	if err != nil {
		log.Error("Load all DeployInformation error: %s", err)
		return nil, err
	} else {
		return deployInformationSlice, nil
	}
}

func (storageCassandra *StorageCassandra) DeleteDeployBlueGreen(imageInformation string) error {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("DELETE FROM deploy_blue_green WHERE image_information = ?",
		imageInformation).Exec(); err != nil {
		log.Error("Delete DeleteDeployBlueGreen with imageInformation %s error: %s", imageInformation, err)
		return err
	}
	return nil
}

func (storageCassandra *StorageCassandra) saveDeployBlueGreen(deployBlueGreen *DeployBlueGreen) error {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("INSERT INTO deploy_blue_green ( image_information, current_namespace, node_port, description, session_affinity) VALUES (?, ?, ?, ?, ?)",
		deployBlueGreen.ImageInformation,
		deployBlueGreen.Namespace,
		deployBlueGreen.NodePort,
		deployBlueGreen.Description,
		deployBlueGreen.SessionAffinity,
	).Exec(); err != nil {
		log.Error("Save DeployBlueGreen %s error: %s", deployBlueGreen, err)
		return err
	}
	return nil
}

func (storageCassandra *StorageCassandra) LoadDeployBlueGreen(imageInformation string) (*DeployBlueGreen, error) {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return nil, err
	}
	deployBlueGreen := new(DeployBlueGreen)
	err = session.Query("SELECT image_information, current_namespace, node_port, description, session_affinity FROM deploy_blue_green WHERE image_information = ?", imageInformation).Scan(
		&deployBlueGreen.ImageInformation,
		&deployBlueGreen.Namespace,
		&deployBlueGreen.NodePort,
		&deployBlueGreen.Description,
		&deployBlueGreen.SessionAffinity,
	)
	if err != nil {
		log.Error("Load DeployBlueGreen imageInformation %s error: %s", imageInformation, err)
		return nil, err
	} else {
		return deployBlueGreen, nil
	}
}

func (storageCassandra *StorageCassandra) LoadAllDeployBlueGreen() ([]DeployBlueGreen, error) {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return nil, err
	}
	iter := session.Query("SELECT image_information, current_namespace, node_port, description, session_affinity FROM deploy_blue_green").Iter()

	deployBlueGreenSlice := make([]DeployBlueGreen, 0)
	deployBlueGreen := new(DeployBlueGreen)

	for iter.Scan(
		&deployBlueGreen.ImageInformation,
		&deployBlueGreen.Namespace,
		&deployBlueGreen.NodePort,
		&deployBlueGreen.Description,
		&deployBlueGreen.SessionAffinity,
	) {
		deployBlueGreenSlice = append(deployBlueGreenSlice, *deployBlueGreen)
		deployBlueGreen = new(DeployBlueGreen)
	}

	err = iter.Close()
	if err != nil {
		log.Error("Load all DeployBlueGreen error: %s", err)
		return nil, err
	} else {
		return deployBlueGreenSlice, nil
	}
}
