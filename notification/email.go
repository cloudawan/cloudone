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
	"errors"
	"github.com/cloudawan/kubernetes_management/utility/configuration"
	"net/smtp"
	"strconv"
)

var senderAccount string
var senderPassword string
var senderHost string
var senderPort int

func init() {
	if err := ReloadEmail(); err != nil {
		panic(err)
	}
}

func ReloadEmail() error {
	var ok bool
	senderAccount, ok = configuration.LocalConfiguration.GetString("emailSenderAccount")
	if ok == false {
		log.Error("Fail to get email configuration senderAccount")
		return errors.New("Fail to get email configuration senderAccount")
	}
	senderPassword, ok = configuration.LocalConfiguration.GetString("emailSenderPassword")
	if ok == false {
		log.Error("Fail to get email configuration senderPassword")
		return errors.New("Fail to get email configuration senderPassword")
	}
	senderHost, ok = configuration.LocalConfiguration.GetString("emailSenderHost")
	if ok == false {
		log.Error("Fail to get email configuration senderHost")
		return errors.New("Fail to get email configuration senderHost")
	}
	senderPort, ok = configuration.LocalConfiguration.GetInt("emailSenderPort")
	if ok == false {
		log.Error("Fail to get email configuration senderPort")
		return errors.New("Fail to get email configuration senderPort")
	}
	return nil
}

func SendEmail(senderAccount string, senderPassword string, senderHost string,
	senderPort int, receiverAccountSlice []string, subject string, body string) error {
	// Set up authentication information.
	auth := smtp.PlainAuth(
		"",
		senderAccount,
		senderPassword,
		senderHost,
	)

	receiverAccountText := ""
	for _, receiverAccount := range receiverAccountSlice {
		receiverAccountText += receiverAccount + ","
	}

	message := "From: " + senderAccount + "\n" +
		"To: " + receiverAccountText + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(
		senderHost+":"+strconv.Itoa(senderPort),
		auth,
		senderAccount,
		receiverAccountSlice,
		[]byte(message),
	)
	return err
}

type NotifierEmail struct {
	ReceiverAccountSlice []string
}

func (notifierEmail NotifierEmail) notify(message string) error {
	return SendEmail(senderAccount, senderPassword, senderHost, senderPort,
		notifierEmail.ReceiverAccountSlice, "Abnormal Notification", message)
}
