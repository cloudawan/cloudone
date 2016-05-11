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

package image

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/cloudawan/cloudone/authorization"
	"github.com/cloudawan/cloudone/host"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/cloudawan/cloudone_utility/restclient"
	"github.com/cloudawan/cloudone_utility/sshclient"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ImageIdentifier struct {
	Repository string
	Tag        string
}

func DeleteImageInformationAndRelatedRecord(imageInformationName string) error {
	imageRecordSlice, err := GetStorage().LoadImageRecordWithImageInformationName(imageInformationName)
	if err != nil {
		log.Error(err)
		return err
	}

	imageIdentifierSlice := make([]ImageIdentifier, 0)
	for _, imageRecord := range imageRecordSlice {
		if imageRecord.Failure == false {
			repository := imageRecord.Path[:len(imageRecord.Path)-(len(imageRecord.Version)+1)] // Remove :version. +1 due to :
			imageIdentifierSlice = append(imageIdentifierSlice, ImageIdentifier{
				repository,
				imageRecord.Version,
			})
		}
	}

	hasError := false
	buffer := bytes.Buffer{}

	if len(imageIdentifierSlice) > 0 {
		err = RemoveImageFromPrivateRegistry(imageIdentifierSlice)
		if err != nil {
			hasError = true
			buffer.WriteString(err.Error())
		}

		err = RemoveImageFromAllHost(imageIdentifierSlice)
		if err != nil {
			hasError = true
			buffer.WriteString(err.Error())
		}
	}

	err = GetStorage().DeleteImageInformationAndRelatedRecord(imageInformationName)
	if err != nil {
		hasError = true
		buffer.WriteString(err.Error())
	}

	err = RequestDeleteBuildLogBelongingToImageInformation(imageInformationName)
	if err != nil {
		hasError = true
		buffer.WriteString(err.Error())
	}

	if hasError {
		log.Error(buffer.String())
		return errors.New(buffer.String())
	} else {
		return nil
	}
}

func DeleteImageRecord(imageInformationName string, imageRecordVersion string) error {
	imageRecord, err := GetStorage().LoadImageRecord(imageInformationName, imageRecordVersion)
	if err != nil {
		log.Error(err)
		return err
	}

	repository := imageRecord.Path[:len(imageRecord.Path)-(len(imageRecord.Version)+1)] // Remove :version. +1 due to :

	imageIdentifierSlice := make([]ImageIdentifier, 0)
	imageIdentifierSlice = append(imageIdentifierSlice, ImageIdentifier{
		repository,
		imageRecord.Version,
	})

	hasError := false
	buffer := bytes.Buffer{}

	if imageRecord.Failure == false {
		err = RemoveImageFromPrivateRegistry(imageIdentifierSlice)
		if err != nil {
			hasError = true
			buffer.WriteString(err.Error())
		}

		err = RemoveImageFromAllHost(imageIdentifierSlice)
		if err != nil {
			hasError = true
			buffer.WriteString(err.Error())
		}
	}

	err = GetStorage().DeleteImageRecord(imageInformationName, imageRecordVersion)
	if err != nil {
		hasError = true
		buffer.WriteString(err.Error())
	}

	err = RequestDeleteBuildLog(imageInformationName, imageRecordVersion)
	if err != nil {
		hasError = true
		buffer.WriteString(err.Error())
	}

	if hasError {
		log.Error(buffer.String())
		return errors.New(buffer.String())
	} else {
		return nil
	}
}

// Due to the docker registry API, the delete is only make it unavailable but the image is not removed from storage.
func RemoveImageFromPrivateRegistry(imageIdentifierSlice []ImageIdentifier) error {
	hasError := false
	buffer := bytes.Buffer{}
	for _, imageIdentifier := range imageIdentifierSlice {
		splitSlice := strings.Split(imageIdentifier.Repository, "/")
		if len(splitSlice) != 2 {
			hasError = true
			errorMessage := fmt.Sprintf("Invalid repository format %v.", imageIdentifier.Repository)
			log.Error(errorMessage)
			buffer.WriteString(errorMessage)
		} else {
			hostAndPort := splitSlice[0]
			repositoryName := splitSlice[1]

			request, err := http.NewRequest("GET", "https://"+hostAndPort+"/v2/"+repositoryName+"/manifests/"+imageIdentifier.Tag, nil)
			if err != nil {
				hasError = true
				errorMessage := fmt.Sprintf("Error during creating the request with imageIdentifier %v error %v.", imageIdentifier, err)
				log.Error(errorMessage)
				buffer.WriteString(errorMessage)
			} else {
				// For registry version 2.3 and later
				request.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
				response, err := restclient.GetInsecureHTTPSClient().Do(request)
				if err != nil {
					hasError = true
					errorMessage := fmt.Sprintf("Error during the request with imageIdentifier %v error %v.", imageIdentifier, err)
					log.Error(errorMessage)
					buffer.WriteString(errorMessage)
				} else {
					digest := response.Header.Get("Docker-Content-Digest")

					_, err := restclient.RequestDelete("https://"+hostAndPort+"/v2/"+repositoryName+"/manifests/"+digest, nil, nil, false)
					if err != nil {
						hasError = true
						errorMessage := fmt.Sprintf("Delete imageIdentifier %v error %v.", imageIdentifier, err)
						log.Error(errorMessage)
						buffer.WriteString(errorMessage)
					}
				}
			}
		}
	}

	if hasError {
		return errors.New(buffer.String())
	} else {
		return nil
	}
}

