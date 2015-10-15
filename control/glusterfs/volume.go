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
	"github.com/cloudawan/kubernetes_management/utility/configuration"
	"github.com/cloudawan/kubernetes_management_utility/logger"
	"os/exec"
	"strconv"
	"strings"
)

func CreateGlusterfsVolumeControl() (*GlusterfsVolumeControl, error) {
	var ok bool

	glusterfsClusterIPSlice, ok := configuration.LocalConfiguration.GetStringSlice("glusterfsClusterIP")
	if ok == false {
		log.Error("Can't load glusterfsClusterIPSlice")
		return nil, errors.New("Can't load glusterfsClusterIPSlice")
	}

	glusterfsPath, ok := configuration.LocalConfiguration.GetString("glusterfsPath")
	if ok == false {
		log.Error("Can't load glusterfsPath")
		return nil, errors.New("Can't load glusterfsPath")
	}

	glusterfsVolumeControl := &GlusterfsVolumeControl{glusterfsClusterIPSlice, glusterfsPath}

	return glusterfsVolumeControl, nil
}

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

type GlusterfsVolumeControl struct {
	GlusterfsClusterIPSlice []string
	GlusterfsPath           string
}

func parseVolumeInfo(byteSlice []byte) (returnedGlusterfsVolumeSlice []GlusterfsVolume, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("parseVolumeInfo Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedGlusterfsVolumeSlice = nil
			returnedError = err.(error)
		}
	}()

	glusterfsVolumeSlice := make([]GlusterfsVolume, 0)

	scanner := bufio.NewScanner(bytes.NewBuffer(byteSlice))
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
			var err error
			glusterfsVolume.Size, err = strconv.Atoi(strings.TrimSpace(strings.Split(line, "=")[1]))
			if err != nil {
				log.Error("Parse brick error %s", err)
				return nil, err
			}
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
			// Should not go to here
			return nil, errors.New("Unexpected line: " + line)
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

func healthCheck(ip string) bool {
	command := exec.Command("gluster", "--remote-host="+ip, "peer", "status")
	_, err := command.CombinedOutput()
	if err != nil {
		return false
	} else {
		return true
	}
}

func (glusterfsVolumeControl *GlusterfsVolumeControl) GetAllVolume() ([]GlusterfsVolume, error) {
	for _, ip := range glusterfsVolumeControl.GlusterfsClusterIPSlice {
		if healthCheck(ip) {
			command := exec.Command("gluster", "--remote-host="+ip, "volume", "info")
			out, err := command.CombinedOutput()
			textOut := string(out)
			log.Debug(textOut)

			if err != nil {
				log.Error("Get all volume error %s output %s", err, textOut)
				return nil, errors.New(textOut)
			} else {
				glusterfsVolumeSlice, err := parseVolumeInfo(out)
				if err != nil {
					log.Error("Parse volume info error %s", err)
					return nil, err
				} else {
					return glusterfsVolumeSlice, nil
				}
			}
		}
	}
	return nil, errors.New("No glusterfs server responses")
}

