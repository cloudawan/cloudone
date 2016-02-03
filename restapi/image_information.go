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

package restapi

import (
	"fmt"
	"github.com/cloudawan/cloudone/image"
	"github.com/emicklei/go-restful"
	"net/http"
)

type ImageInformationCreateInput struct {
	Name           string
	Kind           string
	Description    string
	CurrentVersion string
	BuildParameter interface{}
}

type ImageInformationUpgradeInput struct {
	ImageInformationName string
	Description          string
}

func registerWebServiceImageInformation() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/imageinformations")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").To(getAllImageInformation).
		Doc("Get all of the image information").
		Do(returns200AllImageInformation, returns404, returns500))

	ws.Route(ws.DELETE("/{imageinformationname}").To(deleteImageInformationAndRelatedRecords).
		Doc("Delete image information and related records").
		Param(ws.PathParameter("imageinformationname", "Image information name").DataType("string")).
		Do(returns200, returns404, returns500))

	ws.Route(ws.POST("/create").To(postImageInformationCreate).
		Doc("Create image build from source code").
		Do(returns200, returns400, returns404, returns500).
		Reads(ImageInformationCreateInput{}))

	ws.Route(ws.PUT("/upgrade").To(putImageInformationUpgrade).
		Doc("Upgrade image build from source code").
		Do(returns200, returns400, returns404, returns500).
		Reads(ImageInformationUpgradeInput{}))
}

func getAllImageInformation(request *restful.Request, response *restful.Response) {
	imageInformationSlice, err := image.GetStorage().LoadAllImageInformation()
	if err != nil {
		errorText := fmt.Sprintf("Get all image information failure %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(imageInformationSlice, "[]InformationSlice")
}

func deleteImageInformationAndRelatedRecords(request *restful.Request, response *restful.Response) {
	imageInformationName := request.PathParameter("imageinformationname")
	err := image.GetStorage().DeleteImageInformationAndRelatedRecord(imageInformationName)
	if err != nil {
		errorText := fmt.Sprintf("Delete image information %s and related records failure %s", imageInformationName, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func postImageInformationCreate(request *restful.Request, response *restful.Response) {
	imageInformation := new(image.ImageInformation)
	err := request.ReadEntity(&imageInformation)

	if err != nil {
		errorText := fmt.Sprintf("POST failure with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = image.BuildCreate(imageInformation)
	if err != nil {
		errorText := fmt.Sprintf("Build create failure imageInformation %s error %s", imageInformation, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func putImageInformationUpgrade(request *restful.Request, response *restful.Response) {
	imageInformationUpgradeInput := new(ImageInformationUpgradeInput)
	err := request.ReadEntity(&imageInformationUpgradeInput)

	if err != nil {
		errorText := fmt.Sprintf("PUT failure with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = image.BuildUpgrade(
		imageInformationUpgradeInput.ImageInformationName,
		imageInformationUpgradeInput.Description)
	if err != nil {
		errorText := fmt.Sprintf("Build upgrade failure imageInformationName %s error %s",
			imageInformationUpgradeInput.ImageInformationName, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func returns200AllImageInformation(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []image.ImageInformation{})
}