func RemoveImageFromAllHost(imageIdentifierSlice []ImageIdentifier) error {
	credentialSlice, err := host.GetStorage().LoadAllCredential()
	if err != nil {
		log.Error(err)
		return err
	}

	amount := len(imageIdentifierSlice)

	commandSlice := make([]string, 0)
	for _, imageIdentifier := range imageIdentifierSlice {
		commandSlice = append(commandSlice, "sudo docker rmi -f "+imageIdentifier.Repository+":"+imageIdentifier.Tag+"\n")
	}

	hasError := false
	buffer := bytes.Buffer{}
	for _, credential := range credentialSlice {
		interactiveMap := make(map[string]string)
		interactiveMap["[sudo]"] = credential.SSH.Password + "\n"

		resultSlice, err := sshclient.InteractiveSSH(
			2*time.Second,
			time.Duration(amount)*time.Minute,
			credential.IP,
			credential.SSH.Port,
			credential.SSH.User,
			credential.SSH.Password,
			commandSlice,
			interactiveMap)

		log.Info("Issue command via ssh with result:\n %v", resultSlice)

		if err != nil {
			hasError = true
			errorMessage := fmt.Sprintf("Error message: %v Result Output: %v .", err, resultSlice)
			log.Error(errorMessage)
			buffer.WriteString(errorMessage)
		}
	}

	if hasError {
		return errors.New(buffer.String())
	} else {
		return nil
	}
}

func RequestDeleteBuildLogBelongingToImageInformation(imageInformationName string) error {
	cloudoneAnalysisHost, ok := configuration.LocalConfiguration.GetString("cloudoneAnalysisHost")
	if ok == false {
		log.Error("Fail to get configuration cloudoneAnalysisHost")
		return errors.New("Fail to get configuration cloudoneAnalysisHost")
	}
	cloudoneAnalysisPort, ok := configuration.LocalConfiguration.GetInt("cloudoneAnalysisPort")
	if ok == false {
		log.Error("Fail to get configuration cloudoneAnalysisPort")
		return errors.New("Fail to get configuration cloudoneAnalysisPort")
	}

	url := "https://" + cloudoneAnalysisHost + ":" + strconv.Itoa(cloudoneAnalysisPort) + "/api/v1/buildlogs/" + imageInformationName

	headerMap := make(map[string]string)
	headerMap["token"] = authorization.SystemAdminToken

	_, err := restclient.RequestDelete(url, nil, headerMap, false)
	if err != nil {
		log.Error("Fail to request delete build image information %s log with error %s", imageInformationName, err)
	}

	return err
}

func RequestDeleteBuildLog(imageInformationName string, imageRecordVersion string) error {
	cloudoneAnalysisHost, ok := configuration.LocalConfiguration.GetString("cloudoneAnalysisHost")
	if ok == false {
		log.Error("Fail to get configuration cloudoneAnalysisHost")
		return errors.New("Fail to get configuration cloudoneAnalysisHost")
	}
	cloudoneAnalysisPort, ok := configuration.LocalConfiguration.GetInt("cloudoneAnalysisPort")
	if ok == false {
		log.Error("Fail to get configuration cloudoneAnalysisPort")
		return errors.New("Fail to get configuration cloudoneAnalysisPort")
	}

	url := "https://" + cloudoneAnalysisHost + ":" + strconv.Itoa(cloudoneAnalysisPort) + "/api/v1/buildlogs/" + imageInformationName + "/" + imageRecordVersion

	headerMap := make(map[string]string)
	headerMap["token"] = authorization.SystemAdminToken

	_, err := restclient.RequestDelete(url, nil, headerMap, false)
	if err != nil {
		log.Error("Fail to request delete build image information %s version %s log with error %s", imageInformationName, imageRecordVersion, err)
	}

	return err
}
