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
	"github.com/cloudawan/cloudone/registry"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/cloudawan/cloudone_utility/restclient"
	"github.com/cloudawan/cloudone_utility/sshclient"
	"strconv"
	"time"
)

func DeleteImageInformationAndRelatedRecord(imageInformationName string) error {
	imageRecordSlice, err := GetStorage().LoadImageRecordWithImageInformationName(imageInformationName)
	if err != nil {
		log.Error(err)
		return err
	}

	filteredImageRecordSlice := make([]ImageRecord, 0)
	for _, imageRecord := range imageRecordSlice {
		// Image is successfully built before
		if imageRecord.Failure == false {
			filteredImageRecordSlice = append(filteredImageRecordSlice, imageRecord)
		}
	}

	hasError := false
	buffer := bytes.Buffer{}

	if len(filteredImageRecordSlice) > 0 {
		imageRecord := filteredImageRecordSlice[0]
		privateRegistry, err := registry.GetPrivateRegistryFromPathAndTestAvailable(imageRecord.Path)
		if err != nil {
			hasError = true
			buffer.WriteString(err.Error())
		} else {
			err := privateRegistry.DeleteAllImageInRepository(imageRecord.ImageInformation)
			if err != nil {
				hasError = true
				buffer.WriteString(err.Error())
			}
		}

		err = RemoveImageFromAllHost(filteredImageRecordSlice)
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

	hasError := false
	buffer := bytes.Buffer{}

	deletedTagSlice := make([]string, 0)
	// Image is successfully built before
	if imageRecord.Failure == false {
		privateRegistry, err := registry.GetPrivateRegistryFromPathAndTestAvailable(imageRecord.Path)
		if err != nil {
			hasError = true
			buffer.WriteString(err.Error())
		} else {
			beforeDeleteTagSlice, err := privateRegistry.GetAllImageTag(imageRecord.ImageInformation)
			if err != nil {
				hasError = true
				buffer.WriteString(err.Error())
			} else {
				err := privateRegistry.DeleteImageInRepository(imageRecord.ImageInformation, imageRecord.Version)
				if err != nil {
					hasError = true
					buffer.WriteString(err.Error())
				} else {
					afterDeleteTagSlice, err := privateRegistry.GetAllImageTag(imageRecord.ImageInformation)
					if err != nil {
						hasError = true
						buffer.WriteString(err.Error())
					} else {
						for _, beforeDeleteTag := range beforeDeleteTagSlice {
							if isTagInSlice(beforeDeleteTag, afterDeleteTagSlice) == false {
								deletedTagSlice = append(deletedTagSlice, beforeDeleteTag)
							}
						}
					}
				}
			}
		}

		imageRecordSlice := make([]ImageRecord, 0)
		imageRecordSlice = append(imageRecordSlice, *imageRecord)
		err = RemoveImageFromAllHost(imageRecordSlice)
		if err != nil {
			hasError = true
			buffer.WriteString(err.Error())
		}
	} else {
		// Failure means no image is created and pushed to private-registry and saved in any host
		deletedTagSlice = append(deletedTagSlice, imageRecordVersion)
	}

	for _, deletedTag := range deletedTagSlice {
		err = GetStorage().DeleteImageRecord(imageInformationName, deletedTag)
		if err != nil {
			hasError = true
			buffer.WriteString(err.Error())
		}
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

func RemoveImageFromAllHost(imageRecordSlcie []ImageRecord) error {
	credentialSlice, err := host.GetStorage().LoadAllCredential()
	if err != nil {
		log.Error(err)
		return err
	}

	amount := len(imageRecordSlcie)

	commandSlice := make([]string, 0)
	// Delete all stopped instance so image could be removed
	commandSlice = append(commandSlice, "sudo docker rm $(sudo docker ps -aqf status=exited | xargs)\n")
	for _, imageRecord := range imageRecordSlcie {
		commandSlice = append(commandSlice, "sudo docker rmi "+imageRecord.Path+"\n")
	}

	hasError := false
	buffer := bytes.Buffer{}
	for _, credential := range credentialSlice {
		if credential.Disabled == false {
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

func isTagInSlice(targetTag string, tagSlice []string) bool {
	for _, tag := range tagSlice {
		if tag == targetTag {
			return true
		}
	}
	return false
}
