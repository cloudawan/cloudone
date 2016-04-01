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

package notification

import (
	"bytes"
	"errors"
	"github.com/cloudawan/cloudone_utility/restclient"
	"net/url"
)

type SMSNexmo struct {
	Name      string
	Url       string
	APIKey    string
	APISecret string
}

func (smsNexmo *SMSNexmo) SendSMSNexmo(smsNexmoURL string, smsNexmoAPIKey string, smsNexmoAPISecret string,
	sender string, receiverNumberSlice []string, text string) error {

	hasError := false
	buffer := bytes.Buffer{}
	for _, receiverNumber := range receiverNumberSlice {
		convertedURL, err := url.Parse(smsNexmoURL)
		if err != nil {
			log.Error("Parse url %s error %s", smsNexmoURL, err)
			return err
		}

		parameters := url.Values{}
		parameters.Add("api_key", smsNexmoAPIKey)
		parameters.Add("api_secret", smsNexmoAPISecret)
		parameters.Add("from", sender)
		parameters.Add("to", receiverNumber)
		parameters.Add("text", text)
		convertedURL.RawQuery = parameters.Encode()

		result, err := restclient.RequestGet(convertedURL.String(), nil, true)
		bodyJsonMap, _ := result.(map[string]interface{})
		if err != nil {
			log.Error("Request url %s error bodyJsonMap %s error %s", convertedURL.String(), bodyJsonMap, err)
			hasError = true
			buffer.WriteString(err.Error())
		} else {
			log.Info("SMS send to %s, bodyJsonMap %s", convertedURL.String(), bodyJsonMap)
		}
	}

	if hasError {
		return errors.New(buffer.String())
	} else {
		return nil
	}
}

type NotifierSMSNexmo struct {
	Destination         string
	Sender              string
	ReceiverNumberSlice []string
}

func (notifierSMSNexmo NotifierSMSNexmo) notify(message string) error {
	smsNexmo, err := GetStorage().LoadSMSNexmo(notifierSMSNexmo.Destination)
	if err != nil {
		log.Error(err)
		return nil
	} else {
		return smsNexmo.SendSMSNexmo(smsNexmo.Url, smsNexmo.APIKey, smsNexmo.APISecret,
			notifierSMSNexmo.Sender, notifierSMSNexmo.ReceiverNumberSlice, message)
	}
}
