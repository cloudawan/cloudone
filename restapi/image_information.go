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
	"encoding/json"
	"github.com/cloudawan/cloudone/deploy"
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

	ws.Route(ws.GET("/").Filter(authorize).Filter(auditLog).To(getAllImageInformation).
		Doc("Get all of the image information").
		Do(returns200AllImageInformation, returns422, returns500))

	ws.Route(ws.DELETE("/{imageinformationname}").Filter(authorize).Filter(auditLog).To(deleteImageInformationAndRelatedRecords).
		Doc("Delete image information and related records").
		Param(ws.PathParameter("imageinformationname", "Image information name").DataType("string")).
		Do(returns200, returns400, returns422, returns500))

	ws.Route(ws.POST("/create").Filter(authorize).Filter(auditLog).To(postImageInformationCreate).
		Doc("Create image build from source code").
		Do(returns200, returns400, returns422, returns500).
		Reads(ImageInformationCreateInput{}))

	ws.Route(ws.PUT("/upgrade").Filter(authorize).Filter(auditLog).To(putImageInformationUpgrade).
		Doc("Upgrade image build from source code").
		Do(returns200, returns400, returns422, returns500).
		Reads(ImageInformationUpgradeInput{}))
}

func getAllImageInformation(request *restful.Request, response *restful.Response) {
	imageInformationSlice, err := image.GetStorage().LoadAllImageInformation()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all image information failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(imageInformationSlice, "[]InformationSlice")
}

func deleteImageInformationAndRelatedRecords(request *restful.Request, response *restful.Response) {
	imageInformationName := request.PathParameter("imageinformationname")

	used, err := deploy.IsImageInformationUsed(imageInformationName)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Check whether image information is used failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["imageInformationName"] = imageInformationName
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
	if used {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Can't delete the used image information"
		jsonMap["imageInformationName"] = imageInformationName
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = image.DeleteImageInformationAndRelatedRecord(imageInformationName)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete image information and related records failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["imageInformationName"] = imageInformationName
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func postImageInformationCreate(request *restful.Request, response *restful.Response) {
	imageInformation := new(image.ImageInformation)
	err := request.ReadEntity(&imageInformation)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	outputMessage, err := image.BuildCreate(imageInformation)

	resultJsonMap := make(map[string]interface{})
	resultJsonMap["OutputMessage"] = outputMessage
	statusCode := 200
	if err != nil {
		statusCode = 422
		resultJsonMap["Error"] = "Create build failure"
		resultJsonMap["ErrorMessage"] = err.Error()
		resultJsonMap["imageInformation"] = imageInformation
		log.Error(resultJsonMap)
	}
	result, err := json.Marshal(resultJsonMap)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Marshal output message failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["resultJsonMap"] = resultJsonMap
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteErrorString(statusCode, string(result))

	return
}

func putImageInformationUpgrade(request *restful.Request, response *restful.Response) {
	imageInformationUpgradeInput := new(ImageInformationUpgradeInput)
	err := request.ReadEntity(&imageInformationUpgradeInput)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	outputMessage, err := image.BuildUpgrade(
		imageInformationUpgradeInput.ImageInformationName,
		imageInformationUpgradeInput.Description)

	resultJsonMap := make(map[string]interface{})
	resultJsonMap["OutputMessage"] = outputMessage
	statusCode := 200
	if err != nil {
		statusCode = 422
		resultJsonMap["Error"] = "Upgrade build failure"
		resultJsonMap["ErrorMessage"] = err.Error()
		resultJsonMap["imageInformationUpgradeInput"] = imageInformationUpgradeInput
		log.Error(resultJsonMap)
	}
	result, err := json.Marshal(resultJsonMap)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Marshal output message failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["resultJsonMap"] = resultJsonMap
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteErrorString(statusCode, string(result))
	return
}

func returns200AllImageInformation(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []image.ImageInformation{})
}
