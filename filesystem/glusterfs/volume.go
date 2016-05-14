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

package glusterfs

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/cloudawan/cloudone_utility/logger"
	"github.com/cloudawan/cloudone_utility/random"
	"github.com/cloudawan/cloudone_utility/sshclient"
	"strconv"
	"strings"
)

type GlusterfsVolume struct {
	VolumeName     string
	Type           string
	VolumeID       string
	Status         string
	NumberOfBricks string
	TransportType  string
	Bricks         []string
	Size           int
}

type GlusterfsVolumeCreateParameter struct {
	ClusterName  string
	VolumeName   string
	Stripe       int
	Replica      int
	Arbiter      int
	Disperse     int
	DisperseData int
	Redundancy   int
	Transport    string
	HostSlice    []string
}

func (glusterfsCluster *GlusterfsCluster) parseSize(field string) (returnedSize int, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("Parse size error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedSize = -1
			returnedError = err.(error)
		}
	}()

	var value string
	if strings.Contains(field, "=") {
		value = strings.Split(field, "=")[1]
	} else {
		value = field
	}

	size, err := strconv.Atoi(strings.TrimSpace(value))

	if err != nil {
		log.Error("Parse size error %s", err)
		return -1, err
	} else {
		return size, nil
	}
}

func (glusterfsCluster *GlusterfsCluster) parseVolumeInfo(text string) (returnedGlusterfsVolumeSlice []GlusterfsVolume, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("parseVolumeInfo Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedGlusterfsVolumeSlice = nil
			returnedError = err.(error)
		}
	}()

	glusterfsVolumeSlice := make([]GlusterfsVolume, 0)

	scanner := bufio.NewScanner(bytes.NewBufferString(text))
	var glusterfsVolume *GlusterfsVolume = nil
	for scanner.Scan() {
		line := scanner.Text()
		if line == " " {
			if glusterfsVolume != nil {
				glusterfsVolumeSlice = append(glusterfsVolumeSlice, *glusterfsVolume)
			}
			glusterfsVolume = &GlusterfsVolume{}
		} else if strings.HasPrefix(line, "Volume Name: ") {
			glusterfsVolume.VolumeName = line[len("Volume Name: "):]
		} else if strings.HasPrefix(line, "Type: ") {
			glusterfsVolume.Type = line[len("Type: "):]
		} else if strings.HasPrefix(line, "Volume ID: ") {
			glusterfsVolume.VolumeID = line[len("Volume ID: "):]
		} else if strings.HasPrefix(line, "Status: ") {
			glusterfsVolume.Status = line[len("Status: "):]
		} else if strings.HasPrefix(line, "Number of Bricks: ") {
			glusterfsVolume.NumberOfBricks = line[len("Number of Bricks: "):]
			glusterfsVolume.Size, _ = glusterfsCluster.parseSize(glusterfsVolume.NumberOfBricks)
		} else if strings.HasPrefix(line, "Transport-type: ") {
			glusterfsVolume.TransportType = line[len("Transport-type: "):]
		} else if line == "Bricks:" {
			brickSlice := make([]string, 0)
			for i := 0; i < glusterfsVolume.Size; i++ {
				scanner.Scan()
				brickSlice = append(brickSlice, scanner.Text())
			}
			glusterfsVolume.Bricks = brickSlice
		} else {
			// Ignore unexpected data
		}
	}
	if glusterfsVolume != nil {
		glusterfsVolumeSlice = append(glusterfsVolumeSlice, *glusterfsVolume)
	}
	if err := scanner.Err(); err != nil {
		log.Error("Scanner error %s", err)
		return nil, err
	}
	return glusterfsVolumeSlice, nil
}

func (glusterfsCluster *GlusterfsCluster) getAvailableHost() (*string, error) {
	commandSlice := make([]string, 0)
	commandSlice = append(commandSlice, "exit\n")

	for _, host := range glusterfsCluster.HostSlice {
		_, err := sshclient.InteractiveSSH(
			glusterfsCluster.SSHDialTimeout,
			glusterfsCluster.SSHSessionTimeout,
			host,
			glusterfsCluster.SSHPort,
			glusterfsCluster.SSHUser,
			glusterfsCluster.SSHPassword,
			commandSlice,
			nil)
		if err == nil {
			return &host, nil
		} else {
			log.Error(err)
		}
	}

	return nil, errors.New("No available host")
}

