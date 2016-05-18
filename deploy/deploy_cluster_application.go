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
	"github.com/cloudawan/cloudone/application"
	"github.com/cloudawan/cloudone/control"
	"github.com/cloudawan/cloudone_utility/random"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type DeployClusterApplication struct {
	Name                              string
	Namespace                         string
	Size                              int
	EnvironmentSlice                  []interface{}
	ReplicationControllerExtraJsonMap map[string]interface{}
	ServiceName                       string
	ReplicationControllerNameSlice    []string
	CreatedTime                       time.Time
}

func getServiceNameAndReplicationControllerNameSlice(kubeapiHost string, kubeapiPort int, namespace string, name string) (bool, string, []string, error) {
	replicationControllerAndRelatedPodSlice, err := control.GetAllReplicationControllerAndRelatedPodSlice(kubeapiHost, kubeapiPort, namespace)
	if err != nil {
		log.Error("Fail to get all replication controller name error %s", err)
		return false, "", nil, err
	}

	serviceSlice, err := control.GetAllService(kubeapiHost, kubeapiPort, namespace)
	if err != nil {
		log.Error("Fail to get all service error %s", err)
		return false, "", nil, err
	}

	// Service
	serviceName := name
	serviceExist := false
	selectorMap := make(map[string]interface{})
	for _, service := range serviceSlice {
		if service.Name == serviceName {
			serviceExist = true
			for key, value := range service.Selector {
				selectorMap[key] = value
			}
		}
	}

	// Replication Controller
	replicationControllerNameSlice := make([]string, 0)
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
				replicationControllerNameSlice = append(replicationControllerNameSlice, replicationControllerAndRelatedPod.Name)
			}
		}
	}

	return serviceExist, serviceName, replicationControllerNameSlice, nil
}

func InitializeDeployClusterApplication(kubeapiHost string, kubeapiPort int, namespace string, name string, environmentSlice []interface{}, size int, replicationControllerExtraJsonMap map[string]interface{}) error {
	serviceExist, serviceName, replicationControllerNameSlice, err := getServiceNameAndReplicationControllerNameSlice(kubeapiHost, kubeapiPort, namespace, name)
	if err != nil {
		log.Error(err)
		return err
	}

	// Service is must since it is used to provide endpoint for others to use
	if serviceExist {
		deployClusterApplication := &DeployClusterApplication{
			name,
			namespace,
			size,
			environmentSlice,
			replicationControllerExtraJsonMap,
			serviceName,
			replicationControllerNameSlice,
			time.Now(),
		}

		return GetStorage().SaveDeployClusterApplication(deployClusterApplication)
	} else {
		log.Error("The service doesn't exist for cluster application deployment with name %s", name)
		return errors.New("The service doesn't exist for cluster application deployment with name " + name)
	}
}

func GetAllDeployClusterApplication() ([]DeployClusterApplication, error) {
	return GetStorage().LoadAllDeployClusterApplication()
}

func GetAllDeployClusterApplicationInNamespace(namespace string) ([]DeployClusterApplication, error) {
	deployClusterApplicationSlice, err := GetStorage().LoadAllDeployClusterApplication()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	filteredDeployClusterApplicationSlice := make([]DeployClusterApplication, 0)
	for _, deployClusterApplication := range deployClusterApplicationSlice {
		if deployClusterApplication.Namespace == namespace {
			filteredDeployClusterApplicationSlice = append(filteredDeployClusterApplicationSlice, deployClusterApplication)
		}
	}

	return filteredDeployClusterApplicationSlice, nil
}

func GetDeployClusterApplication(namespace string, name string) (*DeployClusterApplication, error) {
	return GetStorage().LoadDeployClusterApplication(namespace, name)
}

func ResizeDeployClusterApplication(kubeapiHost string, kubeapiPort int, namespace string, name string, environmentSlice []interface{}, size int) error {
	cluster, err := application.GetStorage().LoadClusterApplication(name)
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
		deployClusterApplication, err := GetDeployClusterApplication(namespace, name)
		if err != nil {
			log.Error("Get deploy cluster application error %s", err)
			return err
		}
		for _, replicationControllerName := range deployClusterApplication.ReplicationControllerNameSlice {
			err := control.UpdateReplicationControllerSize(kubeapiHost, kubeapiPort, namespace, replicationControllerName, size)
			if err != nil {
				log.Error("Resize replication controller %s error %s", replicationControllerName, err)
				return err
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
			return errors.New("Error: " + err.Error() + " Output: " + string(out))
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

	// Update deploy data
	deployClusterApplication, err := GetStorage().LoadDeployClusterApplication(namespace, name)
	if err != nil {
		log.Error("Update the deploy cluster application with error %s", err)
		return err
	}
	_, serviceName, replicationControllerNameSlice, err := getServiceNameAndReplicationControllerNameSlice(kubeapiHost, kubeapiPort, namespace, name)
	if err != nil {
		log.Error(err)
		return err
	}

	deployClusterApplication.Size = size
	deployClusterApplication.EnvironmentSlice = environmentSlice
	deployClusterApplication.ServiceName = serviceName
	deployClusterApplication.ReplicationControllerNameSlice = replicationControllerNameSlice
	err = GetStorage().SaveDeployClusterApplication(deployClusterApplication)
	if err != nil {
		log.Error("Save the deploy cluster application error %s", err)
		return err
	}

	return nil
}

func DeleteDeployClusterApplication(kubeapiHost string, kubeapiPort int, namespace string, name string) error {
	deployClusterApplication, err := GetDeployClusterApplication(namespace, name)
	if err != nil {
		log.Error("Get deploy cluster application error %s", err)
		return err
	}

	// Remove deploy cluster application
	err = GetStorage().DeleteDeployClusterApplication(namespace, name)
	if err != nil {
		log.Error("Delete the deploy cluster application error %s", err)
		return err
	}

	cluster, err := application.GetStorage().LoadClusterApplication(name)
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
		for _, replicationControllerName := range deployClusterApplication.ReplicationControllerNameSlice {
			err := control.DeleteReplicationControllerAndRelatedPod(kubeapiHost, kubeapiPort, namespace, replicationControllerName)
			if err != nil {
				log.Error("Delete replication controller %s error %s", replicationControllerName, err)
				return err
			}
		}

		err = control.DeleteService(kubeapiHost, kubeapiPort, namespace, deployClusterApplication.ServiceName)
		if err != nil {
			log.Error("Delete service %s error %s", deployClusterApplication.ServiceName, err)
			return err
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
			return errors.New("Error: " + err.Error() + " Output: " + string(out))
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
