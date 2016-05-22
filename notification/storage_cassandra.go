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

package notification

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/cassandra"
	"time"
)

type StorageCassandra struct {
	dummyError DummyError
}

func (storageCassandra *StorageCassandra) initialize() error {
	storageCassandra.dummyError = DummyError{"Dummy support nothing"}

	tableSchemaNotifier := `
	CREATE TABLE IF NOT EXISTS notifier (
	check boolean,
	cool_down_duration bigint,
	remaining_cool_down bigint,
	kubeapi_host varchar,
	kubeapi_port int,
	namespace varchar,
	kind varchar,
	name varchar,
	notifier_slice blob,
	indicator_slice blob,
	PRIMARY KEY (namespace, kind, name));
	`

	err := cassandra.CassandraClient.CreateTableIfNotExist(tableSchemaNotifier, 3, time.Second*5)
	if err != nil {
		log.Critical("Fail to create table with schema %s", tableSchemaNotifier)
		return err
	}

	return nil
}

func (storageCassandra *StorageCassandra) DeleteReplicationControllerNotifierSerializable(namespace string, kind string, name string) error {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("DELETE FROM notifier WHERE namespace = ? AND kind = ? AND name = ?", namespace, kind, name).Exec(); err != nil {
		log.Error("Delete notifier with namespace %s kind %s name %s error: %s", namespace, kind, name, err)
		return err
	}
	return nil
}

func (storageCassandra *StorageCassandra) SaveReplicationControllerNotifierSerializable(replicationControllerNotifierSerializable *ReplicationControllerNotifierSerializable) error {
	notifierSliceByteSlice, err := json.Marshal(replicationControllerNotifierSerializable.NotifierSlice)
	if err != nil {
		log.Error("Marshal notifier slice error replicationControllerNotifierSerializable %s error: %s", replicationControllerNotifierSerializable, err)
		return err
	}
	indicatorSliceByteSlice, err := json.Marshal(replicationControllerNotifierSerializable.IndicatorSlice)
	if err != nil {
		log.Error("Marshal indicator slice error replicationControllerNotifierSerializable %s error: %s", replicationControllerNotifierSerializable, err)
		return err
	}

	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("INSERT INTO notifier (check, cool_down_duration, remaining_cool_down, kubeapi_host, kubeapi_port, namespace, kind, name, notifier_slice, indicator_slice) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		replicationControllerNotifierSerializable.Check,
		replicationControllerNotifierSerializable.CoolDownDuration,
		replicationControllerNotifierSerializable.RemainingCoolDown,
		replicationControllerNotifierSerializable.KubeApiServerEndPoint,
		replicationControllerNotifierSerializable.KubeApiServerToken,
		replicationControllerNotifierSerializable.Namespace,
		replicationControllerNotifierSerializable.Kind,
		replicationControllerNotifierSerializable.Name,
		notifierSliceByteSlice,
		indicatorSliceByteSlice,
	).Exec(); err != nil {
		log.Error("Save replicationControllerNotifierSerializable %s error: %s", replicationControllerNotifierSerializable, err)
		return err
	}
	return nil
}

func (storageCassandra *StorageCassandra) LoadReplicationControllerNotifierSerializable(namespace string, kind string, name string) (*ReplicationControllerNotifierSerializable, error) {
	notifierSliceByteSlice := make([]byte, 0)
	indicatorSliceByteSlice := make([]byte, 0)
	replicationControllerNotifierSerializable := new(ReplicationControllerNotifierSerializable)

	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return nil, err
	}
	err = session.Query("SELECT check, cool_down_duration, remaining_cool_down, kubeapi_host, kubeapi_port, namespace, kind, name, notifier_slice, indicator_slice FROM notifier WHERE namespace = ? AND kind = ? AND name = ?", namespace, kind, name).Scan(
		&replicationControllerNotifierSerializable.Check,
		&replicationControllerNotifierSerializable.CoolDownDuration,
		&replicationControllerNotifierSerializable.RemainingCoolDown,
		&replicationControllerNotifierSerializable.KubeApiServerEndPoint,
		&replicationControllerNotifierSerializable.KubeApiServerToken,
		&replicationControllerNotifierSerializable.Namespace,
		&replicationControllerNotifierSerializable.Kind,
		&replicationControllerNotifierSerializable.Name,
		&notifierSliceByteSlice,
		&indicatorSliceByteSlice,
	)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(indicatorSliceByteSlice, &replicationControllerNotifierSerializable.IndicatorSlice)
	if err != nil {
		log.Error("Unmarshal indicator slice error replicationControllerNotifierSerializable %s error: %s", replicationControllerNotifierSerializable, err)
		return nil, err
	}
	err = json.Unmarshal(notifierSliceByteSlice, &replicationControllerNotifierSerializable.NotifierSlice)
	if err != nil {
		log.Error("Unmarshal notifier slice error replicationControllerNotifierSerializable %s error: %s", replicationControllerNotifierSerializable, err)
		return nil, err
	}

	return replicationControllerNotifierSerializable, nil
}

