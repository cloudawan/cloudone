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
	"net/smtp"
	"strconv"
)

type EmailServerSMTP struct {
	Name     string
	Account  string
	Password string
	Host     string
	Port     int
}

func (emailServerSMTP *EmailServerSMTP) SendEmail(senderAccount string, senderPassword string, senderHost string,
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
	Destination          string
	ReceiverAccountSlice []string
}

func (notifierEmail NotifierEmail) notify(message string) error {
	emailServerSMTP, err := GetStorage().LoadEmailServerSMTP(notifierEmail.Destination)
	if err != nil {
		log.Error(err)
		return nil
	} else {
		return emailServerSMTP.SendEmail(
			emailServerSMTP.Account, emailServerSMTP.Password,
			emailServerSMTP.Host, emailServerSMTP.Port,
			notifierEmail.ReceiverAccountSlice, "Abnormal Notification", message)
	}
}
