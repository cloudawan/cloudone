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
		Do(returns200ImageRecordSlice, returns404, returns500))

	ws.Route(ws.DELETE("/{imageinformationname}/{imagerecordversion}").Filter(authorize).Filter(auditLog).To(deleteImageRecordBelongToImageInformation).
		Doc("Delete image record belong to the image information").
		Param(ws.PathParameter("imageinformationname", "Image information name").DataType("string")).
		Param(ws.PathParameter("imagerecordversion", "Image record version").DataType("string")).
		Do(returns200, returns400, returns404, returns500))
}

func getImageRecordBelongToImageInformation(request *restful.Request, response *restful.Response) {
	imageInformationName := request.PathParameter("imageinformationname")
	imageRecordSlice, err := image.GetStorage().LoadImageRecordWithImageInformationName(imageInformationName)
	if err != nil {
		errorText := fmt.Sprintf("Get image record belong to the image information %s failure %s", imageInformationName, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(imageRecordSlice, "[]ImageRecord")
}

func deleteImageRecordBelongToImageInformation(request *restful.Request, response *restful.Response) {
	imageInformationName := request.PathParameter("imageinformationname")
	imageRecordVersion := request.PathParameter("imagerecordversion")

	used, err := deploy.IsImageRecordUsed(imageInformationName, imageRecordVersion)
	if err != nil {
		errorText := fmt.Sprintf("Check whether image record is used error version %s belong to the image information %s failure %s", imageRecordVersion, imageInformationName, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
	if used {
		errorText := fmt.Sprintf("Image record is used version %s belong to the image information %s", imageRecordVersion, imageInformationName)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = image.DeleteImageRecord(imageInformationName, imageRecordVersion)
	if err != nil {
		errorText := fmt.Sprintf("Delete image record version %s belong to the image information %s failure %s", imageRecordVersion, imageInformationName, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func returns200ImageRecordSlice(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []image.ImageRecord{})
}