func (storageCassandra *StorageCassandra) LoadAllReplicationControllerNotifierSerializable() ([]ReplicationControllerNotifierSerializable, error) {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return nil, err
	}
	iter := session.Query("SELECT check, cool_down_duration, remaining_cool_down, kubeapi_host, kubeapi_port, namespace, kind, name, notifier_slice, indicator_slice FROM notifier").Iter()

	replicationControllerNotifierSerializableSlice := make([]ReplicationControllerNotifierSerializable, 0)
	replicationControllerNotifierSerializable := new(ReplicationControllerNotifierSerializable)
	notifierSliceByteSlice := make([]byte, 0)
	indicatorSliceByteSlice := make([]byte, 0)

	for iter.Scan(
		&replicationControllerNotifierSerializable.Check,
		&replicationControllerNotifierSerializable.CoolDownDuration,
		&replicationControllerNotifierSerializable.RemainingCoolDown,
		&replicationControllerNotifierSerializable.KubeApiServerEndPoint,
		&replicationControllerNotifierSerializable.KubeApiServerToken,
		&replicationControllerNotifierSerializable.Namespace,
		&replicationControllerNotifierSerializable.Kind,
		&replicationControllerNotifierSerializable.Name,
		&notifierSliceByteSlice,
		&indicatorSliceByteSlice,
	) {
		err := json.Unmarshal(indicatorSliceByteSlice, &replicationControllerNotifierSerializable.IndicatorSlice)
		if err != nil {
			log.Error("Unmarshal indicator slice error replicationControllerNotifierSerializable %s error: %s", replicationControllerNotifierSerializable, err)
			return nil, err
		}
		err = json.Unmarshal(notifierSliceByteSlice, &replicationControllerNotifierSerializable.NotifierSlice)
		if err != nil {
			log.Error("Unmarshal notifier slice error replicationControllerNotifierSerializable %s error: %s", replicationControllerNotifierSerializable, err)
			return nil, err
		}

		replicationControllerNotifierSerializableSlice = append(replicationControllerNotifierSerializableSlice, *replicationControllerNotifierSerializable)
		replicationControllerNotifierSerializable = new(ReplicationControllerNotifierSerializable)
		notifierSliceByteSlice = make([]byte, 0)
		indicatorSliceByteSlice = make([]byte, 0)
	}

	err = iter.Close()
	if err != nil {
		return nil, err
	}

	return replicationControllerNotifierSerializableSlice, nil
}

func (storageCassandra *StorageCassandra) DeleteEmailServerSMTP(name string) error {
	return &storageCassandra.dummyError
}

func (storageCassandra *StorageCassandra) SaveEmailServerSMTP(emailServerSMTP *EmailServerSMTP) error {
	return &storageCassandra.dummyError
}

func (storageCassandra *StorageCassandra) LoadEmailServerSMTP(name string) (*EmailServerSMTP, error) {
	return nil, &storageCassandra.dummyError
}

func (storageCassandra *StorageCassandra) LoadAllEmailServerSMTP() ([]EmailServerSMTP, error) {
	return nil, &storageCassandra.dummyError
}

func (storageCassandra *StorageCassandra) DeleteSMSNexmo(name string) error {
	return &storageCassandra.dummyError
}

func (storageCassandra *StorageCassandra) SaveSMSNexmo(sMSNexmo *SMSNexmo) error {
	return &storageCassandra.dummyError
}

func (storageCassandra *StorageCassandra) LoadSMSNexmo(name string) (*SMSNexmo, error) {
	return nil, &storageCassandra.dummyError
}

func (storageCassandra *StorageCassandra) LoadAllSMSNexmo() ([]SMSNexmo, error) {
	return nil, &storageCassandra.dummyError
}
