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

package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/cloudawan/cloudone/authorization"
	"github.com/cloudawan/cloudone/image"
	"github.com/coreos/etcd/client"
)

func getGitHashSignature(secret string, message string) string {
	key := []byte(secret)
	hash := hmac.New(sha1.New, key)
	hash.Write([]byte(message))

	return "sha1=" + hex.EncodeToString(hash.Sum(nil))
}

func Notify(username string, imageInformationName string, signature string, payload string) error {
	if len(username) == 0 {
		log.Error("User couldn't be empty. Signature %s", signature)
		log.Debug(payload)
		return errors.New("User couldn't be empty")
	}

	user, err := authorization.GetStorage().LoadUser(username)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return errors.New("The user " + username + " doesn't exist")
	}
	if err != nil {
		log.Error("Get user error %s. User name %s, signature %s", err, username, signature)
		log.Debug(payload)
		return err
	}

	if len(signature) > 0 {
		// If secret is used
		githubWebhookSecret := user.MetaDataMap["githubWebhookSecret"]
		generatedSignature := getGitHashSignature(githubWebhookSecret, payload)
		if generatedSignature != signature {
			log.Error("The signature is invalid. User name %s, signature %s, generated signature %s", username, signature, generatedSignature)
			log.Debug(payload)
			return errors.New("The signature is invalid")
		}
	}

	jsonMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(payload), &jsonMap)
	if err != nil {
		log.Error("Unmarshal payload error %s. User name %s, signature %s", err, username, signature)
		log.Debug(payload)
		return err
	}

	pusherJsonMap, _ := jsonMap["pusher"].(map[string]interface{})
	pusherName, _ := pusherJsonMap["name"].(string)

	repositoryJsonMap, _ := jsonMap["repository"].(map[string]interface{})
	cloneUrl, _ := repositoryJsonMap["clone_url"].(string)

	if len(cloneUrl) == 0 {
		log.Error("Can't find clone_url in github payload. User name %s, signature %s", username, signature)
		log.Debug(payload)
		return errors.New("Can't find clone_url in github payload")
	}

	imageInformation, err := image.GetStorage().LoadImageInformation(imageInformationName)
	etcdError, _ = err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return errors.New("The repository " + imageInformationName + " doesn't exist")
	}
	if err != nil {
		log.Error(err)
		return err
	}

	sourceCodeURL := imageInformation.BuildParameter["sourceCodeURL"]
	if sourceCodeURL != cloneUrl {
		// Not the target, ignore.
		return errors.New("Source code url " + sourceCodeURL + " is different from the github clone_url " + cloneUrl)
	}

	if len(imageInformationName) == 0 {
		log.Error("Can't find image information using the github url %s. User name %s, signature %s", cloneUrl, username, signature)
		log.Debug(payload)
		return errors.New("Can't find image information using the github url")
	}

	// Asyncronized build
	go func() {
		outputMessage, err := image.BuildUpgrade(imageInformationName, "Github webhook. Pusher: "+pusherName)
		if err != nil {
			log.Error(err)
			log.Debug(outputMessage)
		}
	}()

	return nil
}
