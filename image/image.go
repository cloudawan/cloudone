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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/cloudawan/cloudone/authorization"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/cloudawan/cloudone/utility/lock"
	"github.com/cloudawan/cloudone_utility/build"
	"github.com/cloudawan/cloudone_utility/filetransfer/sftp"
	"github.com/cloudawan/cloudone_utility/logger"
	"github.com/cloudawan/cloudone_utility/restclient"
	"github.com/sfreiberg/simplessh"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	LockKind = "image_information"
)

type ImageInformation struct {
	Name           string // "test"
	Kind           string // "git"
	Description    string // ""
	CurrentVersion string // ""
	BuildParameter map[string]string
}

type ImageRecord struct {
	ImageInformation string
	Version          string
	Path             string
	VersionInfo      map[string]string
	Environment      map[string]string
	Description      string
	CreatedTime      time.Time
	Failure          bool
}

func BuildCreate(imageInformation *ImageInformation) (returnedOutputMessage string, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("BuildCreate Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedOutputMessage = ""
			returnedError = err.(error)
		}
	}()

	var buildError error = nil
	imageRecord, outputMessage, err := Build(imageInformation, imageInformation.Description)
	if err != nil {
		log.Error("Build error: %s Output message: %s", err, outputMessage)
		if imageRecord == nil {
			return outputMessage, err
		} else {
			buildError = err
			outputMessage = outputMessage + "\n" + err.Error()
		}
	}

	// Save image record
	err = GetStorage().saveImageRecord(imageRecord)
	if err != nil {
		log.Error("Save image record error: %s", err)
		return outputMessage, err
	}

	// Save image information with version
	imageInformation.CurrentVersion = imageRecord.Version
	err = GetStorage().SaveImageInformation(imageInformation)
	if err != nil {
		log.Error("Update image information error: %s", err)
		return outputMessage, err
	}

	// Save build log
	err = SendBuildLog(imageRecord, outputMessage)
	if err != nil {
		log.Error("Save build log error: %s", err)
		return outputMessage, err
	}

	return outputMessage, buildError
}

func BuildUpgrade(imageInformationName string, description string) (returnedOutputMessage string, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("BuildUpgrade Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedOutputMessage = ""
			returnedError = err.(error)
		}
	}()

	imageInformation, err := GetStorage().LoadImageInformation(imageInformationName)
	if err != nil {
		log.Error("Load image information error: %s", err)
		return "", err
	}

	var buildError error = nil
	imageRecord, outputMessage, err := Build(imageInformation, description)
	if err != nil {
		log.Error("Build error: %s Output message: %s", err, outputMessage)
		if imageRecord == nil {
			return outputMessage, err
		} else {
			buildError = err
			outputMessage = outputMessage + "\n" + err.Error()
		}
	}

	// Save image record
	err = GetStorage().saveImageRecord(imageRecord)
	if err != nil {
		log.Error("Save image record error: %s", err)
		return outputMessage, err
	}

	imageInformation.CurrentVersion = imageRecord.Version
	// Save image information
	err = GetStorage().SaveImageInformation(imageInformation)
	if err != nil {
		log.Error("Save image information error: %s", err)
		return outputMessage, err
	}

	// Save build log
	err = SendBuildLog(imageRecord, outputMessage)
	if err != nil {
		log.Error("Save build log error: %s", err)
		return outputMessage, err
	}

	return outputMessage, buildError
}

func Build(imageInformation *ImageInformation, description string) (*ImageRecord, string, error) {
	if lock.AcquireLock(LockKind, imageInformation.Name, 0) == false {
		log.Error("Image %s is under construction", imageInformation.Name)
		return nil, "", errors.New("Image is under construction")
	}
	defer lock.ReleaseLock(LockKind, imageInformation.Name)

	switch imageInformation.Kind {
	case "git":
		return BuildFromGit(imageInformation, description)
	case "scp":
		return BuildFromSCP(imageInformation, description)
	case "sftp":
		return BuildFromSFTP(imageInformation, description)
	default:
		return nil, "", errors.New("No such kind: " + imageInformation.Kind)
	}
}

