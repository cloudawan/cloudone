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
	"github.com/cloudawan/kubernetes_management/control/glusterfs"
	"github.com/emicklei/go-restful"
	"net/http"
)

type GlusterfsVolumeInput struct {
	Name      string
	Stripe    int
	Replica   int
	Transport string
	IpList    []string
}

func registerWebServiceGlusterfsVolume() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/glusterfsvolumes")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").To(getAllGlusterfsVolume).
		Doc("Get all of the glusterfs volume").
		Do(returns200AllGlusterfsVolume, returns404, returns500))

	ws.Route(ws.POST("/").To(postGlusterfsVolume).
		Doc("Create and start the glusterfs volume").
		Do(returns200, returns400, returns404, returns500).
		Reads(GlusterfsVolumeInput{}))

	ws.Route(ws.DELETE("/{volume}").To(deleteGlusterfsVolume).
		Doc("Stop and delete the gluster volume deployment").
		Param(ws.PathParameter("volume", "Volume name").DataType("string")).
		Do(returns200, returns404, returns500))

	ws.Route(ws.GET("/configuration").To(getGlusterfsServerConfiguration).
		Doc("Get glusterfs server configuration").
		Do(returns200GlusterfsServerConfiguration, returns404, returns500))
}

func getAllGlusterfsVolume(request *restful.Request, response *restful.Response) {
	glusterfsVolumeControl, err := glusterfs.CreateGlusterfsVolumeControl()
	if err != nil {
		errorText := fmt.Sprintf("Could not get glusterfs volume configuration with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	glusterfsVolumeSlice, err := glusterfsVolumeControl.GetAllVolume()

	if err != nil {
		errorText := fmt.Sprintf("Get all gluster volume failure with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(glusterfsVolumeSlice, "[]GlusterfsVolume")
}

func postGlusterfsVolume(request *restful.Request, response *restful.Response) {
	glusterfsVolumeInput := GlusterfsVolumeInput{}
	err := request.ReadEntity(&glusterfsVolumeInput)

	if err != nil {
		errorText := fmt.Sprintf("POST parse glusterfs volume input failure with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	glusterfsVolumeControl, err := glusterfs.CreateGlusterfsVolumeControl()
	if err != nil {
		errorText := fmt.Sprintf("Could not get glusterfs volume configuration with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	err = glusterfsVolumeControl.CreateVolume(glusterfsVolumeInput.Name,
		glusterfsVolumeInput.Stripe, glusterfsVolumeInput.Replica,
		glusterfsVolumeInput.Transport, glusterfsVolumeInput.IpList)

	if err != nil {
		errorText := fmt.Sprintf("Create glusterfs volume %s error %s", glusterfsVolumeInput, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	err = glusterfsVolumeControl.StartVolume(glusterfsVolumeInput.Name)

	if err != nil {
		errorText := fmt.Sprintf("Start glusterfs volume %s error %s", glusterfsVolumeInput.Name, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func deleteGlusterfsVolume(request *restful.Request, response *restful.Response) {
	volume := request.PathParameter("volume")

	glusterfsVolumeControl, err := glusterfs.CreateGlusterfsVolumeControl()
	if err != nil {
		errorText := fmt.Sprintf("Could not get glusterfs volume configuration with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	err = glusterfsVolumeControl.StopVolume(volume)

	if err != nil {
		errorText := fmt.Sprintf("Stop glusterfs volume %s error %s", volume, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	err = glusterfsVolumeControl.DeleteVolume(volume)

	if err != nil {
		errorText := fmt.Sprintf("Delete glusterfs volume %s error %s", volume, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func getGlusterfsServerConfiguration(request *restful.Request, response *restful.Response) {
	glusterfsVolumeControl, err := glusterfs.CreateGlusterfsVolumeControl()
	if err != nil {
		errorText := fmt.Sprintf("Could not get glusterfs volume configuration with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(glusterfsVolumeControl, "[]GlusterfsVolumeControl")
}

func returns200AllGlusterfsVolume(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []glusterfs.GlusterfsVolume{})
}

func returns200GlusterfsServerConfiguration(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", glusterfs.GlusterfsVolumeControl{})
}
