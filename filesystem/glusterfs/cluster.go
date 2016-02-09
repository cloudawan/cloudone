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

package glusterfs

import (
	"time"
)

type GlusterfsCluster struct {
	Name              string
	HostSlice         []string
	Path              string
	SSHDialTimeout    time.Duration
	SSHSessionTimeout time.Duration
	SSHPort           int
	SSHUser           string
	SSHPassword       string
}

func CreateGlusterfsCluster(name string, hostSlice []string, path string,
	sSHDialTimeoutInMilliSecond int, sSHSessionTimeoutInMilliSecond int, sSHPort int,
	sSHUser string, sSHPassword string) *GlusterfsCluster {

	glusterfsCluster := &GlusterfsCluster{
		name,
		hostSlice,
		path,
		time.Duration(sSHDialTimeoutInMilliSecond) * time.Millisecond,
		time.Duration(sSHSessionTimeoutInMilliSecond) * time.Millisecond,
		sSHPort,
		sSHUser,
		sSHPassword,
	}

	return glusterfsCluster
}
