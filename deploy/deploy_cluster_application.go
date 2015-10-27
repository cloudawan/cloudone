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
	"encoding/json"
	"errors"
	"github.com/cloudawan/kubernetes_management/application"
	"github.com/cloudawan/kubernetes_management/control"
	"github.com/cloudawan/kubernetes_management_utility/random"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
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

	replicationControllerAndRelatedPodSlice, err := control.GetAllReplicationControllerAndRelatedPodSlice(kubeapiHost, kubeapiPort, namespace)
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
		// Service
		serviceExist := false
		selectorMap := make(map[string]interface{})
		for _, service := range serviceSlice {
			if service.Name == cluster.Name {
				serviceExist = true
				for key, value := range service.Selector {
					selectorMap[key] = value
				}
			}
		}

		// Replication Controller
		size := 0
		owningReplicationControllerNameSlice := make([]string, 0)
		if len(selectorMap) > 0 {
			for _, replicationControllerAndRelatedPod := range replicationControllerAndRelatedPodSlice {
				// if all selectors in service are in the replication controller, the service owns the replication controller
				allFit := true
				for key, value := range selectorMap {
					if value != replicationControllerAndRelatedPod.Selector[key] {
						allFit = false
						break
					}
				}
				if allFit {
					size += len(replicationControllerAndRelatedPod.PodSlice)
					owningReplicationControllerNameSlice = append(owningReplicationControllerNameSlice, replicationControllerAndRelatedPod.Name)
				}
			}
		}

		// Service is must since it is used to provide endpoint for others to use
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

func ResizeDeployClusterApplication(kubeapiHost string, kubeapiPort int, namespace string, name string, environmentSlice []interface{}, size int) error {
	cluster, err := application.LoadClusterApplication(name)
	if err != nil {
		log.Error("Load cluster application error %s", err)
		return err
	}

	replicationControllerJsonMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(cluster.ReplicationControllerJson), &replicationControllerJsonMap)
	if err != nil {
		log.Error("Unmarshal replication controller json for cluster application error %s", err)
		return err
	}

	// Add environment variable
	if environmentSlice != nil {
		if replicationControllerJsonMap["spec"] != nil {
			containerSlice := replicationControllerJsonMap["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})
			for i := 0; i < len(containerSlice); i++ {
				_, ok := containerSlice[i].(map[string]interface{})["env"].([]interface{})
				if ok {
					for _, environment := range environmentSlice {
						containerSlice[i].(map[string]interface{})["env"] = append(containerSlice[i].(map[string]interface{})["env"].([]interface{}), environment)
					}
				} else {
					containerSlice[i].(map[string]interface{})["env"] = environmentSlice
				}
			}
		}
	}

	replicationControllerByteSlice, err := json.Marshal(replicationControllerJsonMap)
	if err != nil {
		log.Error("Marshal replication controller json for cluster application error %s", err)
		return err
	}

	environmentByteSlice, err := json.Marshal(environmentSlice)
	if err != nil {
		log.Error("Marshal environment json for cluster application error %s", err)
		return err
	}

	// Generate random work space
	workingDirectory := "/tmp/tmp_" + random.UUID()
	replicationControllerFileName := "replication-controller.json"
	scriptFileName := "script"
	environmentFileName := "environment.json"

	// Check working space
	if _, err := os.Stat(workingDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(workingDirectory, os.ModePerm)
		if err != nil {
			log.Error("Create non-existing directory %s error: %s", workingDirectory, err)
			return err
		}
	}

	err = ioutil.WriteFile(workingDirectory+string(os.PathSeparator)+environmentFileName, environmentByteSlice, os.ModePerm)
	if err != nil {
		log.Error("Write environment json file for cluster application error %s", err)
		return err
	}

	err = ioutil.WriteFile(workingDirectory+string(os.PathSeparator)+replicationControllerFileName, replicationControllerByteSlice, os.ModePerm)
	if err != nil {
		log.Error("Write replication controller json file for cluster application error %s", err)
		return err
	}

	err = ioutil.WriteFile(workingDirectory+string(os.PathSeparator)+scriptFileName, []byte(cluster.ScriptContent), os.ModePerm)
	if err != nil {
		log.Error("Write script file for cluster application error %s", err)
		return err
	}

	switch cluster.ScriptType {
	case "none":
		deployClusterApplicationSlice, err := GetAllDeployClusterApplication(kubeapiHost, kubeapiPort, namespace)
		if err != nil {
			log.Error("Get deploy cluster application slice error %s", err)
			return err
		}
		for _, deployClusterApplication := range deployClusterApplicationSlice {
			if deployClusterApplication.Name == name {
				for _, replicationControllerName := range deployClusterApplication.ReplicationControllerNameSlice {
					err := control.UpdateReplicationControllerSize(kubeapiHost, kubeapiPort, namespace, replicationControllerName, size)
					if err != nil {
						log.Error("Resize replication controller %s error %s", replicationControllerName, err)
						return err
					}
				}
			}
		}
	case "python":
		command := exec.Command("python", scriptFileName,
			"--application_name="+name,
			"--kubeapi_host_and_port=http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort),
			"--namespace="+namespace,
			"--size="+strconv.Itoa(size),
			"--replication_controller_file_name="+replicationControllerFileName,
			"--environment_file_name="+environmentFileName,
			"--timeout_in_second=120",
			"--action=resize")
		command.Dir = workingDirectory
		out, err := command.CombinedOutput()
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
	case "none":
		deployClusterApplicationSlice, err := GetAllDeployClusterApplication(kubeapiHost, kubeapiPort, namespace)
		if err != nil {
			log.Error("Get deploy cluster application slice error %s", err)
			return err
		}
		for _, deployClusterApplication := range deployClusterApplicationSlice {
			if deployClusterApplication.Name == name {
				for _, replicationControllerName := range deployClusterApplication.ReplicationControllerNameSlice {
					err := control.DeleteReplicationController(kubeapiHost, kubeapiPort, namespace, replicationControllerName)
					if err != nil {
						log.Error("Delete replication controller %s error %s", replicationControllerName, err)
						return err
					}
				}
				err := control.DeleteService(kubeapiHost, kubeapiPort, namespace, deployClusterApplication.ServiceName)
				if err != nil {
					log.Error("Delete service %s error %s", deployClusterApplication.ServiceName, err)
					return err
				}
			}
		}
	case "python":
		command := exec.Command("python", scriptFileName,
			"--application_name="+name,
			"--kubeapi_host_and_port=http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort),
			"--namespace="+namespace,
			"--timeout_in_second=120",
			"--action=delete")
		command.Dir = workingDirectory
		out, err := command.CombinedOutput()
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
