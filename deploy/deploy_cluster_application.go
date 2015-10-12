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
	"errors"
	"github.com/cloudawan/kubernetes_management/application"
	"github.com/cloudawan/kubernetes_management/control"
	"github.com/cloudawan/kubernetes_management_utility/random"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type DeployClusterApplication struct {
	Name                           string
	Size                           int
	ServiceName                    string
	ReplicationControllerNameSlice []string
}

func GetAllDeployClusterApplication(kubeapiHost string, kubeapiPort int, namespace string) ([]DeployClusterApplication, error) {
	clusterSlice, err := application.LoadAllClusterApplication()
	if err != nil {
		log.Error("Fail to load all cluster application error %s", err)
		return nil, err
	}

	replicationControllerNameSlice, err := control.GetAllReplicationControllerName(kubeapiHost, kubeapiPort, namespace)
	if err != nil {
		log.Error("Fail to get all replication controller name error %s", err)
		return nil, err
	}

	serviceSlice, err := control.GetAllService(kubeapiHost, kubeapiPort, namespace)
	if err != nil {
		log.Error("Fail to get all service error %s", err)
		return nil, err
	}

	deployClusterApplicationSlice := make([]DeployClusterApplication, 0)
	for _, cluster := range clusterSlice {
		owningReplicationControllerNameSlice := make([]string, 0)
		for _, replicationControllerName := range replicationControllerNameSlice {
			if strings.HasPrefix(replicationControllerName, cluster.Name+"-instance-") {
				owningReplicationControllerNameSlice = append(owningReplicationControllerNameSlice, replicationControllerName)
			}
		}
		size := len(owningReplicationControllerNameSlice)
		serviceExist := false
		for _, service := range serviceSlice {
			if service.Name == cluster.Name {
				serviceExist = true
			}
		}

		if serviceExist {
			deployClusterApplication := DeployClusterApplication{
				cluster.Name,
				size,
				cluster.Name,
				owningReplicationControllerNameSlice,
			}
			deployClusterApplicationSlice = append(deployClusterApplicationSlice, deployClusterApplication)
		}
	}

	return deployClusterApplicationSlice, nil
}

func DeleteDeployClusterApplication(kubeapiHost string, kubeapiPort int, namespace string, name string) error {
	cluster, err := application.LoadClusterApplication(name)
	if err != nil {
		log.Error("Load cluster application error %s", err)
		return err
	}

	// Generate random work space
	workingDirectory := "/tmp/tmp_" + random.UUID()
	scriptFileName := "script"

	// Check working space
	if _, err := os.Stat(workingDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(workingDirectory, os.ModePerm)
		if err != nil {
			log.Error("Create non-existing directory %s error: %s", workingDirectory, err)
			return err
		}
	}

	err = ioutil.WriteFile(workingDirectory+string(os.PathSeparator)+scriptFileName, []byte(cluster.ScriptContent), os.ModePerm)
	if err != nil {
		log.Error("Write script file for cluster application error %s", err)
		return err
	}

	switch cluster.ScriptType {
	case "python":
		command := exec.Command("python", scriptFileName,
			"--application_name="+name,
			"--kubeapi_host_and_port=http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort),
			"--namespace="+namespace,
			"--timeout_in_second=120",
			"--action=delete")
		command.Dir = workingDirectory
		out, err := command.CombinedOutput()
		command.CombinedOutput()
		log.Debug(string(out))
		if err != nil {
			log.Error("Run python script file for cluster application error %s", err)
			return err
		}
	default:
		log.Error("No such script type: %s", cluster.ScriptType)
		return errors.New("No such script type: " + cluster.ScriptType)
	}

	// Remove working space
	err = os.RemoveAll(workingDirectory)
	if err != nil {
		log.Error("Remove the working directory for cluster application error %s", err)
		return err
	}

	return nil
}
