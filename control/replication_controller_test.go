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

package control

/*
import (
	"fmt"
	//"strings"
	"testing"
	//"time"
)

func TestUpdateReplicationControllerSize(t *testing.T) {
	fmt.Println(UpdateReplicationControllerSize("172.16.0.113", 8080, "default", "flask", 1))
}


*/
/*
func TestGetReplicationController(t *testing.T) {
	fmt.Println(GetReplicationController("192.168.0.33", 8080, "default", "flask"))
}
*/

/*
func TestResizeReplicationController(t *testing.T) {
	fmt.Println(ResizeReplicationController(-1, "192.168.0.33", 8080, "default", "nginx"))
}
*/

/*
func TestGetAllReplicationControllerName(t *testing.T) {
	fmt.Println(GetAllReplicationControllerName("192.168.0.33", 8080, "default"))
}
*/
/*
func TestCreateReplicationController(t *testing.T) {

	replicationControllerContainerPortSlice := make([]ReplicationControllerContainerPort, 0)
	replicationControllerContainerPortSlice = append(replicationControllerContainerPortSlice, ReplicationControllerContainerPort{"http-server", 8086})

	replicationControllerContainerSlice := make([]ReplicationControllerContainer, 0)
	replicationControllerContainerSlice = append(replicationControllerContainerSlice, ReplicationControllerContainer{"flask2", "private-repository:31000/flask", replicationControllerContainerPortSlice})

	labelMap := make(map[string]string)
	labelMap["name"] = "flask2"

	replicationController := ReplicationController{"flask2", "flask2", 1, labelMap, labelMap, replicationControllerContainerSlice}
	fmt.Println(CreateReplicationController("192.168.0.33", 8080, "default", replicationController))
}
*/
/*
func TestDeleteReplicationControllerAndRelatedPod(t *testing.T) {
	fmt.Println(DeleteReplicationControllerAndRelatedPod("192.168.0.33", 8080, "default", "flask2"))
}
*/
/*
func TestRollingUpdateReplicationControllerWithSingleContainer(t *testing.T) {
	fmt.Println(RollingUpdateReplicationControllerWithSingleContainer("192.168.0.33", 8080, "default", "test2015-07-11-06-30-04", "test2015-07-11-06-26-12", "192.168.0.33:5000/test:2015-07-11-06-26-12", "2015-07-11-06-26-12", 10*time.Second))
}
*/
/*
func TestRollingUpdateReplicationControllerWithSingleContainer2(t *testing.T) {
	fmt.Println(RollingUpdateReplicationControllerWithSingleContainer("192.168.0.33", 8080, "default", "test2015-06-21-08-53-55", "test2015-06-21-08-51-26", "192.168.0.33:5000/test:2015-06-21-08-51-26", "2015-06-21-08-51-26", 10*time.Second))
}
*/
/*
func TestCreateReplicationController2(t *testing.T) {
	version := "2015-06-21-08-51-26"
	selectorName := "test"
	name := selectorName + version
	image := "private-repository:31000/test:2015-06-21-08-51-26"

	replicationControllerContainerPortSlice := make([]ReplicationControllerContainerPort, 0)
	replicationControllerContainerPortSlice = append(replicationControllerContainerPortSlice, ReplicationControllerContainerPort{"http-server", 8080})

	replicationControllerContainerSlice := make([]ReplicationControllerContainer, 0)
	replicationControllerContainerSlice = append(replicationControllerContainerSlice, ReplicationControllerContainer{name, image, replicationControllerContainerPortSlice})

	replicationController := ReplicationController{
		name,
		2,
		ReplicationControllerSelector{
			selectorName,
			version,
		},
		ReplicationControllerLabel{
			name,
		},
		replicationControllerContainerSlice,
	}
	fmt.Println(CreateReplicationController("192.168.0.33", 8080, "default", replicationController))
}
*/
