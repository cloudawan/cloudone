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

package application

import (
	"encoding/json"
	"errors"
	"github.com/cloudawan/cloudone/control"
	"github.com/cloudawan/cloudone_utility/deepcopy"
	"github.com/cloudawan/cloudone_utility/random"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

type Cluster struct {
	Name                      string
	Description               string
	ReplicationControllerJson string
	ServiceJson               string
	Environment               map[string]string
	ScriptType                string
	ScriptContent             string
}

func LaunchClusterApplication(kubeapiHost string, kubeapiPort int, namespace string, name string, environmentSlice []interface{}, size int, replicationControllerExtraJsonMap map[string]interface{}) error {
	cluster, err := GetStorage().LoadClusterApplication(name)
	if err != nil {
		log.Error("Load cluster application error %s", err)
		return err
	}

	switch cluster.ScriptType {
	case "none":
		return LaunchClusterApplicationNoScript(kubeapiHost, kubeapiPort, namespace, cluster, environmentSlice, size, replicationControllerExtraJsonMap)
	case "python":
		return LaunchClusterApplicationPython(kubeapiHost, kubeapiPort, namespace, cluster, environmentSlice, size, replicationControllerExtraJsonMap)
	default:
		log.Error("No such script type: %s", cluster.ScriptType)
		return errors.New("No such script type: " + cluster.ScriptType)
	}
}

func LaunchClusterApplicationNoScript(kubeapiHost string, kubeapiPort int, namespace string, cluster *Cluster, environmentSlice []interface{}, size int, replicationControllerExtraJsonMap map[string]interface{}) error {
	replicationControllerJsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(cluster.ReplicationControllerJson), &replicationControllerJsonMap)
	if err != nil {
		log.Error("Unmarshal replication controller json for cluster application error %s", err)
		return err
	}

	serviceJsonMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(cluster.ServiceJson), &serviceJsonMap)
	if err != nil {
		log.Error("Unmarshal service json for cluster application error %s", err)
		return err
	}

	// Configure extra json body
	// It is used for user to input any configuration
	if replicationControllerExtraJsonMap != nil {
		deepcopy.DeepOverwriteJsonMap(replicationControllerExtraJsonMap, replicationControllerJsonMap)
	}

	// Add environment variable
	if environmentSlice != nil {
		specJsonMap, ok := replicationControllerJsonMap["spec"].(map[string]interface{})
		if ok {
			templateJsonMap, ok := specJsonMap["template"].(map[string]interface{})
			if ok {
				specJsonMap, ok := templateJsonMap["spec"].(map[string]interface{})
				if ok {
					containerSlice, ok := specJsonMap["containers"].([]interface{})
					if ok {
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
			}
		}
	}

	// Change size
	specJsonMap, ok := replicationControllerJsonMap["spec"].(map[string]interface{})
	if ok {
		specJsonMap["replicas"] = size
	}

	err = control.CreateReplicationControllerWithJson(kubeapiHost, kubeapiPort, namespace, replicationControllerJsonMap)
	if err != nil {
		log.Error("CreateReplicationControllerWithJson error %s", err)
		return err
	}
	err = control.CreateServiceWithJson(kubeapiHost, kubeapiPort, namespace, serviceJsonMap)
	if err != nil {
		log.Error("CreateReplicationControllerWithJson error %s", err)
		return err
	}

	return nil
}

func LaunchClusterApplicationPython(kubeapiHost string, kubeapiPort int, namespace string, cluster *Cluster, environmentSlice []interface{}, size int, replicationControllerExtraJsonMap map[string]interface{}) error {
	replicationControllerJsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(cluster.ReplicationControllerJson), &replicationControllerJsonMap)
	if err != nil {
		log.Error("Unmarshal replication controller json for cluster application error %s", err)
		return err
	}

	// Configure extra json body
	// It is used for user to input any configuration
	if replicationControllerExtraJsonMap != nil {
		deepcopy.DeepOverwriteJsonMap(replicationControllerExtraJsonMap, replicationControllerJsonMap)
	}

	// Add environment variable
	if environmentSlice != nil {
		specJsonMap, ok := replicationControllerJsonMap["spec"].(map[string]interface{})
		if ok {
			templateJsonMap, ok := specJsonMap["template"].(map[string]interface{})
			if ok {
				specJsonMap, ok := templateJsonMap["spec"].(map[string]interface{})
				if ok {
					containerSlice, ok := specJsonMap["containers"].([]interface{})
					if ok {
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
	serviceFileName := "service.json"
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

	err = ioutil.WriteFile(workingDirectory+string(os.PathSeparator)+serviceFileName, []byte(cluster.ServiceJson), os.ModePerm)
	if err != nil {
		log.Error("Write service json file for cluster application error %s", err)
		return err
	}

	err = ioutil.WriteFile(workingDirectory+string(os.PathSeparator)+scriptFileName, []byte(cluster.ScriptContent), os.ModePerm)
	if err != nil {
		log.Error("Write script file for cluster application error %s", err)
		return err
	}

	command := exec.Command("python", scriptFileName,
		"--application_name="+cluster.Name,
		"--kubeapi_host_and_port=http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort),
		"--namespace="+namespace,
		"--size="+strconv.Itoa(size),
		"--service_file_name="+serviceFileName,
		"--replication_controller_file_name="+replicationControllerFileName,
		"--environment_file_name="+environmentFileName,
		"--timeout_in_second=120",
		"--action=create")
	command.Dir = workingDirectory
	out, err := command.CombinedOutput()
	log.Debug(string(out))
	if err != nil {
		log.Error("Run python script file for cluster application error %s", err)
		return errors.New("Error: " + err.Error() + " Output: " + string(out))
	}

	// Remove working space
	err = os.RemoveAll(workingDirectory)
	if err != nil {
		log.Error("Remove the working directory for cluster application error %s", err)
		return err
	}

	return nil
}
