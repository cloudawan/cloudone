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

	// Update image information with version
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
		return nil, "", errors.New("Image is under building")
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
	outputByteSlice := make([]byte, 0)

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

	// Check working space
	if _, err := os.Stat(workingDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(workingDirectory, os.ModePerm)
		if err != nil {
			log.Error("Create non-existing directory %s error: %s", workingDirectory, err)
			return imageRecord, string(outputByteSlice), err
		}
	}

	// First time
	command := exec.Command("git", "clone", sourceCodeURL)
	command.Dir = workingDirectory
	out, err := command.CombinedOutput()
	outputMessage := string(out)
	outputByteSlice = append(outputByteSlice, out...)
	if err != nil {
		if err.Error() == "exit status 128" && strings.HasPrefix(outputMessage, "fatal: destination path") && strings.Contains(outputMessage, "already exists and is not an empty directory.") {
			// Already cloned
		} else {
			log.Error("Git clone %s error: %s", imageInformation, err)
			return imageRecord, string(outputByteSlice), err
		}
	}

	command = exec.Command("git", "pull")
	command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
	out, err = command.CombinedOutput()
	outputByteSlice = append(outputByteSlice, out...)
	if err != nil {
		log.Error("Git pull %s error: %s", imageInformation, err)
		return imageRecord, string(outputByteSlice), err
	}

	// Get git version
	command = exec.Command("git", "log", "-1")
	command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
	out, err = command.CombinedOutput()
	outputByteSlice = append(outputByteSlice, out...)
	if err != nil {
		log.Error("Git log %s -l error: %s", imageInformation, err)
		return imageRecord, string(outputByteSlice), err
	}
	gitVersionSlice, err := parseGitVersion(string(out))
	currentGitVersion := gitVersionSlice[0]
	imageRecord.VersionInfo["Commit"] = currentGitVersion.Commit
	imageRecord.VersionInfo["Autor"] = currentGitVersion.Autor
	imageRecord.VersionInfo["Date"] = currentGitVersion.Date

	if sourceCodeMakeScript != "" {
		commandSlice := strings.Split(sourceCodeMakeScript, " ")
		if len(commandSlice) == 1 {
			command = exec.Command(sourceCodeMakeScript)
		} else {
			command = exec.Command(commandSlice[0], commandSlice[1:]...)
		}
		command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
		out, err = command.CombinedOutput()
		outputByteSlice = append(outputByteSlice, out...)
		if err != nil {
			log.Error("Run make script %s error: %s", imageInformation, err)
			return imageRecord, string(outputByteSlice), err
		}
	}

	// User defined version
	version := ""
	if versionFile != "" {
		// open input file
		inputFile, err := os.Open(workingDirectory + string(os.PathSeparator) +
			sourceCodeProject + string(os.PathSeparator) + versionFile)
		if err != nil {
			log.Error("Open version file error: %s", err)
			return imageRecord, string(outputByteSlice), err
		}
		defer inputFile.Close()

		byteSlice := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			// read a chunk
			n, err := inputFile.Read(buffer)
			if err != nil && err != io.EOF {
				log.Error("Read version file error: %s", err)
				return imageRecord, string(outputByteSlice), err
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
			log.Error("Open environment file error: %s", err)
			return imageRecord, string(outputByteSlice), err
		}
		defer inputFile.Close()

		byteSlice := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			// read a chunk
			n, err := inputFile.Read(buffer)
			if err != nil && err != io.EOF {
				log.Error("Read version file error: %s", err)
				return imageRecord, string(outputByteSlice), err
			}
			if n == 0 {
				break
			}

			byteSlice = append(byteSlice, buffer[0:n]...)
		}

		jsonMap := make(map[string]interface{})
		err = json.Unmarshal(byteSlice, &jsonMap)
		if err != nil {
			log.Error("Unmarshal environment file error: %s", err)
			return imageRecord, string(outputByteSlice), err
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
	out, err = command.CombinedOutput()
	outputByteSlice = append(outputByteSlice, out...)
	if err != nil {
		log.Error("Docker build %s error: %s", imageInformation, err)
		return imageRecord, string(outputByteSlice), err
	}

	command = exec.Command("docker", "push", imageRecord.Path)
	command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
	out, err = command.CombinedOutput()
	outputByteSlice = append(outputByteSlice, out...)
	if err != nil {
		log.Error("Docker push %s error: %s", imageInformation, err)
		return imageRecord, string(outputByteSlice), err
	}

	// Remove working space
	os.RemoveAll(workingDirectory)

	// Success
	imageRecord.Failure = false

	return imageRecord, string(outputByteSlice), nil
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
	outputByteSlice := make([]byte, 0)

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

	// Check working space
	if _, err := os.Stat(workingDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(workingDirectory, os.ModePerm)
		if err != nil {
			log.Error("Create non-existing directory %s error: %s", workingDirectory, err)
			return imageRecord, string(outputByteSlice), err
		}
	}

	client, err := simplessh.ConnectWithPasswordTimeout(hostAndPort, username, password, scpTimeout)
	if err != nil {
		log.Error("Login scp hostAndPort %s, username %s, password %s error: %s", hostAndPort, username, password, err)
		return imageRecord, string(outputByteSlice), err
	}
	defer client.Close()

	remoteFilePath := sourcePath + string(os.PathSeparator) + compressFileName
	localFilePath := workingDirectory + string(os.PathSeparator) + compressFileName
	if err := client.Download(remoteFilePath, localFilePath); err != nil {
		log.Error("Download remoteFilePath %s localFilePath %s with scp error: %s", remoteFilePath, localFilePath, err)
		return imageRecord, string(outputByteSlice), err
	}

	unpackageCommandSlice := strings.Split(unpackageCommand, " ")
	unpackageCommandSlice = append(unpackageCommandSlice, compressFileName)

	command := exec.Command(unpackageCommandSlice[0], unpackageCommandSlice[1:]...)
	command.Dir = workingDirectory
	out, err := command.CombinedOutput()
	outputByteSlice = append(outputByteSlice, out...)
	if err != nil {
		log.Error("Unpackage compress file %s error: %s", unpackageCommand, err)
		return imageRecord, string(outputByteSlice), err
	}

	if sourceCodeMakeScript != "" {
		command = exec.Command(sourceCodeMakeScript)
		command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
		out, err = command.CombinedOutput()
		outputByteSlice = append(outputByteSlice, out...)
		if err != nil {
			log.Error("Run make script %s error: %s", imageInformation, err)
			return imageRecord, string(outputByteSlice), err
		}
	}

	// User defined version
	version := ""
	if versionFile != "" {
		// open input file
		inputFile, err := os.Open(workingDirectory + string(os.PathSeparator) +
			sourceCodeProject + string(os.PathSeparator) + versionFile)
		if err != nil {
			log.Error("Open version file error: %s", err)
			return imageRecord, string(outputByteSlice), err
		}
		defer inputFile.Close()

		byteSlice := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			// read a chunk
			n, err := inputFile.Read(buffer)
			if err != nil && err != io.EOF {
				log.Error("Read version file error: %s", err)
				return imageRecord, string(outputByteSlice), err
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
			log.Error("Open environment file error: %s", err)
			return imageRecord, string(outputByteSlice), err
		}
		defer inputFile.Close()

		byteSlice := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			// read a chunk
			n, err := inputFile.Read(buffer)
			if err != nil && err != io.EOF {
				log.Error("Read environment file error: %s", err)
				return imageRecord, string(outputByteSlice), err
			}
			if n == 0 {
				break
			}

			byteSlice = append(byteSlice, buffer[0:n]...)
		}

		jsonMap := make(map[string]interface{})
		err = json.Unmarshal(byteSlice, &jsonMap)
		if err != nil {
			log.Error("Unmarshal environment file error: %s", err)
			return imageRecord, string(outputByteSlice), err
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
	out, err = command.CombinedOutput()
	outputByteSlice = append(outputByteSlice, out...)
	if err != nil {
		log.Error("Docker build %s error: %s", imageInformation, err)
		return imageRecord, string(outputByteSlice), err
	}

	command = exec.Command("docker", "push", imageRecord.Path)
	command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
	out, err = command.CombinedOutput()
	outputByteSlice = append(outputByteSlice, out...)
	if err != nil {
		log.Error("Docker push %s error: %s", imageInformation, err)
		return imageRecord, string(outputByteSlice), err
	}

	// Remove working space
	os.RemoveAll(workingDirectory)

	// Success
	imageRecord.Failure = false

	return imageRecord, string(outputByteSlice), nil
}

func BuildFromSFTP(imageInformation *ImageInformation, description string) (*ImageRecord, string, error) {
	outputByteSlice := make([]byte, 0)

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

	// Check working space
	if _, err := os.Stat(workingDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(workingDirectory, os.ModePerm)
		if err != nil {
			log.Error("Create non-existing directory %s error: %s", workingDirectory, err)
			return imageRecord, string(outputByteSlice), err
		}
	}

	if err := sftp.DownLoadDirectoryRecurrsively(hostAndPort, username,
		password, sourcePath, workingDirectory); err != nil {
		log.Error("Download from sftp hostAndPort %s, username %s, password %s, sourcePath %s, workingDirectory %s error: %s",
			hostAndPort, username, password, sourcePath, workingDirectory, err)
		return imageRecord, "", err
	}

	if sourceCodeMakeScript != "" {
		command := exec.Command(sourceCodeMakeScript)
		command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
		out, err := command.CombinedOutput()
		outputByteSlice = append(outputByteSlice, out...)
		if err != nil {
			log.Error("Run make script %s error: %s", imageInformation, err)
			return imageRecord, string(outputByteSlice), err
		}
	}

	// User defined version
	version := ""
	if versionFile != "" {
		// open input file
		inputFile, err := os.Open(workingDirectory + string(os.PathSeparator) +
			sourceCodeProject + string(os.PathSeparator) + versionFile)
		if err != nil {
			log.Error("Open version file error: %s", err)
			return imageRecord, string(outputByteSlice), err
		}
		defer inputFile.Close()

		byteSlice := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			// read a chunk
			n, err := inputFile.Read(buffer)
			if err != nil && err != io.EOF {
				log.Error("Read version file error: %s", err)
				return imageRecord, string(outputByteSlice), err
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
			log.Error("Open environment file error: %s", err)
			return imageRecord, string(outputByteSlice), err
		}
		defer inputFile.Close()

		byteSlice := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			// read a chunk
			n, err := inputFile.Read(buffer)
			if err != nil && err != io.EOF {
				log.Error("Read environment file error: %s", err)
				return imageRecord, string(outputByteSlice), err
			}
			if n == 0 {
				break
			}

			byteSlice = append(byteSlice, buffer[0:n]...)
		}

		jsonMap := make(map[string]interface{})
		err = json.Unmarshal(byteSlice, &jsonMap)
		if err != nil {
			log.Error("Unmarshal environment file error: %s", err)
			return imageRecord, string(outputByteSlice), err
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
	out, err := command.CombinedOutput()
	outputByteSlice = append(outputByteSlice, out...)
	if err != nil {
		log.Error("Docker build %s error: %s", imageInformation, err)
		return imageRecord, string(outputByteSlice), err
	}

	command = exec.Command("docker", "push", imageRecord.Path)
	command.Dir = workingDirectory + string(os.PathSeparator) + sourceCodeProject
	out, err = command.CombinedOutput()
	outputByteSlice = append(outputByteSlice, out...)
	if err != nil {
		log.Error("Docker push %s error: %s", imageInformation, err)
		return imageRecord, string(outputByteSlice), err
	}

	// Remove working space
	os.RemoveAll(workingDirectory)

	// Success
	imageRecord.Failure = false

	return imageRecord, string(outputByteSlice), nil
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
