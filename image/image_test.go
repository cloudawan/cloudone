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

/*
import (
	"fmt"
	"testing"
)

func TestBuildSFTP(t *testing.T) {
	buildParameter := make(map[string]string)
	buildParameter["workingDirectory"] = "/tmp/aaa"
	buildParameter["repositoryPath"] = "private-repository:31000/tp"
	buildParameter["hostAndPort"] = "172.16.0.113:22"
	buildParameter["username"] = "cloudawan"
	buildParameter["password"] = "cloud4win"
	buildParameter["sourcePath"] = "/home/cloudawan/test_sftp"
	buildParameter["sourceCodeProject"] = "tp"
	buildParameter["sourceCodeDirectory"] = "src"
	buildParameter["sourceCodeMakeScript"] = ""
	buildParameter["versionFile"] = "version"
	imageInformation := &ImageInformation{
		"tp",
		"sftp",
		"description",
		"",
		buildParameter,
	}
	_, output, err := Build(imageInformation, "d2")
	fmt.Println(output)
	if err != nil {
		t.Error(err)
	}
}


func TestLoadImageInformation(t *testing.T) {
	fmt.Println(LoadImageInformation("a"))
}

func TestBuildSCP(t *testing.T) {
	buildParameter := make(map[string]string)
	buildParameter["workingDirectory"] = "/var/lib/cloudone"
	buildParameter["repositoryPath"] = "private-repository:31000/tp"
	buildParameter["hostAndPort"] = "172.16.0.113:22"
	buildParameter["username"] = "cloudawan"
	buildParameter["password"] = "cloud4win"
	buildParameter["sourcePath"] = "/home/cloudawan/test_scp"
	buildParameter["compressFileName"] = "tp.tar.gz"
	buildParameter["unpackageCommand"] = "tar zxvf"
	buildParameter["sourceCodeProject"] = "tp"
	buildParameter["sourceCodeDirectory"] = "src"
	buildParameter["sourceCodeMakeScript"] = ""
	buildParameter["versionFile"] = "version"
	imageInformation := &ImageInformation{
		"tp",
		"scp",
		"description",
		"",
		buildParameter,
	}
	_, output, err := Build(imageInformation, "d2")
	fmt.Println(output)
	if err != nil {
		t.Error(err)
	}
}

func TestBuild(t *testing.T) {
	imageInformation := &ImageInformation{
		"test",
		"git",
		"/var/lib/cloudone",
		"private-repository:31000/test",
		"https://github.com/cloudawan/test.git",
		"test",
		"src",
		"",
		"",
		"",
		"description",
		"",
	}
	_, output, err := Build(imageInformation, "v2")
	fmt.Println(output)
	if err != nil {
		t.Error(err)
	}
}
*/
