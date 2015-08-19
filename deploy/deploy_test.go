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

/*
import (
	"fmt"
	"github.com/cloudawan/kubernetes_management/control"
	"testing"
)

func TestDeployRollback(t *testing.T) {
	err := DeleteDeployInformation("qa", "test")
	if err != nil {
		t.Error(err)
	}
}
*/
/*
func TestDeployFirst(t *testing.T) {
	imageInformation := &ImageInformation{
		"test",
		"git",
		"/var/lib/kubernetes_management",
		"private-repository:31000/test",
		"https://github.com/cloudawan/test.git",
		"test",
		"src",
		"",
		"",
		"",
		"description",
		"",
	}

	replicaAmount := 2

	replicationControllerContainerPortSlice := make([]control.ReplicationControllerContainerPort, 0)
	replicationControllerContainerPortSlice = append(replicationControllerContainerPortSlice, control.ReplicationControllerContainerPort{"http-server", 8080})

	err := DeployFirst("192.168.0.33", 8080, "default", imageInformation, "record1", replicaAmount, replicationControllerContainerPortSlice)
	if err != nil {
		t.Error(err)
	}
}


func TestDeployUpdate(t *testing.T) {
	err := DeployUpdate("192.168.0.33", 8080, "default", "test", "record2")
	if err != nil {
		t.Error(err)
	}
}
*/
/*
func TestDeployRollback(t *testing.T) {

	imageRecordSlice, err := loadImageRecordWithImageInformationName("test")
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Rollback to version ", imageRecordSlice[0].Version)

	err = DeployRollback("192.168.0.33", 8080, "default", "test", imageRecordSlice[0].Version)
	if err != nil {
		t.Error(err)
	}
}
*/
