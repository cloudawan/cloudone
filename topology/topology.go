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

package topology

import (
	"github.com/cloudawan/cloudone/control"
	"github.com/cloudawan/cloudone/deploy"
	"time"
)

type Topology struct {
	Name            string
	SourceNamespace string
	CreatedUser     string
	CreatedDate     time.Time
	Description     string
	LaunchSlice     []Launch
}

type Launch struct {
	Order                    int
	LaunchApplication        *LaunchApplication
	LaunchClusterApplication *LaunchClusterApplication
}

type LaunchApplication struct {
	ImageInformationName string
	Version              string
	Description          string
	ReplicaAmount        int
	PortSlice            []deploy.DeployContainerPort
	EnvironmentSlice     []control.ReplicationControllerContainerEnvironment
	ResourceMap          map[string]interface{}
	ExtraJsonMap         map[string]interface{}
}

type LaunchClusterApplication struct {
	Name                              string
	Size                              int
	EnvironmentSlice                  []interface{}
	ReplicationControllerExtraJsonMap map[string]interface{}
}