func BuildFromGit(imageInformation *ImageInformation, description string) (*ImageRecord, string, error) {
	// OutputBuffer for the whole result
	// OutputFile for external tail
	outputBuffer := &bytes.Buffer{}
	outputFile, err := os.Create(GetProcessingOutMessageFilePathAndName(imageInformation.Name))
	defer func() {
		outputFile.Close()
		// For websocket to have time to read
		time.Sleep(time.Second)
		// Remove the output file used for tail during the process
		os.Remove(GetProcessingOutMessageFilePathAndName(imageInformation.Name))
	}()

	imageRecord := &ImageRecord{}
	imageRecord.ImageInformation = imageInformation.Name
	imageRecord.VersionInfo = make(map[string]string)
	imageRecord.CreatedTime = time.Now()
	imageRecord.Version = imageRecord.CreatedTime.Format("2006-01-02-15-04-05")
	imageRecord.Description = description
	imageRecord.Failure = true

	// Build parameter
	workingDirectory := imageInformation.BuildParameter["workingDirectory"]         // "/var/lib/cloudone"
	repositoryPath := imageInformation.BuildParameter["repositoryPath"]             // "private-repository:31000/test"
	sourceCodeURL := imageInformation.BuildParameter["sourceCodeURL"]               // "https://github.com/cloudawan/test.git"
	sourceCodeProject := imageInformation.BuildParameter["sourceCodeProject"]       // "test"
	sourceCodeDirectory := imageInformation.BuildParameter["sourceCodeDirectory"]   // "src"
	sourceCodeMakeScript := imageInformation.BuildParameter["sourceCodeMakeScript"] // ""
	versionFile := imageInformation.BuildParameter["versionFile"]                   // "version"
	environmentFile := imageInformation.BuildParameter["environmentFile"]           // "environment"

	// Path
	imageRecord.Path = repositoryPath + ":" + imageRecord.Version

	// Clean the previous work space if existing
	os.RemoveAll(workingDirectory)
	// Check working space
	if _, err := os.Stat(workingDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(workingDirectory, os.ModePerm)
		if err != nil {
			log.Error("Create non-existing directory %s error: %s", workingDirectory, err)
			outputBuffer.WriteString("The error phase: Check working space\n")
			outputBuffer.WriteString("Create non-existing directory " + workingDirectory + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Check working space\n")
			outputFile.WriteString("Create non-existing directory " + workingDirectory + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}
	}

	// First time
	sourceCodeURLWithUserAndPasword := ""
	if strings.Contains(sourceCodeURL, "@") {
		// If already containing user and password
	} else {
		// If the repository doesn't exist or authorized, git clone will prompt for password and suspend the command so use fake user and password to fail it.
		sourceCodeURLWithUserAndPasword = "https://username:password@" + sourceCodeURL[len("https://"):]
	}
	command := exec.Command("git", "clone", sourceCodeURLWithUserAndPasword)
	command.Dir = workingDirectory
	_, _, err = executeCommandAndTailTheOutput(command, outputBuffer, outputFile)
	if err != nil {
		log.Error("Git clone %s error: %s", sourceCodeURL, err)
		outputBuffer.WriteString("The error phase: Git clone\n")
		outputBuffer.WriteString("Git clone " + sourceCodeURL + " error: " + err.Error() + "\n")
		outputFile.WriteString("The error phase: Git clone\n")
		outputFile.WriteString("Git clone " + sourceCodeURL + " error: " + err.Error() + "\n")
		return imageRecord, outputBuffer.String(), err
	}

	command = exec.Command("git", "pull")
	command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
	_, _, err = executeCommandAndTailTheOutput(command, outputBuffer, outputFile)
	if err != nil {
		log.Error("Git pull %s error: %s", sourceCodeURL, err)
		outputBuffer.WriteString("The error phase: Git pull\n")
		outputBuffer.WriteString("Git pull " + sourceCodeURL + " error: " + err.Error() + "\n")
		outputFile.WriteString("The error phase: Git pull\n")
		outputFile.WriteString("Git pull " + sourceCodeURL + " error: " + err.Error() + "\n")
		return imageRecord, outputBuffer.String(), err
	}

	// Get git version
	command = exec.Command("git", "log", "-1")
	command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
	outputText, _, err := executeCommandAndTailTheOutput(command, outputBuffer, outputFile)
	if err != nil {
		log.Error("Git log %s -l error: %s", imageInformation, err)
		outputBuffer.WriteString("The error phase: Git log\n")
		outputBuffer.WriteString("Git log " + sourceCodeURL + " error: " + err.Error() + "\n")
		outputFile.WriteString("The error phase: Git log\n")
		outputFile.WriteString("Git log " + sourceCodeURL + " error: " + err.Error() + "\n")
		return imageRecord, outputBuffer.String(), err
	}

	gitVersionSlice, err := parseGitVersion(outputText)
	if gitVersionSlice != nil && len(gitVersionSlice) > 0 {
		currentGitVersion := gitVersionSlice[0]
		imageRecord.VersionInfo["Commit"] = currentGitVersion.Commit
		imageRecord.VersionInfo["Autor"] = currentGitVersion.Autor
		imageRecord.VersionInfo["Date"] = currentGitVersion.Date
	}

	if sourceCodeMakeScript != "" {
		commandSlice := strings.Split(sourceCodeMakeScript, " ")
		if len(commandSlice) == 1 {
			command = exec.Command(sourceCodeMakeScript)
		} else {
			command = exec.Command(commandSlice[0], commandSlice[1:]...)
		}
		command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
		_, _, err = executeCommandAndTailTheOutput(command, outputBuffer, outputFile)
		if err != nil {
			log.Error("Run make script %s error: %s", sourceCodeMakeScript, err)
			outputBuffer.WriteString("The error phase: Run make script\n")
			outputBuffer.WriteString("Run make script " + sourceCodeMakeScript + " on " + command.Dir + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Run make script\n")
			outputFile.WriteString("Run make script " + sourceCodeMakeScript + " on " + command.Dir + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}
	}

	// User defined version
	version := ""
	if versionFile != "" {
		// open input file
		inputFile, err := os.Open(workingDirectory + string(os.PathSeparator) +
			sourceCodeProject + string(os.PathSeparator) + versionFile)
		if err != nil {
			log.Error("Open version file %s error: %s", versionFile, err)
			outputBuffer.WriteString("The error phase: Open version file\n")
			outputBuffer.WriteString("Open version file " + versionFile + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Open version file\n")
			outputFile.WriteString("Open version file " + versionFile + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}
		defer inputFile.Close()

		byteSlice := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			// read a chunk
			n, err := inputFile.Read(buffer)
			if err != nil && err != io.EOF {
				log.Error("Read version file %s error: %s", versionFile, err)
				outputBuffer.WriteString("The error phase: Read version file\n")
				outputBuffer.WriteString("Read version file " + versionFile + " error: " + err.Error() + "\n")
				outputFile.WriteString("The error phase: Read version file\n")
				outputFile.WriteString("Read version file " + versionFile + " error: " + err.Error() + "\n")
				return imageRecord, outputBuffer.String(), err
			}
			if n == 0 {
				break
			}

			byteSlice = append(byteSlice, buffer[0:n]...)
		}

		version = string(byteSlice)
	}
	imageRecord.VersionInfo["Version"] = version

	// Environment map
	environmentMap := make(map[string]string)
	if environmentFile != "" {
		// open input file
		inputFile, err := os.Open(workingDirectory + string(os.PathSeparator) +
			sourceCodeProject + string(os.PathSeparator) + environmentFile)
		if err != nil {
			log.Error("Open environment file %s error: %s", environmentFile, err)
			outputBuffer.WriteString("The error phase: Open environment file\n")
			outputBuffer.WriteString("Open environment file " + environmentFile + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Open environment file\n")
			outputFile.WriteString("Open environment file " + environmentFile + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}
		defer inputFile.Close()

		byteSlice := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			// read a chunk
			n, err := inputFile.Read(buffer)
			if err != nil && err != io.EOF {
				log.Error("Read environment file %s error: %s", environmentFile, err)
				outputBuffer.WriteString("The error phase: Read environment file\n")
				outputBuffer.WriteString("Read environment file " + environmentFile + " error: " + err.Error() + "\n")
				outputFile.WriteString("The error phase: Read environment file\n")
				outputFile.WriteString("Read environment file " + environmentFile + " error: " + err.Error() + "\n")
				return imageRecord, outputBuffer.String(), err
			}
			if n == 0 {
				break
			}

			byteSlice = append(byteSlice, buffer[0:n]...)
		}

		jsonMap := make(map[string]interface{})
		err = json.Unmarshal(byteSlice, &jsonMap)
		if err != nil {
			log.Error("Unmarshal environment file %s error: %s", environmentFile, err)
			outputBuffer.WriteString("The error phase: Unmarshal environment file\n")
			outputBuffer.WriteString("Unmarshal environment file " + environmentFile + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Unmarshal environment file\n")
			outputFile.WriteString("Unmarshal environment file " + environmentFile + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}

		for key, value := range jsonMap {
			environmentDescription, ok := value.(string)
			if ok {
				environmentMap[key] = environmentDescription
			}
		}
	}
	imageRecord.Environment = environmentMap

	command = exec.Command("docker", "build", "-t", imageRecord.Path, sourceCodeDirectory)
	command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
	_, _, err = executeCommandAndTailTheOutput(command, outputBuffer, outputFile)
	if err != nil {
		log.Error("Docker build %s error: %s", imageRecord.Path, err)
		outputBuffer.WriteString("The error phase: Docker build\n")
		outputBuffer.WriteString("Docker build " + imageRecord.Path + " error: " + err.Error() + "\n")
		outputFile.WriteString("The error phase: Docker build\n")
		outputFile.WriteString("Docker build " + imageRecord.Path + " error: " + err.Error() + "\n")
		return imageRecord, outputBuffer.String(), err
	}

	command = exec.Command("docker", "push", imageRecord.Path)
	command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
	_, _, err = executeCommandAndTailTheOutput(command, outputBuffer, outputFile)
	if err != nil {
		log.Error("Docker push %s error: %s", imageRecord.Path, err)
		outputBuffer.WriteString("The error phase: Docker push\n")
		outputBuffer.WriteString("Docker push " + imageRecord.Path + " error: " + err.Error() + "\n")
		outputFile.WriteString("The error phase: Docker push\n")
		outputFile.WriteString("Docker push " + imageRecord.Path + " error: " + err.Error() + "\n")
		return imageRecord, outputBuffer.String(), err
	}

	// Success
	imageRecord.Failure = false

	// Only remove working space if it is successful so user could check the failed data
	if imageRecord.Failure == false {
		os.RemoveAll(workingDirectory)
	}

	return imageRecord, outputBuffer.String(), nil
}

type GitVersion struct {
	Commit string
	Autor  string
	Date   string
}

func parseGitVersion(text string) ([]GitVersion, error) {
	commitSlice := make([]string, 0)
	authorSlice := make([]string, 0)
	dateSlice := make([]string, 0)

	lineSlice := strings.Split(text, "\n")
	for _, line := range lineSlice {
		if strings.HasPrefix(line, "commit") {
			commitSlice = append(commitSlice, line[7:len(line)])
		} else if strings.HasPrefix(line, "Author") {
			authorSlice = append(authorSlice, line[8:len(line)])
		} else if strings.HasPrefix(line, "Date") {
			dateSlice = append(dateSlice, line[8:len(line)])
		}
	}

	if len(commitSlice) != len(authorSlice) || len(commitSlice) != len(dateSlice) {
		return nil, errors.New("Length of commitSlice " + strconv.Itoa(len(commitSlice)) + ", authorSlice " + strconv.Itoa(len(authorSlice)) + ", dateSlice " + strconv.Itoa(len(dateSlice)) + " are different")
	}

	gitVersionSlice := make([]GitVersion, 0)
	for i, commit := range commitSlice {
		gitVersionSlice = append(gitVersionSlice, GitVersion{commit, authorSlice[i], dateSlice[i]})
	}

	return gitVersionSlice, nil
}

var scpTimeout time.Duration = time.Second * 10

func BuildFromSCP(imageInformation *ImageInformation, description string) (*ImageRecord, string, error) {
	// OutputBuffer for the whole result
	// OutputFile for external tail
	outputBuffer := &bytes.Buffer{}
	outputFile, err := os.Create(GetProcessingOutMessageFilePathAndName(imageInformation.Name))
	defer func() {
		outputFile.Close()
		// For websocket to have time to read
		time.Sleep(time.Second)
		// Remove the output file used for tail during the process
		os.Remove(GetProcessingOutMessageFilePathAndName(imageInformation.Name))
	}()

	imageRecord := &ImageRecord{}
	imageRecord.ImageInformation = imageInformation.Name
	imageRecord.VersionInfo = make(map[string]string)
	imageRecord.CreatedTime = time.Now()
	imageRecord.Version = imageRecord.CreatedTime.Format("2006-01-02-15-04-05")
	imageRecord.Description = description
	imageRecord.Failure = true

	// Build parameter
	workingDirectory := imageInformation.BuildParameter["workingDirectory"]         // "/var/lib/cloudone"
	repositoryPath := imageInformation.BuildParameter["repositoryPath"]             // "private-repository:31000/test"
	hostAndPort := imageInformation.BuildParameter["hostAndPort"]                   // "172.16.0.113:22"
	username := imageInformation.BuildParameter["username"]                         // "cloudawan"
	password := imageInformation.BuildParameter["password"]                         // "cloud4win"
	sourcePath := imageInformation.BuildParameter["sourcePath"]                     // "/home/cloudawan/test_scp"
	compressFileName := imageInformation.BuildParameter["compressFileName"]         //"tp.tar.gz"
	unpackageCommand := imageInformation.BuildParameter["unpackageCommand"]         // "tar zxvf"
	sourceCodeProject := imageInformation.BuildParameter["sourceCodeProject"]       // "tp"
	sourceCodeDirectory := imageInformation.BuildParameter["sourceCodeDirectory"]   // "src"
	sourceCodeMakeScript := imageInformation.BuildParameter["sourceCodeMakeScript"] // ""
	versionFile := imageInformation.BuildParameter["versionFile"]                   // "version"
	environmentFile := imageInformation.BuildParameter["environmentFile"]           // "environment"

	// Path
	imageRecord.Path = repositoryPath + ":" + imageRecord.Version

	// Clean the previous work space if existing
	os.RemoveAll(workingDirectory)
	// Check working space
	if _, err := os.Stat(workingDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(workingDirectory, os.ModePerm)
		if err != nil {
			log.Error("Create non-existing directory %s error: %s", workingDirectory, err)
			outputBuffer.WriteString("The error phase: Check working space\n")
			outputBuffer.WriteString("Create non-existing directory " + workingDirectory + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Check working space\n")
			outputFile.WriteString("Create non-existing directory " + workingDirectory + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}
	}

	client, err := simplessh.ConnectWithPasswordTimeout(hostAndPort, username, password, scpTimeout)
	if err != nil {
		log.Error("Login scp hostAndPort %s, username %s, password %s error: %s", hostAndPort, username, password, err)
		outputBuffer.WriteString("The error phase: Login scp\n")
		outputBuffer.WriteString("Login scp " + hostAndPort + " error: " + err.Error() + "\n")
		outputFile.WriteString("The error phase: Login scp\n")
		outputFile.WriteString("Login scp " + hostAndPort + " error: " + err.Error() + "\n")
		return imageRecord, outputBuffer.String(), err
	}
	defer client.Close()

	remoteFilePath := sourcePath + string(os.PathSeparator) + compressFileName
	localFilePath := workingDirectory + string(os.PathSeparator) + compressFileName
	if err := client.Download(remoteFilePath, localFilePath); err != nil {
		log.Error("Download remoteFilePath %s localFilePath %s with scp error: %s", remoteFilePath, localFilePath, err)
		outputBuffer.WriteString("The error phase: Download\n")
		outputBuffer.WriteString("Download " + remoteFilePath + " error: " + err.Error() + "\n")
		outputFile.WriteString("The error phase: Download\n")
		outputFile.WriteString("Download " + remoteFilePath + " error: " + err.Error() + "\n")
		return imageRecord, outputBuffer.String(), err
	}

	unpackageCommandSlice := strings.Split(unpackageCommand, " ")
	unpackageCommandSlice = append(unpackageCommandSlice, compressFileName)

	command := exec.Command(unpackageCommandSlice[0], unpackageCommandSlice[1:]...)
	command.Dir = workingDirectory
	_, _, err = executeCommandAndTailTheOutput(command, outputBuffer, outputFile)
	if err != nil {
		log.Error("Unpackage compress file %s error: %s", unpackageCommand, err)
		outputBuffer.WriteString("The error phase: Unpackage compress\n")
		outputBuffer.WriteString("Unpackage compress " + unpackageCommand + " error: " + err.Error() + "\n")
		outputFile.WriteString("The error phase: Unpackage compress\n")
		outputFile.WriteString("Unpackage compress " + unpackageCommand + " error: " + err.Error() + "\n")
		return imageRecord, outputBuffer.String(), err
	}

	if sourceCodeMakeScript != "" {
		command = exec.Command(sourceCodeMakeScript)
		command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
		_, _, err = executeCommandAndTailTheOutput(command, outputBuffer, outputFile)
		if err != nil {
			log.Error("Run make script %s error: %s", sourceCodeMakeScript, err)
			outputBuffer.WriteString("The error phase: Run make script\n")
			outputBuffer.WriteString("Run make script " + sourceCodeMakeScript + " on " + command.Dir + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Run make script\n")
			outputFile.WriteString("Run make script " + sourceCodeMakeScript + " on " + command.Dir + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}
	}

	// User defined version
	version := ""
	if versionFile != "" {
		// open input file
		inputFile, err := os.Open(workingDirectory + string(os.PathSeparator) +
			sourceCodeProject + string(os.PathSeparator) + versionFile)
		if err != nil {
			log.Error("Open version file %s error: %s", versionFile, err)
			outputBuffer.WriteString("The error phase: Open version file\n")
			outputBuffer.WriteString("Open version file " + versionFile + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Open version file\n")
			outputFile.WriteString("Open version file " + versionFile + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}
		defer inputFile.Close()

		byteSlice := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			// read a chunk
			n, err := inputFile.Read(buffer)
			if err != nil && err != io.EOF {
				log.Error("Read version file %s error: %s", versionFile, err)
				outputBuffer.WriteString("The error phase: Read version file\n")
				outputBuffer.WriteString("Read version file " + versionFile + " error: " + err.Error() + "\n")
				outputFile.WriteString("The error phase: Read version file\n")
				outputFile.WriteString("Read version file " + versionFile + " error: " + err.Error() + "\n")
				return imageRecord, outputBuffer.String(), err
			}
			if n == 0 {
				break
			}

			byteSlice = append(byteSlice, buffer[0:n]...)
		}

		version = string(byteSlice)
	}
	imageRecord.VersionInfo["Version"] = version

	// Environment map
	environmentMap := make(map[string]string)
	if environmentFile != "" {
		// open input file
		inputFile, err := os.Open(workingDirectory + string(os.PathSeparator) +
			sourceCodeProject + string(os.PathSeparator) + environmentFile)
		if err != nil {
			log.Error("Open environment file %s error: %s", environmentFile, err)
			outputBuffer.WriteString("The error phase: Open environment file\n")
			outputBuffer.WriteString("Open environment file " + environmentFile + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Open environment file\n")
			outputFile.WriteString("Open environment file " + environmentFile + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}
		defer inputFile.Close()

		byteSlice := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			// read a chunk
			n, err := inputFile.Read(buffer)
			if err != nil && err != io.EOF {
				log.Error("Read environment file %s error: %s", environmentFile, err)
				outputBuffer.WriteString("The error phase: Read environment file\n")
				outputBuffer.WriteString("Read environment file " + environmentFile + " error: " + err.Error() + "\n")
				outputFile.WriteString("The error phase: Read environment file\n")
				outputFile.WriteString("Read environment file " + environmentFile + " error: " + err.Error() + "\n")
				return imageRecord, outputBuffer.String(), err
			}
			if n == 0 {
				break
			}

			byteSlice = append(byteSlice, buffer[0:n]...)
		}

		jsonMap := make(map[string]interface{})
		err = json.Unmarshal(byteSlice, &jsonMap)
		if err != nil {
			log.Error("Unmarshal environment file %s error: %s", environmentFile, err)
			outputBuffer.WriteString("The error phase: Unmarshal environment file\n")
			outputBuffer.WriteString("Unmarshal environment file " + environmentFile + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Unmarshal environment file\n")
			outputFile.WriteString("Unmarshal environment file " + environmentFile + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}

		for key, value := range jsonMap {
			description, ok := value.(string)
			if ok {
				environmentMap[key] = description
			}
		}
	}
	imageRecord.Environment = environmentMap

	command = exec.Command("docker", "build", "-t", imageRecord.Path, sourceCodeDirectory)
	command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
	_, _, err = executeCommandAndTailTheOutput(command, outputBuffer, outputFile)
	if err != nil {
		log.Error("Docker build %s error: %s", imageRecord.Path, err)
		outputBuffer.WriteString("The error phase: Docker build\n")
		outputBuffer.WriteString("Docker build " + imageRecord.Path + " error: " + err.Error() + "\n")
		outputFile.WriteString("The error phase: Docker build\n")
		outputFile.WriteString("Docker build " + imageRecord.Path + " error: " + err.Error() + "\n")
		return imageRecord, outputBuffer.String(), err
	}

	command = exec.Command("docker", "push", imageRecord.Path)
	command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
	_, _, err = executeCommandAndTailTheOutput(command, outputBuffer, outputFile)
	if err != nil {
		log.Error("Docker push %s error: %s", imageRecord.Path, err)
		outputBuffer.WriteString("The error phase: Docker push\n")
		outputBuffer.WriteString("Docker push " + imageRecord.Path + " error: " + err.Error() + "\n")
		outputFile.WriteString("The error phase: Docker push\n")
		outputFile.WriteString("Docker push " + imageRecord.Path + " error: " + err.Error() + "\n")
		return imageRecord, outputBuffer.String(), err
	}

	// Success
	imageRecord.Failure = false

	// Only remove working space if it is successful so user could check the failed data
	if imageRecord.Failure == false {
		os.RemoveAll(workingDirectory)
	}

	return imageRecord, outputBuffer.String(), nil
}

func BuildFromSFTP(imageInformation *ImageInformation, description string) (*ImageRecord, string, error) {
	// OutputBuffer for the whole result
	// OutputFile for external tail
	outputBuffer := &bytes.Buffer{}
	outputFile, err := os.Create(GetProcessingOutMessageFilePathAndName(imageInformation.Name))
	defer func() {
		outputFile.Close()
		// For websocket to have time to read
		time.Sleep(time.Second)
		// Remove the output file used for tail during the process
		os.Remove(GetProcessingOutMessageFilePathAndName(imageInformation.Name))
	}()

	imageRecord := &ImageRecord{}
	imageRecord.ImageInformation = imageInformation.Name
	imageRecord.VersionInfo = make(map[string]string)
	imageRecord.CreatedTime = time.Now()
	imageRecord.Version = imageRecord.CreatedTime.Format("2006-01-02-15-04-05")
	imageRecord.Description = description
	imageRecord.Failure = true

	// Build parameter
	workingDirectory := imageInformation.BuildParameter["workingDirectory"]         // "/var/lib/cloudone"
	repositoryPath := imageInformation.BuildParameter["repositoryPath"]             // "private-repository:31000/test"
	hostAndPort := imageInformation.BuildParameter["hostAndPort"]                   // "172.16.0.113:22"
	username := imageInformation.BuildParameter["username"]                         // "cloudawan"
	password := imageInformation.BuildParameter["password"]                         // "cloud4win"
	sourcePath := imageInformation.BuildParameter["sourcePath"]                     // "/home/cloudawan/test_sftp"
	sourceCodeProject := imageInformation.BuildParameter["sourceCodeProject"]       // "tp"
	sourceCodeDirectory := imageInformation.BuildParameter["sourceCodeDirectory"]   // "src"
	sourceCodeMakeScript := imageInformation.BuildParameter["sourceCodeMakeScript"] // ""
	versionFile := imageInformation.BuildParameter["versionFile"]                   // "version"
	environmentFile := imageInformation.BuildParameter["environmentFile"]           // "environment"

	// Path
	imageRecord.Path = repositoryPath + ":" + imageRecord.Version

	// Clean the previous work space if existing
	os.RemoveAll(workingDirectory)
	// Check working space
	if _, err := os.Stat(workingDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(workingDirectory, os.ModePerm)
		if err != nil {
			log.Error("Create non-existing directory %s error: %s", workingDirectory, err)
			outputBuffer.WriteString("The error phase: Check working space\n")
			outputBuffer.WriteString("Create non-existing directory " + workingDirectory + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Check working space\n")
			outputFile.WriteString("Create non-existing directory " + workingDirectory + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}
	}

	if err := sftp.DownLoadDirectoryRecurrsively(hostAndPort, username,
		password, sourcePath, workingDirectory); err != nil {
		log.Error("Download from sftp hostAndPort %s, username %s, password %s, sourcePath %s, workingDirectory %s error: %s",
			hostAndPort, username, password, sourcePath, workingDirectory, err)
		outputBuffer.WriteString("The error phase: Download from sftp\n")
		outputBuffer.WriteString("Download from sftp " + hostAndPort + " error: " + err.Error() + "\n")
		outputFile.WriteString("The error phase: Download from sftp\n")
		outputFile.WriteString("Download from sftp " + hostAndPort + " error: " + err.Error() + "\n")
		return imageRecord, "", err
	}

	if sourceCodeMakeScript != "" {
		command := exec.Command(sourceCodeMakeScript)
		command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
		_, _, err = executeCommandAndTailTheOutput(command, outputBuffer, outputFile)
		if err != nil {
			log.Error("Run make script %s error: %s", sourceCodeMakeScript, err)
			outputBuffer.WriteString("The error phase: Run make script\n")
			outputBuffer.WriteString("Run make script " + sourceCodeMakeScript + " on " + command.Dir + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Run make script\n")
			outputFile.WriteString("Run make script " + sourceCodeMakeScript + " on " + command.Dir + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}
	}

	// User defined version
	version := ""
	if versionFile != "" {
		// open input file
		inputFile, err := os.Open(workingDirectory + string(os.PathSeparator) +
			sourceCodeProject + string(os.PathSeparator) + versionFile)
		if err != nil {
			log.Error("Open version file %s error: %s", versionFile, err)
			outputBuffer.WriteString("The error phase: Open version file\n")
			outputBuffer.WriteString("Open version file " + versionFile + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Open version file\n")
			outputFile.WriteString("Open version file " + versionFile + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}
		defer inputFile.Close()

		byteSlice := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			// read a chunk
			n, err := inputFile.Read(buffer)
			if err != nil && err != io.EOF {
				log.Error("Read version file %s error: %s", versionFile, err)
				outputBuffer.WriteString("The error phase: Read version file\n")
				outputBuffer.WriteString("Read version file " + versionFile + " error: " + err.Error() + "\n")
				outputFile.WriteString("The error phase: Read version file\n")
				outputFile.WriteString("Read version file " + versionFile + " error: " + err.Error() + "\n")
				return imageRecord, outputBuffer.String(), err
			}
			if n == 0 {
				break
			}

			byteSlice = append(byteSlice, buffer[0:n]...)
		}

		version = string(byteSlice)
	}
	imageRecord.VersionInfo["Version"] = version

	// Environment map
	environmentMap := make(map[string]string)
	if environmentFile != "" {
		// open input file
		inputFile, err := os.Open(workingDirectory + string(os.PathSeparator) +
			sourceCodeProject + string(os.PathSeparator) + environmentFile)
		if err != nil {
			log.Error("Open environment file %s error: %s", environmentFile, err)
			outputBuffer.WriteString("The error phase: Open environment file\n")
			outputBuffer.WriteString("Open environment file " + environmentFile + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Open environment file\n")
			outputFile.WriteString("Open environment file " + environmentFile + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}
		defer inputFile.Close()

		byteSlice := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			// read a chunk
			n, err := inputFile.Read(buffer)
			if err != nil && err != io.EOF {
				log.Error("Read environment file %s error: %s", environmentFile, err)
				outputBuffer.WriteString("The error phase: Read environment file\n")
				outputBuffer.WriteString("Read environment file " + environmentFile + " error: " + err.Error() + "\n")
				outputFile.WriteString("The error phase: Read environment file\n")
				outputFile.WriteString("Read environment file " + environmentFile + " error: " + err.Error() + "\n")
				return imageRecord, outputBuffer.String(), err
			}
			if n == 0 {
				break
			}

			byteSlice = append(byteSlice, buffer[0:n]...)
		}

		jsonMap := make(map[string]interface{})
		err = json.Unmarshal(byteSlice, &jsonMap)
		if err != nil {
			log.Error("Unmarshal environment file %s error: %s", environmentFile, err)
			outputBuffer.WriteString("The error phase: Unmarshal environment file\n")
			outputBuffer.WriteString("Unmarshal environment file " + environmentFile + " error: " + err.Error() + "\n")
			outputFile.WriteString("The error phase: Unmarshal environment file\n")
			outputFile.WriteString("Unmarshal environment file " + environmentFile + " error: " + err.Error() + "\n")
			return imageRecord, outputBuffer.String(), err
		}

		for key, value := range jsonMap {
			description, ok := value.(string)
			if ok {
				environmentMap[key] = description
			}
		}
	}
	imageRecord.Environment = environmentMap

	command := exec.Command("docker", "build", "-t", imageRecord.Path, sourceCodeDirectory)
	command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
	_, _, err = executeCommandAndTailTheOutput(command, outputBuffer, outputFile)
	if err != nil {
		log.Error("Docker build %s error: %s", imageRecord.Path, err)
		outputBuffer.WriteString("The error phase: Docker build\n")
		outputBuffer.WriteString("Docker build " + imageRecord.Path + " error: " + err.Error() + "\n")
		outputFile.WriteString("The error phase: Docker build\n")
		outputFile.WriteString("Docker build " + imageRecord.Path + " error: " + err.Error() + "\n")
		return imageRecord, outputBuffer.String(), err
	}

	command = exec.Command("docker", "push", imageRecord.Path)
	command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
	_, _, err = executeCommandAndTailTheOutput(command, outputBuffer, outputFile)
	if err != nil {
		log.Error("Docker push %s error: %s", imageRecord.Path, err)
		outputBuffer.WriteString("The error phase: Docker push\n")
		outputBuffer.WriteString("Docker push " + imageRecord.Path + " error: " + err.Error() + "\n")
		outputFile.WriteString("The error phase: Docker push\n")
		outputFile.WriteString("Docker push " + imageRecord.Path + " error: " + err.Error() + "\n")
		return imageRecord, outputBuffer.String(), err
	}

	// Success
	imageRecord.Failure = false

	// Only remove working space if it is successful so user could check the failed data
	if imageRecord.Failure == false {
		os.RemoveAll(workingDirectory)
	}

	return imageRecord, outputBuffer.String(), nil
}

func SendBuildLog(imageRecord *ImageRecord, outputMessage string) error {
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

	buildLog := build.BuildLog{
		imageRecord.ImageInformation,
		imageRecord.Version,
		imageRecord.VersionInfo,
		imageRecord.CreatedTime,
		outputMessage,
	}

	url := "https://" + cloudoneAnalysisHost + ":" + strconv.Itoa(cloudoneAnalysisPort) + "/api/v1/buildlogs"

	headerMap := make(map[string]string)
	headerMap["token"] = authorization.SystemAdminToken

	_, err := restclient.RequestPost(url, buildLog, headerMap, false)
	if err != nil {
		log.Error("Fail to send build log %v with error %s", buildLog, err)
	}

	return err
}

const (
	processOutMessageFilePathAndNamePrefix = "/tmp/processingBuildLog"
)

func TouchOutMessageFile(imageInformation string) error {
	outputFile, err := os.Create(GetProcessingOutMessageFilePathAndName(imageInformation))
	if err != nil {
		log.Error(err)
		return err
	}
	outputFile.WriteString("\n")
	outputFile.Close()
	return nil
}

func GetProcessingOutMessageFilePathAndName(imageInformationName string) string {
	return processOutMessageFilePathAndNamePrefix + imageInformationName
}

func executeCommandAndTailTheOutput(command *exec.Cmd, outputBuffer *bytes.Buffer, outputFile *os.File) (string, int, error) {
	if command == nil {
		log.Error("Command can't be nil")
		return "", 0, errors.New("Command can't be nil")
	}

	buffer := bytes.Buffer{}

	readCloser, err := command.StdoutPipe()
	if err != nil {
		log.Error(err)
		return "", 0, err
	}
	scanner := bufio.NewScanner(readCloser)

	go func() {
		for scanner.Scan() {
			byteSlice := scanner.Bytes()

			buffer.Write(byteSlice)
			buffer.WriteByte('\n')

			if outputBuffer != nil {
				outputBuffer.Write(byteSlice)
				outputBuffer.WriteByte('\n')
			}
			if outputFile != nil {
				outputFile.Write(byteSlice)
				outputFile.WriteString("\n")
			}
		}
	}()

	err = command.Start()
	if err != nil {
		log.Error(err)
		return "", 0, err
	}

	err = command.Wait()
	exitError, exitErrorOk := err.(*exec.ExitError)
	if exitErrorOk {
		status, statusOk := exitError.Sys().(syscall.WaitStatus)
		if statusOk {
			// Non zero exit status code
			return buffer.String(), status.ExitStatus(), err
		}
	}

	if err != nil {
		log.Error(err)
		return buffer.String(), 0, err
	}

	return buffer.String(), 0, nil
}
