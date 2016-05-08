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
	"github.com/cloudawan/cloudone/execute"
	"github.com/cloudawan/cloudone/monitor"
	"github.com/cloudawan/cloudone/notification"
	"github.com/emicklei/go-restful"
	"net/http"
)

func registerWebServiceReplicationControllerNotifier() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/notifiers")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").Filter(authorize).Filter(auditLog).To(getAllReplicationControllerNotifier).
		Doc("Get all of the configuration of notifier").
		Do(returns200AllReplicationControllerNotifier, returns500))

	ws.Route(ws.GET("/{namespace}/{kind}/{name}").Filter(authorize).Filter(auditLog).To(getReplicationControllerNotifier).
		Doc("Get the configuration of notifier for the replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("kind", "selector or replicationController").DataType("string")).
		Param(ws.PathParameter("name", "name").DataType("string")).
		Do(returns200ReplicationControllerNotifier, returns404, returns422, returns500))

	ws.Route(ws.PUT("/").Filter(authorize).Filter(auditLog).To(putReplicationControllerNotifier).
		Doc("Add (if not existing) or update an notifier for the replication controller in the namespace").
		Do(returns200, returns400, returns422, returns500).
		Reads(notification.ReplicationControllerNotifierSerializable{}))

	ws.Route(ws.DELETE("/{namespace}/{kind}/{name}").Filter(authorize).Filter(auditLog).To(deleteReplicationControllerNotifier).
		Doc("Delete an notifier for the replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("kind", "selector or replicationController").DataType("string")).
		Param(ws.PathParameter("name", "name").DataType("string")).
		Do(returns200, returns422, returns500))

	ws.Route(ws.GET("/emailserversmtp/").Filter(authorize).Filter(auditLog).To(getAllEmailServerSMTP).
		Doc("Get all of the configuration of email server stmp").
		Do(returns200AllEmailServerSMTP, returns422, returns500))

	ws.Route(ws.POST("/emailserversmtp/").Filter(authorize).Filter(auditLogWithoutBody).To(postEmailServerSMTP).
		Doc("Create the configuration of email server stmp").
		Do(returns200, returns400, returns409, returns422, returns500).
		Reads(notification.EmailServerSMTP{}))

	ws.Route(ws.GET("/emailserversmtp/{name}").Filter(authorize).Filter(auditLog).To(getEmailServerSMTP).
		Doc("Get all of the configuration of email server stmp").
		Param(ws.PathParameter("emailserversmtpname", "email server smtp name").DataType("string")).
		Do(returns200EmailServerSMTP, returns422, returns500))

	ws.Route(ws.DELETE("/emailserversmtp/{name}").Filter(authorize).Filter(auditLog).To(deleteEmailServerSMTP).
		Doc("Delete the configuration of email server stmp").
		Param(ws.PathParameter("emailserversmtpname", "email server smtp name").DataType("string")).
		Do(returns200, returns422, returns500))

	ws.Route(ws.GET("/smsnexmo/").Filter(authorize).Filter(auditLog).To(getAllSMSNexmo).
		Doc("Get all of the configuration of sms nexmo").
		Do(returns200AllSMSNexmo, returns422, returns500))

	ws.Route(ws.POST("/smsnexmo/").Filter(authorize).Filter(auditLogWithoutBody).To(postSMSNexmo).
		Doc("Create the configuration of sms nexmo").
		Do(returns200, returns400, returns409, returns422, returns500).
		Reads(notification.SMSNexmo{}))

	ws.Route(ws.GET("/smsnexmo/{name}").Filter(authorize).Filter(auditLog).To(getSMSNexmo).
		Doc("Get all of the configuration of sms nexmo").
		Param(ws.PathParameter("smsnexmo", "sms nexmo name").DataType("string")).
		Do(returns200SMSNexmo, returns422, returns500))

	ws.Route(ws.DELETE("/smsnexmo/{name}").Filter(authorize).Filter(auditLog).To(deleteSMSNexmo).
		Doc("Delete the configuration of sms nexmo").
		Param(ws.PathParameter("smsnexmo", "sms nexmo name").DataType("string")).
		Do(returns200, returns422, returns500))
}

func getAllReplicationControllerNotifier(request *restful.Request, response *restful.Response) {
	replicationControllerNotifierMap := execute.GetReplicationControllerNotifierMap()

	replicationControllerNotifierSerializableSlice := make([]notification.ReplicationControllerNotifierSerializable, 0)
	for _, replicationControllerNotifier := range replicationControllerNotifierMap {
		replicationControllerNotifierSerializable, err := notification.ConvertToSerializable(replicationControllerNotifier)
		if err != nil {
			jsonMap := make(map[string]interface{})
			jsonMap["Error"] = "Convert replication controller notifer failure"
			jsonMap["ErrorMessage"] = err.Error()
			jsonMap["replicationControllerNotifier"] = replicationControllerNotifier
			errorMessageByteSlice, _ := json.Marshal(jsonMap)
			log.Error(jsonMap)
			response.WriteErrorString(422, string(errorMessageByteSlice))
			return
		} else {
			replicationControllerNotifierSerializableSlice = append(replicationControllerNotifierSerializableSlice, replicationControllerNotifierSerializable)
		}
	}

	response.WriteJson(replicationControllerNotifierSerializableSlice, "[]ReplicationControllerNotifierSerializable")
}

func getReplicationControllerNotifier(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	kind := request.PathParameter("kind")
	name := request.PathParameter("name")

	exist, replicationControllerNotifier := execute.GetReplicationControllerNotifier(namespace, kind, name)
	if exist == false {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The replication controller notifer doesn't exist"
		jsonMap["namespace"] = namespace
		jsonMap["kind"] = kind
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	replicationControllerNotifierSerializable, err := notification.ConvertToSerializable(replicationControllerNotifier)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Convert replication controller notifer failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["replicationControllerNotifier"] = replicationControllerNotifier
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(replicationControllerNotifierSerializable, "ReplicationControllerNotifierSerializable")
}

func putReplicationControllerNotifier(request *restful.Request, response *restful.Response) {
	replicationControllerNotifierSerializable := new(notification.ReplicationControllerNotifierSerializable)
	err := request.ReadEntity(&replicationControllerNotifierSerializable)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	switch replicationControllerNotifierSerializable.Kind {
	case "selector":
		nameSlice, err := monitor.GetReplicationControllerNameFromSelector(replicationControllerNotifierSerializable.KubeapiHost, replicationControllerNotifierSerializable.KubeapiPort, replicationControllerNotifierSerializable.Namespace, replicationControllerNotifierSerializable.Name)
		if err != nil {
			for _, name := range nameSlice {
				exist, err := monitor.ExistReplicationController(replicationControllerNotifierSerializable.KubeapiHost, replicationControllerNotifierSerializable.KubeapiPort, replicationControllerNotifierSerializable.Namespace, name)
				if err != nil {
					jsonMap := make(map[string]interface{})
					jsonMap["Error"] = "Check whether the replication controller exists or not failure"
					jsonMap["ErrorMessage"] = err.Error()
					jsonMap["replicationControllerNotifierSerializable"] = replicationControllerNotifierSerializable
					errorMessageByteSlice, _ := json.Marshal(jsonMap)
					log.Error(jsonMap)
					response.WriteErrorString(422, string(errorMessageByteSlice))
					return
				}
				if exist == false {
					jsonMap := make(map[string]interface{})
					jsonMap["Error"] = "The replication controller to notify doesn't exist"
					jsonMap["ErrorMessage"] = err.Error()
					jsonMap["replicationControllerNotifierSerializable"] = replicationControllerNotifierSerializable
					errorMessageByteSlice, _ := json.Marshal(jsonMap)
					log.Error(jsonMap)
					response.WriteErrorString(404, string(errorMessageByteSlice))
					return
				}
			}
		}
	case "replicationController":
		exist, err := monitor.ExistReplicationController(replicationControllerNotifierSerializable.KubeapiHost, replicationControllerNotifierSerializable.KubeapiPort, replicationControllerNotifierSerializable.Namespace, replicationControllerNotifierSerializable.Name)
		if err != nil {
			jsonMap := make(map[string]interface{})
			jsonMap["Error"] = "Check whether the replication controller exists or not failure"
			jsonMap["ErrorMessage"] = err.Error()
			jsonMap["replicationControllerNotifierSerializable"] = replicationControllerNotifierSerializable
			errorMessageByteSlice, _ := json.Marshal(jsonMap)
			log.Error(jsonMap)
			response.WriteErrorString(422, string(errorMessageByteSlice))
			return
		}
		if exist == false {
			jsonMap := make(map[string]interface{})
			jsonMap["Error"] = "The replication controller to notify doesn't exist"
			jsonMap["ErrorMessage"] = err.Error()
			jsonMap["replicationControllerNotifierSerializable"] = replicationControllerNotifierSerializable
			errorMessageByteSlice, _ := json.Marshal(jsonMap)
			log.Error(jsonMap)
			response.WriteErrorString(400, string(errorMessageByteSlice))
			return
		}
	default:
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "No such kind"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["replicationControllerNotifierSerializable"] = replicationControllerNotifierSerializable
		jsonMap["kind"] = replicationControllerNotifierSerializable.Kind
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	replicationControllerNotifier, err := notification.ConvertFromSerializable(*replicationControllerNotifierSerializable)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Convert replication controller notifier failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["replicationControllerNotifierSerializable"] = replicationControllerNotifierSerializable
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = notification.GetStorage().SaveReplicationControllerNotifierSerializable(replicationControllerNotifierSerializable)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save replication controller notifier failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["replicationControllerNotifierSerializable"] = replicationControllerNotifierSerializable
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	execute.AddReplicationControllerNotifier(&replicationControllerNotifier)
}

func deleteReplicationControllerNotifier(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	kind := request.PathParameter("kind")
	name := request.PathParameter("name")

	replicationControllerNotifier := &notification.ReplicationControllerNotifier{}
	replicationControllerNotifier.Namespace = namespace
	replicationControllerNotifier.Kind = kind
	replicationControllerNotifier.Name = name
	replicationControllerNotifier.Check = false

	err := notification.GetStorage().DeleteReplicationControllerNotifierSerializable(namespace, kind, name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete replication controller notifier failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		jsonMap["kind"] = kind
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	execute.AddReplicationControllerNotifier(replicationControllerNotifier)
}

func getAllEmailServerSMTP(request *restful.Request, response *restful.Response) {
	emailServerSMTPSlice, err := notification.GetStorage().LoadAllEmailServerSMTP()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all smtp email server failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(emailServerSMTPSlice, "[]EmailServerSMTP")
}

func postEmailServerSMTP(request *restful.Request, response *restful.Response) {
	emailServerSMTP := &notification.EmailServerSMTP{}
	err := request.ReadEntity(&emailServerSMTP)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	existingEmailServerSMTP, _ := notification.GetStorage().LoadEmailServerSMTP(emailServerSMTP.Name)
	if existingEmailServerSMTP != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The smtp email server to create already exists"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = emailServerSMTP.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(409, string(errorMessageByteSlice))
		return
	}

	err = notification.GetStorage().SaveEmailServerSMTP(emailServerSMTP)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save smtp email server failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["emailServerSMTP"] = emailServerSMTP
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func getEmailServerSMTP(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	emailServerSMTP, err := notification.GetStorage().LoadEmailServerSMTP(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get smtp email server failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(emailServerSMTP, "EmailServerSMTP")
}

func deleteEmailServerSMTP(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	err := notification.GetStorage().DeleteEmailServerSMTP(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete smtp email server failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func getAllSMSNexmo(request *restful.Request, response *restful.Response) {
	smsNexmoSlice, err := notification.GetStorage().LoadAllSMSNexmo()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all sms nexmo failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(smsNexmoSlice, "[]SMSNexmo")
}

func postSMSNexmo(request *restful.Request, response *restful.Response) {
	smsNexmo := &notification.SMSNexmo{}
	err := request.ReadEntity(&smsNexmo)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	existingSMSNexmo, _ := notification.GetStorage().LoadSMSNexmo(smsNexmo.Name)
	if existingSMSNexmo != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The sms nexmo to create already exists"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = smsNexmo.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(409, string(errorMessageByteSlice))
		return
	}

	err = notification.GetStorage().SaveSMSNexmo(smsNexmo)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save sms nexmo failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["smsNexmo"] = smsNexmo
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func getSMSNexmo(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	smsNexmo, err := notification.GetStorage().LoadSMSNexmo(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get sms nexmo failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(smsNexmo, "SMSNexmo")
}

func deleteSMSNexmo(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	err := notification.GetStorage().DeleteSMSNexmo(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete sms nexmo failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func returns200AllReplicationControllerNotifier(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []notification.ReplicationControllerNotifierSerializable{})
}

func returns200ReplicationControllerNotifier(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", notification.ReplicationControllerNotifierSerializable{})
}

func returns200AllEmailServerSMTP(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []notification.EmailServerSMTP{})
}

func returns200EmailServerSMTP(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", notification.EmailServerSMTP{})
}

func returns200AllSMSNexmo(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []notification.SMSNexmo{})
}

func returns200SMSNexmo(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", notification.SMSNexmo{})
}
