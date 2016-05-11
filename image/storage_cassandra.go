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

package image

import (
	"github.com/cloudawan/cloudone/utility/database/cassandra"
	"github.com/gocql/gocql"
	"time"
)

type StorageCassandra struct {
	dummyError DummyError
}

func (storageCassandra *StorageCassandra) initialize() error {
	storageCassandra.dummyError = DummyError{"Dummy support nothing"}

	tableSchemaImageInformation := `
	CREATE TABLE IF NOT EXISTS image_information (
	name varchar,
	kind varchar,
	description varchar,
	current_version varchar,
	build_parameter map<varchar, varchar>,
	PRIMARY KEY (name));
	`

	tableSchemaImageRecord := `
	CREATE TABLE IF NOT EXISTS image_record (
	image_information varchar,
	version varchar,
	path varchar,
	version_info map<varchar, varchar>,
	environment map<varchar, varchar>,
	description varchar,
	created_time timeuuid,
	PRIMARY KEY (image_information, version));
	`

	err := cassandra.CassandraClient.CreateTableIfNotExist(tableSchemaImageInformation, 3, time.Second*5)
	if err != nil {
		log.Critical("Fail to create table with schema %s", tableSchemaImageInformation)
		return err
	}
	err = cassandra.CassandraClient.CreateTableIfNotExist(tableSchemaImageRecord, 3, time.Second*5)
	if err != nil {
		log.Critical("Fail to create table with schema %s", tableSchemaImageRecord)
		return err
	}

	return nil
}

func (storageCassandra *StorageCassandra) DeleteImageInformationAndRelatedRecord(name string) error {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("DELETE FROM image_information WHERE name = ?", name).Exec(); err != nil {
		log.Error("Delete ImageInformation with name %s error: %s", name, err)
		return err
	}
	return storageCassandra.deleteImageRecordWithImageInformationName(name)
}

func (storageCassandra *StorageCassandra) saveImageInformation(imageInformation *ImageInformation) error {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("INSERT INTO image_information (name, kind, description, current_version, build_parameter) VALUES (?, ?, ?, ?, ?)",
		imageInformation.Name,
		imageInformation.Kind,
		imageInformation.Description,
		imageInformation.CurrentVersion,
		imageInformation.BuildParameter,
	).Exec(); err != nil {
		log.Error("Save ImageInformation %s error: %s", imageInformation, err)
		return err
	}
	return nil
}

func (storageCassandra *StorageCassandra) LoadImageInformation(name string) (*ImageInformation, error) {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return nil, err
	}
	imageInformation := new(ImageInformation)
	err = session.Query("SELECT name, kind, description, current_version, build_parameter FROM image_information WHERE name = ?", name).Scan(
		&imageInformation.Name,
		&imageInformation.Kind,
		&imageInformation.Description,
		&imageInformation.CurrentVersion,
		&imageInformation.BuildParameter,
	)
	if err != nil {
		log.Error("Load ImageInformation %s error: %s", name, err)
		return nil, err
	} else {
		return imageInformation, nil
	}
}

func (storageCassandra *StorageCassandra) LoadAllImageInformation() ([]ImageInformation, error) {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return nil, err
	}
	iter := session.Query("SELECT name, kind, description, current_version, build_parameter FROM image_information").Iter()

	imageInformationSlice := make([]ImageInformation, 0)
	imageInformation := new(ImageInformation)

	for iter.Scan(
		&imageInformation.Name,
		&imageInformation.Kind,
		&imageInformation.Description,
		&imageInformation.CurrentVersion,
		&imageInformation.BuildParameter,
	) {
		imageInformationSlice = append(imageInformationSlice, *imageInformation)
		imageInformation = new(ImageInformation)
	}

	err = iter.Close()
	if err != nil {
		log.Error("Load all ImageInformation error: %s", err)
		return nil, err
	} else {
		return imageInformationSlice, nil
	}
}

func (storageCassandra *StorageCassandra) DeleteImageRecord(imageInformationName string, version string) error {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("DELETE FROM image_record WHERE image_information = ? AND version = ?", imageInformationName, version).Exec(); err != nil {
		log.Error("Delete ImageRecord with image information name %s, version %s error: %s", imageInformationName, version, err)
		return err
	}
	return nil
}

func (storageCassandra *StorageCassandra) deleteImageRecordWithImageInformationName(imageInformationName string) error {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("DELETE FROM image_record WHERE image_information = ?", imageInformationName).Exec(); err != nil {
		log.Error("Delete ImageRecord with image information name %s error: %s", imageInformationName, err)
		return err
	}
	return nil
}

func (storageCassandra *StorageCassandra) saveImageRecord(imageRecord *ImageRecord) error {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("INSERT INTO image_record (image_information, version, path, version_info, environment, description, created_time) VALUES (?, ?, ?, ?, ?, ?, ?)",
		imageRecord.ImageInformation, imageRecord.Version, imageRecord.Path, imageRecord.VersionInfo, imageRecord.Environment, imageRecord.Description, gocql.UUIDFromTime(imageRecord.CreatedTime)).Exec(); err != nil {
		log.Error("Save ImageRecord %s error: %s", imageRecord, err)
		return err
	}
	return nil
}

func (storageCassandra *StorageCassandra) LoadImageRecord(imageInformationName string, version string) (*ImageRecord, error) {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return nil, err
	}
	imageRecord := new(ImageRecord)
	var uuid gocql.UUID
	err = session.Query("SELECT image_information, version, path, version_info, environment, description, created_time FROM image_record WHERE image_information = ? AND version = ?", imageInformationName, version).Scan(
		&imageRecord.ImageInformation,
		&imageRecord.Version,
		&imageRecord.Path,
		&imageRecord.VersionInfo,
		&imageRecord.Environment,
		&imageRecord.Description,
		&uuid,
	)
	if err != nil {
		log.Error("Load ImageRecord %s version %s error: %s", imageInformationName, version, err)
		return nil, err
	} else {
		imageRecord.CreatedTime = uuid.Time()
		return imageRecord, nil
	}
}

func (storageCassandra *StorageCassandra) LoadImageRecordWithImageInformationName(imageInformationName string) ([]ImageRecord, error) {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return nil, err
	}
	iter := session.Query("SELECT image_information, version, path, version_info, environment, description, created_time FROM image_record WHERE image_information = ?", imageInformationName).Iter()

	imageRecordSlice := make([]ImageRecord, 0)
	imageRecord := new(ImageRecord)
	var uuid gocql.UUID

	for iter.Scan(&imageRecord.ImageInformation, &imageRecord.Version, &imageRecord.Path, &imageRecord.VersionInfo, &imageRecord.Environment, &imageRecord.Description, &uuid) {
		imageRecord.CreatedTime = uuid.Time()
		imageRecordSlice = append(imageRecordSlice, *imageRecord)
		imageRecord = new(ImageRecord)
	}

	err = iter.Close()
	if err != nil {
		log.Error("Load ImageRecord %s error: %s", imageInformationName, err)
		return nil, err
	} else {
		return imageRecordSlice, nil
	}
}
