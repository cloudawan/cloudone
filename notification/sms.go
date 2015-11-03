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
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/cloudawan/cloudone_utility/restclient"
	"net/url"
)

var smsNexmoURL string
var smsNexmoAPIKey string
var smsNexmoAPISecret string

func init() {
	if err := ReloadSMSNexmo(); err != nil {
		panic(err)
	}
}

func ReloadSMSNexmo() error {
	var ok bool
	smsNexmoURL, ok = configuration.LocalConfiguration.GetString("smsNexmoURL")
	if ok == false {
		log.Error("Fail to get sms configuration smsNexmoURL")
		return errors.New("Fail to get sms configuration smsNexmoURL")
	}
	smsNexmoAPIKey, ok = configuration.LocalConfiguration.GetString("smsNexmoAPIKey")
	if ok == false {
		log.Error("Fail to get sms configuration smsNexmoAPIKey")
		return errors.New("Fail to get sms configuration smsNexmoAPIKey")
	}
	smsNexmoAPISecret, ok = configuration.LocalConfiguration.GetString("smsNexmoAPISecret")
	if ok == false {
		log.Error("Fail to get sms configuration smsNexmoAPISecret")
		return errors.New("Fail to get sms configuration smsNexmoAPISecret")
	}
	return nil
}

func SendSMSNexmo(smsNexmoURL string, smsNexmoAPIKey string, smsNexmoAPISecret string,
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

		result, err := restclient.RequestGet(convertedURL.String(), true)
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
	Sender              string
	ReceiverNumberSlice []string
}

func (notifierSMSNexmo NotifierSMSNexmo) notify(message string) error {
	return SendSMSNexmo(smsNexmoURL, smsNexmoAPIKey, smsNexmoAPISecret,
		notifierSMSNexmo.Sender, notifierSMSNexmo.ReceiverNumberSlice, message)
}
