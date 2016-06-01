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

package registry

import (
	"bytes"
	"errors"
	"github.com/cloudawan/cloudone_utility/restclient"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PrivateRegistry struct {
	Name string
	Host string
	Port int
}

const (
	AvailableTimeoutDuration = time.Second * 1
)

func GetPrivateRegistryFromPathAndTestAvailable(path string) (*PrivateRegistry, error) {
	splitSlice := strings.Split(path, "/")
	if len(splitSlice) != 2 {
		return nil, errors.New("Invalid path format: " + path)
	}
	registry := splitSlice[0]
	splitSlice = strings.Split(registry, ":")
	if len(splitSlice) != 2 {
		return nil, errors.New("Invalid registry format: " + registry)
	}
	host := splitSlice[0]
	portText := splitSlice[1]

	port, err := strconv.Atoi(portText)
	if err != nil {
		log.Error(err)
		return nil, errors.New("Can't parse port: " + portText)
	}

	privateRegistry := &PrivateRegistry{
		"",
		host,
		port,
	}

	err = privateRegistry.IsAvailable()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return privateRegistry, nil
}

func (privateRegistry *PrivateRegistry) IsAvailable() error {
	_, err := restclient.HealthCheck(privateRegistry.getPrivateRegistryEndpoint()+"/v2/", nil, AvailableTimeoutDuration)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (privateRegistry *PrivateRegistry) GetRepositoryPath(repositoryName string) string {
	return privateRegistry.Host + ":" + strconv.Itoa(privateRegistry.Port) + "/" + repositoryName
}

func (privateRegistry *PrivateRegistry) getPrivateRegistryEndpoint() string {
	return "https://" + privateRegistry.Host + ":" + strconv.Itoa(privateRegistry.Port)
}

func (privateRegistry *PrivateRegistry) GetAllRepository() ([]string, error) {
	result, err := restclient.RequestGet(privateRegistry.getPrivateRegistryEndpoint()+"/v2/_catalog", nil, false)
	if err != nil {
		log.Error("Fail to get all repository with private registry: %v, error: %s", privateRegistry, err)
		return nil, err
	}

	jsonMap, _ := result.(map[string]interface{})
	repositoryJsonSlice, _ := jsonMap["repositories"].([]interface{})
	repositorySlice := make([]string, 0)
	for _, repositoryJsonInterface := range repositoryJsonSlice {
		if repository, ok := repositoryJsonInterface.(string); ok {
			repositorySlice = append(repositorySlice, repository)
		}
	}

	return repositorySlice, nil
}

func (privateRegistry *PrivateRegistry) DeleteAllImageInRepository(repositoryName string) error {
	tagSlice, err := privateRegistry.GetAllImageTag(repositoryName)
	if err != nil {
		log.Error(err)
		return err
	}

	hasError := false
	errorBuffer := bytes.Buffer{}
	for _, tag := range tagSlice {
		if err := privateRegistry.DeleteImageInRepository(repositoryName, tag); err != nil {
			log.Error(err)
			errorBuffer.WriteString(err.Error())
			errorBuffer.WriteByte('\n')
			hasError = true
		}
	}

	if hasError {
		return errors.New(errorBuffer.String())
	}

	return nil
}

func (privateRegistry *PrivateRegistry) GetAllImageTag(repositoryName string) ([]string, error) {
	result, err := restclient.RequestGet(privateRegistry.getPrivateRegistryEndpoint()+"/v2/"+repositoryName+"/tags/list", nil, false)
	if err != nil {
		log.Error("Fail to get all image tags with repository %s and private registry: %v, error: %s", repositoryName, privateRegistry, err)
		return nil, err
	}

	jsonMap, _ := result.(map[string]interface{})
	tagJsonSlice, _ := jsonMap["tags"].([]interface{})
	tagSlice := make([]string, 0)
	for _, tagJsonInterface := range tagJsonSlice {
		if tag, ok := tagJsonInterface.(string); ok {
			request, err := http.NewRequest("GET", privateRegistry.getPrivateRegistryEndpoint()+"/v2/"+repositoryName+"/manifests/"+tag, nil)
			if err != nil {
				log.Error(err)
				return nil, err
			}

			// For registry version 2.3 and later
			request.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
			response, err := restclient.GetInsecureHTTPSClient().Do(request)
			if err != nil {
				log.Error(err)
				return nil, err
			}
			digest := response.Header.Get("Docker-Content-Digest")

			if len(digest) > 0 {
				tagSlice = append(tagSlice, tag)
			}
		}
	}

	return tagSlice, nil
}

func (privateRegistry *PrivateRegistry) DeleteImageInRepository(repositoryName string, tag string) error {
	request, err := http.NewRequest("GET", privateRegistry.getPrivateRegistryEndpoint()+"/v2/"+repositoryName+"/manifests/"+tag, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	// For registry version 2.3 and later
	request.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	response, err := restclient.GetInsecureHTTPSClient().Do(request)
	if err != nil {
		log.Error(err)
		return err
	}
	digest := response.Header.Get("Docker-Content-Digest")

	if len(digest) == 0 {
		// The tag has no image
		return nil
	}

	_, err = restclient.RequestDelete(privateRegistry.getPrivateRegistryEndpoint()+"/v2/"+repositoryName+"/manifests/"+digest, nil, nil, false)
	requestError, _ := err.(restclient.RequestError)
	if requestError.StatusCode == 404 {
		// Not found so the target doesn't exist
		return nil
	}
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (privateRegistry *PrivateRegistry) IsImageTagAvailable(repositoryName string, targetTag string) bool {
	tagSlice, err := privateRegistry.GetAllImageTag(repositoryName)
	if err != nil {
		log.Error(err)
		return false
	}

	for _, tag := range tagSlice {
		if tag == targetTag {
			return true
		}
	}

	return false
}
