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
	"github.com/cloudawan/cloudone/filesystem/glusterfs"
	"github.com/emicklei/go-restful"
	"net/http"
)

type GlusterfsClusterInput struct {
	Name                           string
	HostSlice                      []string
	Path                           string
	SSHDialTimeoutInMilliSecond    int
	SSHSessionTimeoutInMilliSecond int
	SSHPort                        int
	SSHUser                        string
	SSHPassword                    string
}

type GlusterfsVolumeInput struct {
	Name      string
	Stripe    int
	Replica   int
	Transport string
	HostSlice []string
}

func registerWebServiceGlusterfs() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/glusterfs")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/clusters/").Filter(authorize).Filter(auditLog).To(getAllGlusterfsCluster).
		Doc("Get all of the gluster cluster configuration").
		Do(returns200AllGlusterfsCluster, returns404, returns500))

	ws.Route(ws.POST("/clusters/").Filter(authorize).Filter(auditLogWithoutBody).To(postGlusterfsCluster).
		Doc("Create gluster cluster configuration").
		Do(returns200, returns400, returns404, returns500).
		Reads(GlusterfsClusterInput{}))

	ws.Route(ws.DELETE("/clusters/{cluster}").Filter(authorize).Filter(auditLog).To(deleteGlusterfsCluster).
		Doc("Delete the gluster cluster configuration").
		Param(ws.PathParameter("cluster", "Cluster name").DataType("string")).
		Do(returns200, returns404, returns500))

	ws.Route(ws.PUT("/clusters/{cluster}").Filter(authorize).Filter(auditLogWithoutBody).To(putGlusterfsCluster).
		Doc("Modify gluster cluster configuration").
		Param(ws.PathParameter("cluster", "Cluster name").DataType("string")).
		Do(returns200, returns400, returns404, returns500).
		Reads(GlusterfsClusterInput{}))

	ws.Route(ws.GET("/clusters/{cluster}").Filter(authorize).Filter(auditLog).To(getGlusterfsCluster).
		Doc("Get all of the glusterfs configuration").
		Param(ws.PathParameter("cluster", "Cluster name").DataType("string")).
		Do(returns200GlusterfsCluster, returns404, returns500))

	ws.Route(ws.GET("/clusters/{cluster}/volumes/").Filter(authorize).Filter(auditLog).To(getAllGlusterfsVolume).
		Doc("Get all of the glusterfs volume").
		Param(ws.PathParameter("cluster", "Cluster name").DataType("string")).
		Do(returns200AllGlusterfsVolume, returns404, returns500))

	ws.Route(ws.POST("/clusters/{cluster}/volumes/").Filter(authorize).Filter(auditLog).To(postGlusterfsVolume).
		Doc("Create and start the glusterfs volume").
		Param(ws.PathParameter("cluster", "Cluster name").DataType("string")).
		Do(returns200, returns400, returns404, returns500).
		Reads(GlusterfsVolumeInput{}))

	ws.Route(ws.DELETE("/clusters/{cluster}/volumes/{volume}").Filter(authorize).Filter(auditLog).To(deleteGlusterfsVolume).
		Doc("Stop and delete the gluster volume deployment").
		Param(ws.PathParameter("cluster", "Cluster name").DataType("string")).
		Param(ws.PathParameter("volume", "Volume name").DataType("string")).
		Do(returns200, returns404, returns500))
}

