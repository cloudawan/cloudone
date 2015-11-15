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

package autoscaler

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/cassandra"
	"time"
)

var tableSchemaAutoscaler = `
	CREATE TABLE IF NOT EXISTS auto_scaler (
	check boolean,
	cool_down_duration bigint,
	remaining_cool_down bigint,
	kubeapi_host varchar,
	kubeapi_port int,
	namespace varchar,
	kind varchar,
	name varchar,
	maximum_replica int,
	minimum_replica int,
	indicator_slice blob,
	PRIMARY KEY (namespace, kind, name));
	`

func init() {
	err := cassandra.CassandraClient.CreateTableIfNotExist(tableSchemaAutoscaler, 3, time.Second*5)
	if err != nil {
		log.Critical("Fail to create table with schema %s", tableSchemaAutoscaler)
		panic(err)
	}
}

func DeleteReplicationControllerAutoScaler(namespace string, kind string, name string) error {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("DELETE FROM auto_scaler WHERE namespace = ? AND kind = ? AND name = ?", namespace, kind, name).Exec(); err != nil {
		log.Error("Delete auto_scaler with namespace %s kind %s name %s error: %s", namespace, kind, name, err)
		return err
	}
	return nil
}

func SaveReplicationControllerAutoScaler(replicationControllerAutoScaler *ReplicationControllerAutoScaler) error {
	indicatorSliceByteSlice, err := json.Marshal(replicationControllerAutoScaler.IndicatorSlice)
	if err != nil {
		log.Error("Marshal indicator slice error replicationControllerAutoScaler %s error: %s", replicationControllerAutoScaler, err)
		return err
	}

	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return err
	}
	if err := session.Query("INSERT INTO auto_scaler (check, cool_down_duration, remaining_cool_down, kubeapi_host, kubeapi_port, namespace, kind, name, maximum_replica, minimum_replica, indicator_slice) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		replicationControllerAutoScaler.Check,
		replicationControllerAutoScaler.CoolDownDuration,
		replicationControllerAutoScaler.RemainingCoolDown,
		replicationControllerAutoScaler.KubeapiHost,
		replicationControllerAutoScaler.KubeapiPort,
		replicationControllerAutoScaler.Namespace,
		replicationControllerAutoScaler.Kind,
		replicationControllerAutoScaler.Name,
		replicationControllerAutoScaler.MaximumReplica,
		replicationControllerAutoScaler.MinimumReplica,
		indicatorSliceByteSlice,
	).Exec(); err != nil {
		log.Error("Save replicationControllerAutoScaler %s error: %s", replicationControllerAutoScaler, err)
		return err
	}
	return nil
}

func LoadReplicationControllerAutoScaler(namespace string, kind string, name string) (*ReplicationControllerAutoScaler, error) {
	indicatorSliceByteSlice := make([]byte, 0)
	replicationControllerAutoScaler := new(ReplicationControllerAutoScaler)

	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return nil, err
	}
	err = session.Query("SELECT check, cool_down_duration, remaining_cool_down, kubeapi_host, kubeapi_port, namespace, kind, name, maximum_replica, minimum_replica, indicator_slice FROM auto_scaler WHERE namespace = ? AND kind = ? AND name = ?", namespace, kind, name).Scan(
		&replicationControllerAutoScaler.Check,
		&replicationControllerAutoScaler.CoolDownDuration,
		&replicationControllerAutoScaler.RemainingCoolDown,
		&replicationControllerAutoScaler.KubeapiHost,
		&replicationControllerAutoScaler.KubeapiPort,
		&replicationControllerAutoScaler.Namespace,
		&replicationControllerAutoScaler.Kind,
		&replicationControllerAutoScaler.Name,
		&replicationControllerAutoScaler.MaximumReplica,
		&replicationControllerAutoScaler.MinimumReplica,
		&indicatorSliceByteSlice,
	)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(indicatorSliceByteSlice, &replicationControllerAutoScaler.IndicatorSlice)
	if err != nil {
		log.Error("Unmarshal indicator slice error replicationControllerAutoScaler %s error: %s", replicationControllerAutoScaler, err)
		return nil, err
	}

	return replicationControllerAutoScaler, nil
}

func LoadAllReplicationControllerAutoScaler() ([]ReplicationControllerAutoScaler, error) {
	session, err := cassandra.CassandraClient.GetSession()
	if err != nil {
		log.Error("Get session error %s", err)
		return nil, err
	}
	iter := session.Query("SELECT check, cool_down_duration, remaining_cool_down, kubeapi_host, kubeapi_port, namespace, kind, name, maximum_replica, minimum_replica, indicator_slice FROM auto_scaler").Iter()

	replicationControllerAutoScalerSlice := make([]ReplicationControllerAutoScaler, 0)
	replicationControllerAutoScaler := new(ReplicationControllerAutoScaler)
	indicatorSliceByteSlice := make([]byte, 0)

	for iter.Scan(
		&replicationControllerAutoScaler.Check,
		&replicationControllerAutoScaler.CoolDownDuration,
		&replicationControllerAutoScaler.RemainingCoolDown,
		&replicationControllerAutoScaler.KubeapiHost,
		&replicationControllerAutoScaler.KubeapiPort,
		&replicationControllerAutoScaler.Namespace,
		&replicationControllerAutoScaler.Kind,
		&replicationControllerAutoScaler.Name,
		&replicationControllerAutoScaler.MaximumReplica,
		&replicationControllerAutoScaler.MinimumReplica,
		&indicatorSliceByteSlice,
	) {
		err := json.Unmarshal(indicatorSliceByteSlice, &replicationControllerAutoScaler.IndicatorSlice)
		if err != nil {
			log.Error("Unmarshal indicator slice error replicationControllerAutoScaler %s error: %s", replicationControllerAutoScaler, err)
			return nil, err
		}

		replicationControllerAutoScalerSlice = append(replicationControllerAutoScalerSlice, *replicationControllerAutoScaler)
		replicationControllerAutoScaler = new(ReplicationControllerAutoScaler)
		indicatorSliceByteSlice = make([]byte, 0)
	}

	err = iter.Close()
	if err != nil {
		return nil, err
	}

	return replicationControllerAutoScalerSlice, nil
}