func (glusterfsVolumeControl *GlusterfsVolumeControl) CreateVolume(name string,
	stripe int, replica int, transport string, ipList []string) error {
	if stripe == 0 {
		if replica != len(ipList) {
			return errors.New("Replica amount is not the same as ip amount")
		}
	} else if replica == 0 {
		if stripe != len(ipList) {
			return errors.New("Stripe amount is not the same as ip amount")
		}
	} else {
		if stripe*replica != len(ipList) {
			return errors.New("Replica * Stripe amount is not the same as ip amount")
		}
	}

	for _, ip := range glusterfsVolumeControl.GlusterfsClusterIPSlice {
		if healthCheck(ip) {
			parameterSlice := make([]string, 0)
			parameterSlice = append(parameterSlice, "--remote-host="+ip)
			parameterSlice = append(parameterSlice, "--mode=script")
			parameterSlice = append(parameterSlice, "volume")
			parameterSlice = append(parameterSlice, "create")
			parameterSlice = append(parameterSlice, name)
			if stripe > 0 {
				parameterSlice = append(parameterSlice, "stripe")
				parameterSlice = append(parameterSlice, strconv.Itoa(stripe))
			}
			if replica > 0 {
				parameterSlice = append(parameterSlice, "replica")
				parameterSlice = append(parameterSlice, strconv.Itoa(replica))
			}
			parameterSlice = append(parameterSlice, "transport")
			parameterSlice = append(parameterSlice, transport)
			for _, ip := range ipList {
				path := ip + ":" + glusterfsVolumeControl.GlusterfsPath + "/" + name
				parameterSlice = append(parameterSlice, path)
			}
			parameterSlice = append(parameterSlice, "force")

			command := exec.Command("gluster", parameterSlice...)
			out, err := command.CombinedOutput()
			textOut := string(out)
			log.Debug(textOut)

			if err != nil {
				log.Error("Create volume error %s output %s", err, textOut)
				return errors.New(textOut)
			} else {
				return nil
			}
		}
	}
	return errors.New("No glusterfs server responses")
}

func (glusterfsVolumeControl *GlusterfsVolumeControl) StartVolume(name string) error {
	for _, ip := range glusterfsVolumeControl.GlusterfsClusterIPSlice {
		if healthCheck(ip) {
			parameterSlice := make([]string, 0)
			parameterSlice = append(parameterSlice, "--remote-host="+ip)
			parameterSlice = append(parameterSlice, "volume")
			parameterSlice = append(parameterSlice, "start")
			parameterSlice = append(parameterSlice, name)

			command := exec.Command("gluster", parameterSlice...)
			out, err := command.CombinedOutput()
			textOut := string(out)
			log.Debug(textOut)

			if err != nil {
				log.Error("Start volume error %s output %s", err, textOut)
				return errors.New(textOut)
			} else {
				return nil
			}
		}
	}
	return errors.New("No glusterfs server responses")
}

func (glusterfsVolumeControl *GlusterfsVolumeControl) StopVolume(name string) error {
	for _, ip := range glusterfsVolumeControl.GlusterfsClusterIPSlice {
		if healthCheck(ip) {
			parameterSlice := make([]string, 0)
			parameterSlice = append(parameterSlice, "--remote-host="+ip)
			parameterSlice = append(parameterSlice, "--mode=script")
			parameterSlice = append(parameterSlice, "volume")
			parameterSlice = append(parameterSlice, "stop")
			parameterSlice = append(parameterSlice, name)

			command := exec.Command("gluster", parameterSlice...)
			out, err := command.CombinedOutput()
			textOut := string(out)
			log.Debug(textOut)

			if err != nil {
				log.Error("Stop volume error %s output %s", err, textOut)
				return errors.New(textOut)
			} else {
				return nil
			}
		}
	}
	return errors.New("No glusterfs server responses")
}

func (glusterfsVolumeControl *GlusterfsVolumeControl) DeleteVolume(name string) error {
	for _, ip := range glusterfsVolumeControl.GlusterfsClusterIPSlice {
		parameterSlice := make([]string, 0)
		parameterSlice = append(parameterSlice, "--remote-host="+ip)
		parameterSlice = append(parameterSlice, "--mode=script")
		parameterSlice = append(parameterSlice, "volume")
		parameterSlice = append(parameterSlice, "delete")
		parameterSlice = append(parameterSlice, name)

		command := exec.Command("gluster", parameterSlice...)
		out, err := command.CombinedOutput()
		textOut := string(out)
		log.Debug(textOut)

		if err != nil {
			log.Error("Delete volume error %s output %s", err, textOut)
			return errors.New(textOut)
		} else {
			return nil
		}
	}
	return errors.New("No glusterfs server responses")
}