func (glusterfsCluster *GlusterfsCluster) GetHostStatus() map[string]bool {
	commandSlice := make([]string, 0)
	commandSlice = append(commandSlice, "exit\n")

	hostStatusMap := make(map[string]bool)

	for _, host := range glusterfsCluster.HostSlice {
		_, err := sshclient.InteractiveSSH(
			glusterfsCluster.SSHDialTimeout,
			glusterfsCluster.SSHSessionTimeout,
			host,
			glusterfsCluster.SSHPort,
			glusterfsCluster.SSHUser,
			glusterfsCluster.SSHPassword,
			commandSlice,
			nil)
		if err == nil {
			hostStatusMap[host] = true
		} else {
			hostStatusMap[host] = false
			log.Error(err)
		}
	}

	return hostStatusMap
}

func (glusterfsCluster *GlusterfsCluster) GetAllVolume() ([]GlusterfsVolume, error) {
	host, err := glusterfsCluster.getAvailableHost()
	if err != nil {
		return nil, err
	}

	commandSlice := make([]string, 0)
	commandSlice = append(commandSlice, "sudo gluster volume info\n")

	interactiveMap := make(map[string]string)
	interactiveMap["[sudo]"] = glusterfsCluster.SSHPassword + "\n"

	resultSlice, err := sshclient.InteractiveSSH(
		glusterfsCluster.SSHDialTimeout,
		glusterfsCluster.SSHSessionTimeout,
		*host,
		glusterfsCluster.SSHPort,
		glusterfsCluster.SSHUser,
		glusterfsCluster.SSHPassword,
		commandSlice,
		interactiveMap)

	glusterfsVolumeSlice, err := glusterfsCluster.parseVolumeInfo(resultSlice[0])
	if err != nil {
		log.Error("Parse volume info error %s", err)
		return nil, err
	} else {
		return glusterfsVolumeSlice, nil
	}
}

func (glusterfsCluster *GlusterfsCluster) GetVolume(name string) (*GlusterfsVolume, error) {
	host, err := glusterfsCluster.getAvailableHost()
	if err != nil {
		return nil, err
	}

	commandSlice := make([]string, 0)
	commandSlice = append(commandSlice, "sudo gluster volume info "+name+"\n")

	interactiveMap := make(map[string]string)
	interactiveMap["[sudo]"] = glusterfsCluster.SSHPassword + "\n"

	resultSlice, err := sshclient.InteractiveSSH(
		glusterfsCluster.SSHDialTimeout,
		glusterfsCluster.SSHSessionTimeout,
		*host,
		glusterfsCluster.SSHPort,
		glusterfsCluster.SSHUser,
		glusterfsCluster.SSHPassword,
		commandSlice,
		interactiveMap)

	glusterfsVolumeSlice, err := glusterfsCluster.parseVolumeInfo(resultSlice[0])
	if err != nil {
		log.Error("Parse volume info error %s", err)
		return nil, err
	} else {
		if len(glusterfsVolumeSlice) == 1 {
			return &glusterfsVolumeSlice[0], nil
		} else {
			log.Error("The result it not the only one. glusterfsVolumeSlice %s", glusterfsVolumeSlice)
			return nil, errors.New("The result it not the only one.")
		}

	}
}

