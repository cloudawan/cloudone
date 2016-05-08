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
		Do(returns200, returns400, returns409, returns422, returns500).
		Reads(GlusterfsClusterInput{}))

	ws.Route(ws.DELETE("/clusters/{cluster}").Filter(authorize).Filter(auditLog).To(deleteGlusterfsCluster).
		Doc("Delete the gluster cluster configuration").
		Param(ws.PathParameter("cluster", "Cluster name").DataType("string")).
		Do(returns200, returns422, returns500))

	ws.Route(ws.PUT("/clusters/{cluster}").Filter(authorize).Filter(auditLogWithoutBody).To(putGlusterfsCluster).
		Doc("Modify gluster cluster configuration").
		Param(ws.PathParameter("cluster", "Cluster name").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(GlusterfsClusterInput{}))

	ws.Route(ws.GET("/clusters/{cluster}").Filter(authorize).Filter(auditLog).To(getGlusterfsCluster).
		Doc("Get the glusterfs configuration").
		Param(ws.PathParameter("cluster", "Cluster name").DataType("string")).
		Do(returns200GlusterfsCluster, returns404, returns500))

	ws.Route(ws.GET("/clusters/{cluster}/volumes/").Filter(authorize).Filter(auditLog).To(getAllGlusterfsVolume).
		Doc("Get all of the glusterfs volume").
		Param(ws.PathParameter("cluster", "Cluster name").DataType("string")).
		Do(returns200AllGlusterfsVolume, returns404, returns422, returns500))

	ws.Route(ws.POST("/clusters/{cluster}/volumes/").Filter(authorize).Filter(auditLog).To(postGlusterfsVolume).
		Doc("Create and start the glusterfs volume").
		Param(ws.PathParameter("cluster", "Cluster name").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(GlusterfsVolumeInput{}))

	ws.Route(ws.DELETE("/clusters/{cluster}/volumes/{volume}").Filter(authorize).Filter(auditLog).To(deleteGlusterfsVolume).
		Doc("Stop and delete the gluster volume deployment").
		Param(ws.PathParameter("cluster", "Cluster name").DataType("string")).
		Param(ws.PathParameter("volume", "Volume name").DataType("string")).
		Do(returns200, returns404, returns422, returns500))
}

func getAllGlusterfsCluster(request *restful.Request, response *restful.Response) {
	glusterfsClusterSlice, err := glusterfs.GetStorage().LoadAllGlusterfsCluster()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all glusterfs cluster failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(glusterfsClusterSlice, "[]GlusterfsCluster")
}

func postGlusterfsCluster(request *restful.Request, response *restful.Response) {
	glusterfsClusterInput := GlusterfsClusterInput{}
	err := request.ReadEntity(&glusterfsClusterInput)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	glusterfsCluster, _ := glusterfs.GetStorage().LoadGlusterfsCluster(glusterfsClusterInput.Name)
	if glusterfsCluster != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The glusterfs cluster to create already exists"
		jsonMap["name"] = glusterfsClusterInput.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(409, string(errorMessageByteSlice))
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
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save glusterfs cluster failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["glusterfsCluster"] = glusterfsCluster
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deleteGlusterfsCluster(request *restful.Request, response *restful.Response) {
	cluster := request.PathParameter("cluster")

	err := glusterfs.GetStorage().DeleteGlusterfsCluster(cluster)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete glusterfs cluster failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["cluster"] = cluster
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func putGlusterfsCluster(request *restful.Request, response *restful.Response) {
	cluster := request.PathParameter("cluster")

	glusterfsClusterInput := GlusterfsClusterInput{}
	err := request.ReadEntity(&glusterfsClusterInput)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["cluster"] = cluster
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	if cluster != glusterfsClusterInput.Name {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Path parameter name is different from name in the body"
		jsonMap["path"] = cluster
		jsonMap["body"] = glusterfsClusterInput.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	glusterfsCluster, _ := glusterfs.GetStorage().LoadGlusterfsCluster(cluster)
	if glusterfsCluster == nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The glusterfs cluster to update doesn't exist"
		jsonMap["name"] = cluster
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
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
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save glusterfs cluster failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["glusterfsCluster"] = glusterfsCluster
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func getGlusterfsCluster(request *restful.Request, response *restful.Response) {
	cluster := request.PathParameter("cluster")

	glusterfsCluster, err := glusterfs.GetStorage().LoadGlusterfsCluster(cluster)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get glusterfs cluster failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["cluster"] = cluster
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(glusterfsCluster, "GlusterfsCluster")
}

func getAllGlusterfsVolume(request *restful.Request, response *restful.Response) {
	cluster := request.PathParameter("cluster")

	glusterfsCluster, err := glusterfs.GetStorage().LoadGlusterfsCluster(cluster)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get glusterfs cluster failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["cluster"] = cluster
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	glusterfsVolumeSlice, err := glusterfsCluster.GetAllVolume()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all glusterfs volume failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["cluster"] = cluster
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(glusterfsVolumeSlice, "[]GlusterfsVolume")
}

func postGlusterfsVolume(request *restful.Request, response *restful.Response) {
	cluster := request.PathParameter("cluster")

	glusterfsVolumeInput := GlusterfsVolumeInput{}
	err := request.ReadEntity(&glusterfsVolumeInput)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["cluster"] = cluster
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	glusterfsCluster, err := glusterfs.GetStorage().LoadGlusterfsCluster(cluster)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get glusterfs cluster failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["cluster"] = cluster
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	err = glusterfsCluster.CreateVolume(glusterfsVolumeInput.Name,
		glusterfsVolumeInput.Stripe, glusterfsVolumeInput.Replica,
		glusterfsVolumeInput.Transport, glusterfsVolumeInput.HostSlice)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Create glusterfs volume failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["cluster"] = cluster
		jsonMap["glusterfsVolumeInput"] = glusterfsVolumeInput
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	err = glusterfsCluster.StartVolume(glusterfsVolumeInput.Name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Start glusterfs volume failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["cluster"] = cluster
		jsonMap["glusterfsVolumeInput"] = glusterfsVolumeInput
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deleteGlusterfsVolume(request *restful.Request, response *restful.Response) {
	cluster := request.PathParameter("cluster")
	volume := request.PathParameter("volume")

	glusterfsCluster, err := glusterfs.GetStorage().LoadGlusterfsCluster(cluster)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get glusterfs cluster failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["cluster"] = cluster
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	err = glusterfsCluster.StopVolume(volume)

	/*
		if err != nil {
			jsonMap := make(map[string]interface{})
			jsonMap["Error"] = "Stop glusterfs volume failure"
			jsonMap["ErrorMessage"] = err.Error()
			jsonMap["cluster"] = cluster
			jsonMap["volume"] = volume
			errorMessageByteSlice, _ := json.Marshal(jsonMap)
			log.Error(jsonMap)
			response.WriteErrorString(422, string(errorMessageByteSlice))
			return
		}
	*/

	err = glusterfsCluster.DeleteVolume(volume)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete glusterfs volume failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["cluster"] = cluster
		jsonMap["volume"] = volume
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
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
