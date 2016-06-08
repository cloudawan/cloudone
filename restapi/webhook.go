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
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/cloudawan/cloudone/webhook"
	"github.com/emicklei/go-restful"
)

type GithubPost struct {
	User             string
	ImageInformation string
	Signature        string
	Payload          string
}

func registerWebServiceWebhook() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/webhooks")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	// The webhook has its own verification
	ws.Route(ws.POST("/github/").Filter(auditLog).To(postGithub).
		Doc("Trigger image build from github webhook data").
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(Namesapce{}))
}

func postGithub(request *restful.Request, response *restful.Response) {
	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	githubPost := GithubPost{}
	err = request.ReadEntity(&githubPost)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = webhook.Notify(githubPost.User, githubPost.ImageInformation, githubPost.Signature, githubPost.Payload, kubeApiServerEndPoint, kubeApiServerToken)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Notify webhook failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}