func getAllGlusterfsCluster(request *restful.Request, response *restful.Response) {
	glusterfsClusterSlice, err := glusterfs.GetStorage().LoadAllGlusterfsCluster()
	if err != nil {
		errorText := fmt.Sprintf("Could not get all glusterfs cluster with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(glusterfsClusterSlice, "[]GlusterfsCluster")
}

func postGlusterfsCluster(request *restful.Request, response *restful.Response) {
	glusterfsClusterInput := GlusterfsClusterInput{}
	err := request.ReadEntity(&glusterfsClusterInput)

	if err != nil {
		errorText := fmt.Sprintf("POST parse glusterfs cluster input failure with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	glusterfsCluster, _ := glusterfs.GetStorage().LoadGlusterfsCluster(glusterfsClusterInput.Name)
	if glusterfsCluster != nil {
		errorText := fmt.Sprintf("The glusterfs cluster with name %s exists", glusterfsClusterInput.Name)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	glusterfsCluster = glusterfs.CreateGlusterfsCluster(
		glusterfsClusterInput.Name,
		glusterfsClusterInput.HostSlice,
		glusterfsClusterInput.Path,
		glusterfsClusterInput.SSHDialTimeoutInMilliSecond,
		glusterfsClusterInput.SSHSessionTimeoutInMilliSecond,
		glusterfsClusterInput.SSHPort,
		glusterfsClusterInput.SSHUser,
		glusterfsClusterInput.SSHPassword)

	err = glusterfs.GetStorage().SaveGlusterfsCluster(glusterfsCluster)
	if err != nil {
		errorText := fmt.Sprintf("Save glusterfs cluster %v with error %s", glusterfsCluster, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func putGlusterfsCluster(request *restful.Request, response *restful.Response) {
	cluster := request.PathParameter("cluster")

	glusterfsClusterInput := GlusterfsClusterInput{}
	err := request.ReadEntity(&glusterfsClusterInput)

	if err != nil {
		errorText := fmt.Sprintf("PUT parse glusterfs cluster input failure with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	if cluster != glusterfsClusterInput.Name {
		errorText := fmt.Sprintf("PUT name %s is different from name %s in the body", cluster, glusterfsClusterInput.Name)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	glusterfsCluster, _ := glusterfs.GetStorage().LoadGlusterfsCluster(glusterfsClusterInput.Name)
	if glusterfsCluster == nil {
		errorText := fmt.Sprintf("The glusterfs cluster with name %s doesn't exist", glusterfsClusterInput.Name)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	glusterfsCluster = glusterfs.CreateGlusterfsCluster(
		glusterfsClusterInput.Name,
		glusterfsClusterInput.HostSlice,
		glusterfsClusterInput.Path,
		glusterfsClusterInput.SSHDialTimeoutInMilliSecond,
		glusterfsClusterInput.SSHSessionTimeoutInMilliSecond,
		glusterfsClusterInput.SSHPort,
		glusterfsClusterInput.SSHUser,
		glusterfsClusterInput.SSHPassword)

	err = glusterfs.GetStorage().SaveGlusterfsCluster(glusterfsCluster)
	if err != nil {
		errorText := fmt.Sprintf("Save glusterfs cluster %v with error %s", glusterfsCluster, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func deleteGlusterfsCluster(request *restful.Request, response *restful.Response) {
	cluster := request.PathParameter("cluster")

	err := glusterfs.GetStorage().DeleteGlusterfsCluster(cluster)
	if err != nil {
		errorText := fmt.Sprintf("Delete glusterfs cluster %s with error %s", cluster, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func getGlusterfsCluster(request *restful.Request, response *restful.Response) {
	cluster := request.PathParameter("cluster")

	glusterfsCluster, err := glusterfs.GetStorage().LoadGlusterfsCluster(cluster)
	if err != nil {
		errorText := fmt.Sprintf("Could not get glusterfs cluster %s with error %s", cluster, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(glusterfsCluster, "GlusterfsCluster")
}

func getAllGlusterfsVolume(request *restful.Request, response *restful.Response) {
	cluster := request.PathParameter("cluster")

	glusterfsCluster, err := glusterfs.GetStorage().LoadGlusterfsCluster(cluster)
	if err != nil {
		errorText := fmt.Sprintf("Could not get glusterfs cluster %s with error %s", cluster, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	glusterfsVolumeSlice, err := glusterfsCluster.GetAllVolume()

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

	cluster := request.PathParameter("cluster")

	glusterfsCluster, err := glusterfs.GetStorage().LoadGlusterfsCluster(cluster)
	if err != nil {
		errorText := fmt.Sprintf("Could not get glusterfs cluster %s with error %s", cluster, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	err = glusterfsCluster.CreateVolume(glusterfsVolumeInput.Name,
		glusterfsVolumeInput.Stripe, glusterfsVolumeInput.Replica,
		glusterfsVolumeInput.Transport, glusterfsVolumeInput.HostSlice)

	if err != nil {
		errorText := fmt.Sprintf("Create glusterfs volume %s error %s", glusterfsVolumeInput, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	err = glusterfsCluster.StartVolume(glusterfsVolumeInput.Name)

	if err != nil {
		errorText := fmt.Sprintf("Start glusterfs volume %s error %s", glusterfsVolumeInput.Name, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func deleteGlusterfsVolume(request *restful.Request, response *restful.Response) {
	volume := request.PathParameter("volume")

	cluster := request.PathParameter("cluster")

	glusterfsCluster, err := glusterfs.GetStorage().LoadGlusterfsCluster(cluster)
	if err != nil {
		errorText := fmt.Sprintf("Could not get glusterfs cluster %s with error %s", cluster, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	err = glusterfsCluster.StopVolume(volume)

	/*
		if err != nil {
			errorText := fmt.Sprintf("Stop glusterfs volume %s error %s", volume, err)
			log.Error(errorText)
			response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
			return
		}
	*/

	err = glusterfsCluster.DeleteVolume(volume)

	if err != nil {
		errorText := fmt.Sprintf("Delete glusterfs volume %s error %s", volume, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func returns200AllGlusterfsCluster(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []glusterfs.GlusterfsCluster{})
}

func returns200GlusterfsCluster(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", glusterfs.GlusterfsCluster{})
}

func returns200AllGlusterfsVolume(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []glusterfs.GlusterfsVolume{})
}
