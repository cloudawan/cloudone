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

func registerWebServiceImageRecord() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/imagerecords")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/{imageinformationname}").Filter(authorize).Filter(auditLog).To(getImageRecordBelongToImageInformation).
		Doc("Get all of the image record belong to the image information").
		Param(ws.PathParameter("imageinformationname", "Image information name").DataType("string")).
		Do(returns200ImageRecordSlice, returns422, returns500))

	ws.Route(ws.DELETE("/{imageinformationname}/{imagerecordversion}").Filter(authorize).Filter(auditLog).To(deleteImageRecordBelongToImageInformation).
		Doc("Delete image record belong to the image information").
		Param(ws.PathParameter("imageinformationname", "Image information name").DataType("string")).
		Param(ws.PathParameter("imagerecordversion", "Image record version").DataType("string")).
		Do(returns200, returns400, returns422, returns500))
}

func getImageRecordBelongToImageInformation(request *restful.Request, response *restful.Response) {
	imageInformationName := request.PathParameter("imageinformationname")

	imageRecordSlice, err := image.GetStorage().LoadImageRecordWithImageInformationName(imageInformationName)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all image reocrd belonging to the image information failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["imageInformationName"] = imageInformationName
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(imageRecordSlice, "[]ImageRecord")
}

func deleteImageRecordBelongToImageInformation(request *restful.Request, response *restful.Response) {
	imageInformationName := request.PathParameter("imageinformationname")
	imageRecordVersion := request.PathParameter("imagerecordversion")

	used, err := deploy.IsImageRecordUsed(imageInformationName, imageRecordVersion)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Check whether image reocrd is used failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["imageInformationName"] = imageInformationName
		jsonMap["imageRecordVersion"] = imageRecordVersion
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
	if used {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Can't delete used image record"
		jsonMap["imageInformationName"] = imageInformationName
		jsonMap["imageRecordVersion"] = imageRecordVersion
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = image.DeleteImageRecord(imageInformationName, imageRecordVersion)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete image reocrd belonging to the image information failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["imageInformationName"] = imageInformationName
		jsonMap["imageRecordVersion"] = imageRecordVersion
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func returns200ImageRecordSlice(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []image.ImageRecord{})
}
