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

package sftp

import (
	"github.com/cloudawan/kubernetes_management/Godeps/_workspace/src/github.com/pkg/sftp"
	"github.com/cloudawan/kubernetes_management/Godeps/_workspace/src/golang.org/x/crypto/ssh"
	"os"
)

func DownLoadDirectoryRecurrsively(hostAndPort string, username string,
	password string, remoteSourceDirectory string, localTargetDirectory string) error {

	remoteSourceDirectoryLength := len(remoteSourceDirectory)

	authMethodSlice := make([]ssh.AuthMethod, 0)
	authMethodSlice = append(authMethodSlice, ssh.Password(password))

	clientConfig := ssh.ClientConfig{
		User: username,
		Auth: authMethodSlice,
	}
	connection, err := ssh.Dial("tcp", hostAndPort, &clientConfig)
	if err != nil {
		return err
	}
	defer connection.Close()

	// open an SFTP session over an existing ssh connection.
	client, err := sftp.NewClient(connection)
	if err != nil {
		return err
	}
	defer client.Close()

	// walk a directory
	walk := client.Walk(remoteSourceDirectory)
	for walk.Step() {
		if err := walk.Err(); err != nil {
			return err
		}

		if walk.Stat().IsDir() {
			directoryPath := localTargetDirectory + walk.Path()[remoteSourceDirectoryLength:]
			if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
				return err
			}
		} else {
			filePath := localTargetDirectory + walk.Path()[remoteSourceDirectoryLength:]
			file, err := client.Open(walk.Path())
			if err != nil {
				return err
			}
			defer file.Close()

			outputFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
			if err != nil {
				return err
			}
			defer outputFile.Close()

			_, err = file.WriteTo(outputFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