func (glusterfsCluster *GlusterfsCluster) CreateVolume(glusterfsVolumeCreateParameter *GlusterfsVolumeCreateParameter) error {
	host, err := glusterfsCluster.getAvailableHost()
	if err != nil {
		return err
	}

	commandBuffer := bytes.Buffer{}
	commandBuffer.WriteString("sudo gluster --mode=script volume create ")
	commandBuffer.WriteString(glusterfsVolumeCreateParameter.VolumeName)

	if glusterfsVolumeCreateParameter.Stripe > 0 {
		commandBuffer.WriteString(" stripe ")
		commandBuffer.WriteString(strconv.Itoa(glusterfsVolumeCreateParameter.Stripe))
	}
	if glusterfsVolumeCreateParameter.Replica > 0 {
		commandBuffer.WriteString(" replica ")
		commandBuffer.WriteString(strconv.Itoa(glusterfsVolumeCreateParameter.Replica))
	}
	if glusterfsVolumeCreateParameter.Arbiter > 0 {
		commandBuffer.WriteString(" arbiter ")
		commandBuffer.WriteString(strconv.Itoa(glusterfsVolumeCreateParameter.Arbiter))
	}
	if glusterfsVolumeCreateParameter.Disperse > 0 {
		commandBuffer.WriteString(" disperse ")
		commandBuffer.WriteString(strconv.Itoa(glusterfsVolumeCreateParameter.Disperse))
	}
	if glusterfsVolumeCreateParameter.DisperseData > 0 {
		commandBuffer.WriteString(" disperse-data ")
		commandBuffer.WriteString(strconv.Itoa(glusterfsVolumeCreateParameter.DisperseData))
	}
	if glusterfsVolumeCreateParameter.Redundancy > 0 {
		commandBuffer.WriteString(" redundancy ")
		commandBuffer.WriteString(strconv.Itoa(glusterfsVolumeCreateParameter.Redundancy))
	}

	// <tcp|rdma|tcp,rdma>
	commandBuffer.WriteString(" transport ")
	commandBuffer.WriteString(glusterfsVolumeCreateParameter.Transport)
	uuid := random.UUID()
	for _, ip := range glusterfsVolumeCreateParameter.HostSlice {
		path := " " + ip + ":" + glusterfsCluster.Path + "/" + glusterfsVolumeCreateParameter.VolumeName + "_" + uuid
		commandBuffer.WriteString(path)
	}
	commandBuffer.WriteString(" force\n")
	commandSlice := make([]string, 0)
	commandSlice = append(commandSlice, commandBuffer.String())

	interactiveMap := make(map[string]string)
	interactiveMap["[sudo]"] = glusterfsCluster.SSHPassword + "\n"

	resultSlice, err := sshclient.InteractiveSSH(
		glusterfsCluster.SSHDialTimeout,
		glusterfsCluster.SSHSessionTimeout,
		*host,
		glusterfsCluster.SSHPort,
		glusterfsCluster.SSHUser,
		glusterfsCluster.SSHPassword,
		commandSlice,
		interactiveMap)

	if err != nil {
		log.Error("Create volume error %s resultSlice %v", err, resultSlice)
		return err
	}

	if strings.Contains(resultSlice[0], "success") == false {
		log.Debug("Issue command: " + commandBuffer.String())
		log.Error("Fail to create volume with error: " + resultSlice[0])
		return errors.New(resultSlice[0])
	}

	err = GetStorage().SaveGlusterfsVolumeCreateParameter(glusterfsVolumeCreateParameter)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (glusterfsCluster *GlusterfsCluster) StartVolume(name string) error {
	host, err := glusterfsCluster.getAvailableHost()
	if err != nil {
		return err
	}

	commandBuffer := bytes.Buffer{}
	commandBuffer.WriteString("sudo gluster --mode=script volume start ")
	commandBuffer.WriteString(name)
	commandBuffer.WriteString(" \n")
	commandSlice := make([]string, 0)
	commandSlice = append(commandSlice, commandBuffer.String())

	interactiveMap := make(map[string]string)
	interactiveMap["[sudo]"] = glusterfsCluster.SSHPassword + "\n"

	resultSlice, err := sshclient.InteractiveSSH(
		glusterfsCluster.SSHDialTimeout,
		glusterfsCluster.SSHSessionTimeout,
		*host,
		glusterfsCluster.SSHPort,
		glusterfsCluster.SSHUser,
		glusterfsCluster.SSHPassword,
		commandSlice,
		interactiveMap)

	if err != nil {
		log.Error("Create volume error %s resultSlice %s", err, resultSlice)
		return err
	} else {
		if strings.Contains(resultSlice[0], "success") {
			return nil
		} else {
			log.Debug("Issue command: " + commandBuffer.String())
			log.Error("Fail to start volume with error: " + resultSlice[0])
			return errors.New(resultSlice[0])
		}
	}
}

func (glusterfsCluster *GlusterfsCluster) StopVolume(name string) error {
	host, err := glusterfsCluster.getAvailableHost()
	if err != nil {
		return err
	}

	commandBuffer := bytes.Buffer{}
	commandBuffer.WriteString("sudo gluster --mode=script volume stop ")
	commandBuffer.WriteString(name)
	commandBuffer.WriteString(" \n")
	commandSlice := make([]string, 0)
	commandSlice = append(commandSlice, commandBuffer.String())

	interactiveMap := make(map[string]string)
	interactiveMap["[sudo]"] = glusterfsCluster.SSHPassword + "\n"

	resultSlice, err := sshclient.InteractiveSSH(
		glusterfsCluster.SSHDialTimeout,
		glusterfsCluster.SSHSessionTimeout,
		*host,
		glusterfsCluster.SSHPort,
		glusterfsCluster.SSHUser,
		glusterfsCluster.SSHPassword,
		commandSlice,
		interactiveMap)

	if err != nil {
		log.Error("Create volume error %s resultSlice %s", err, resultSlice)
		return err
	} else {
		if strings.Contains(resultSlice[0], "success") {
			return nil
		} else {
			log.Debug("Issue command: " + commandBuffer.String())
			log.Error("Fail to stop volume with error: " + resultSlice[0])
			return errors.New(resultSlice[0])
		}
	}
}

func (glusterfsCluster *GlusterfsCluster) DeleteVolume(name string) error {
	host, err := glusterfsCluster.getAvailableHost()
	if err != nil {
		return err
	}

	commandBuffer := bytes.Buffer{}
	commandBuffer.WriteString("sudo gluster --mode=script volume delete ")
	commandBuffer.WriteString(name)
	commandBuffer.WriteString(" \n")
	commandSlice := make([]string, 0)
	commandSlice = append(commandSlice, commandBuffer.String())

	interactiveMap := make(map[string]string)
	interactiveMap["[sudo]"] = glusterfsCluster.SSHPassword + "\n"

	resultSlice, err := sshclient.InteractiveSSH(
		glusterfsCluster.SSHDialTimeout,
		glusterfsCluster.SSHSessionTimeout,
		*host,
		glusterfsCluster.SSHPort,
		glusterfsCluster.SSHUser,
		glusterfsCluster.SSHPassword,
		commandSlice,
		interactiveMap)

	if err != nil {
		log.Error("Create volume error %s resultSlice %s", err, resultSlice)
		return err
	}

	if strings.Contains(resultSlice[0], "success") == false {
		log.Debug("Issue command: " + commandBuffer.String())
		log.Error("Fail to delete volume with error: " + resultSlice[0])
		return errors.New(resultSlice[0])
	}

	err = GetStorage().DeleteGlusterfsVolumeCreateParameter(glusterfsCluster.Name, name)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (glusterfsCluster *GlusterfsCluster) CleanDataOnDisk(glusterfsVolume *GlusterfsVolume) error {
	if glusterfsVolume != nil {
		for _, brick := range glusterfsVolume.Bricks {
			splitSlice := strings.Split(brick, ":")
			if len(splitSlice) != 3 {
				log.Error("brick format error %s", brick)
				return errors.New("brick format error " + brick)
			}
			brickHost := strings.TrimSpace(splitSlice[1])
			brickPath := strings.TrimSpace(splitSlice[2])

			commandBuffer := bytes.Buffer{}
			commandBuffer.WriteString("sudo rm -rf ")
			commandBuffer.WriteString(brickPath)
			commandBuffer.WriteString(" \n")
			commandSlice := make([]string, 0)
			commandSlice = append(commandSlice, commandBuffer.String())

			interactiveMap := make(map[string]string)
			interactiveMap["[sudo]"] = glusterfsCluster.SSHPassword + "\n"

			resultSlice, err := sshclient.InteractiveSSH(
				glusterfsCluster.SSHDialTimeout,
				glusterfsCluster.SSHSessionTimeout,
				brickHost,
				glusterfsCluster.SSHPort,
				glusterfsCluster.SSHUser,
				glusterfsCluster.SSHPassword,
				commandSlice,
				interactiveMap)

			if err != nil {
				log.Error("Delete data on disk error %s resultSlice %s", err, resultSlice)
				return err
			}
		}
	}

	return nil
}

func (glusterfsCluster *GlusterfsCluster) DeleteAndRecreateVolume(name string) error {
	glusterfsVolume, err := glusterfsCluster.GetVolume(name)
	if err != nil {
		log.Error(err)
		return err
	}

	glusterfsVolumeCreateParameter, err := GetStorage().LoadGlusterfsVolumeCreateParameter(glusterfsCluster.Name, name)
	if err != nil {
		log.Error(err)
		return err
	}

	err = glusterfsCluster.StopVolume(name)
	if err != nil {
		log.Error(err)
		//return err
	}

	err = glusterfsCluster.DeleteVolume(name)
	if err != nil {
		log.Error(err)
		return err
	}

	// Delete data on disk in asynchronized way since it may take hours
	go func() {
		glusterfsCluster.CleanDataOnDisk(glusterfsVolume)
	}()

	err = glusterfsCluster.CreateVolume(glusterfsVolumeCreateParameter)
	if err != nil {
		log.Error(err)
		return err
	}

	err = glusterfsCluster.StartVolume(name)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
