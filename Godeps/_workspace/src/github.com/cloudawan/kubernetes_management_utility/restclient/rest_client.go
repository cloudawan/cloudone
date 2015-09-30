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

package restclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

var insecureHTTPSClient *http.Client = nil

func GetInsecureHTTPSClient() *http.Client {
	// Skip the server side certificate checking
	if insecureHTTPSClient == nil {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		insecureHTTPSClient = &http.Client{Transport: transport}
		return insecureHTTPSClient
	} else {
		return insecureHTTPSClient
	}
}

func Request(method string, url string, body interface{},
	useJsonNumberInsteadFloat64ForResultJson bool) (returnedStatusCode int,
	returnedJsonMapOrJsonSlice interface{}, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			returnedStatusCode = 500
			returnedJsonMapOrJsonSlice = nil
			returnedError = err.(error)
		}
	}()

	byteSlice, err := json.Marshal(body)
	if err != nil {
		return 500, nil, err
	} else {
		var request *http.Request
		if body == nil {
			request, err = http.NewRequest(method, url, nil)
		} else {
			request, err = http.NewRequest(method, url, bytes.NewReader(byteSlice))
			request.Header.Add("Content-Type", "application/json")
		}

		if err != nil {
			return 500, nil, err
		} else {
			response, err := GetInsecureHTTPSClient().Do(request)
			if err != nil {
				return 500, nil, err
			} else {
				defer response.Body.Close()
				if response.ContentLength == 0 {
					return response.StatusCode, nil, nil
				}
				responseBody, err := ioutil.ReadAll(response.Body)
				if err != nil {
					return 500, nil, err
				} else {
					if response.StatusCode == 404 {
						return 404, nil, nil
					} else {
						var jsonMap interface{}
						if useJsonNumberInsteadFloat64ForResultJson {
							decoder := json.NewDecoder(bytes.NewReader(responseBody))
							decoder.UseNumber()
							err := decoder.Decode(&jsonMap)
							if err != nil {
								jsonMap = make(map[string]interface{})
								jsonMap.(map[string]interface{})["body"] = string(responseBody)
								return response.StatusCode, jsonMap, err
							} else {
								return response.StatusCode, jsonMap, nil
							}
						} else {
							err := json.Unmarshal(responseBody, &jsonMap)
							if err != nil {
								jsonMap = make(map[string]interface{})
								jsonMap.(map[string]interface{})["body"] = string(responseBody)
								return response.StatusCode, jsonMap, err
							} else {
								return response.StatusCode, jsonMap, nil
							}
						}
					}
				}
			}
		}
	}
}

func RequestGet(url string, useJsonNumberInsteadFloat64ForResultJson bool) (interface{}, error) {
	statusCode, jsonMap, err := Request("GET", url, nil, useJsonNumberInsteadFloat64ForResultJson)
	if err != nil {
		return jsonMap, err
	} else if statusCode == 200 || statusCode == 204 {
		return jsonMap, nil
	} else {
		return jsonMap, errors.New("Status code: " + strconv.Itoa(statusCode) +
			" jsonMap: " + fmt.Sprintf("%s", jsonMap) + " url: " + url)
	}
}

func RequestPost(url string, body interface{}, useJsonNumberInsteadFloat64ForResultJson bool) (interface{}, error) {
	statusCode, jsonMap, err := Request("POST", url, body, useJsonNumberInsteadFloat64ForResultJson)
	if err != nil {
		return jsonMap, err
	} else if statusCode == 200 || statusCode == 201 || statusCode == 202 {
		return jsonMap, nil
	} else {
		return jsonMap, errors.New("Status code: " + strconv.Itoa(statusCode) +
			" jsonMap: " + fmt.Sprintf("%s", jsonMap) + " url: " + url)
	}
}

func RequestPut(url string, body interface{}, useJsonNumberInsteadFloat64ForResultJson bool) (interface{}, error) {
	statusCode, jsonMap, err := Request("PUT", url, body, useJsonNumberInsteadFloat64ForResultJson)
	if err != nil {
		return jsonMap, err
	} else if statusCode == 200 || statusCode == 202 || statusCode == 204 {
		return jsonMap, nil
	} else {
		return jsonMap, errors.New("Status code: " + strconv.Itoa(statusCode) +
			" jsonMap: " + fmt.Sprintf("%s", jsonMap) + " url: " + url)
	}
}

func RequestDelete(url string, body interface{}, useJsonNumberInsteadFloat64ForResultJson bool) (interface{}, error) {
	statusCode, jsonMap, err := Request("DELETE", url, body, useJsonNumberInsteadFloat64ForResultJson)
	if err != nil {
		return jsonMap, err
	} else if statusCode == 200 || statusCode == 202 || statusCode == 204 {
		return jsonMap, nil
	} else {
		return jsonMap, errors.New("Status code: " + strconv.Itoa(statusCode) +
			" jsonMap: " + fmt.Sprintf("%s", jsonMap) + " url: " + url)
	}
}

func RequestWithStructure(method string, url string, body interface{}, returnedStrucutre interface{}) (returnedStatusCode int, returnedDataStrucutre interface{}, returnedError error, returnedResponseBody *string) {
	defer func() {
		if err := recover(); err != nil {
			returnedStatusCode = 500
			returnedDataStrucutre = nil
			returnedError = err.(error)
			returnedResponseBody = nil
		}
	}()

	if returnedStrucutre == nil {
		returnedStrucutre = make(map[string]interface{})
	}

	byteSlice, err := json.Marshal(body)
	if err != nil {
		return 500, nil, err, nil
	} else {
		request, err := http.NewRequest(method, url, bytes.NewReader(byteSlice))
		request.Header.Add("Content-Type", "application/json")
		if err != nil {
			return 500, nil, err, nil
		} else {
			response, err := GetInsecureHTTPSClient().Do(request)
			if err != nil {
				return 500, nil, err, nil
			} else {
				defer response.Body.Close()
				if response.ContentLength == 0 {
					return response.StatusCode, nil, nil, nil
				}
				responseBody, err := ioutil.ReadAll(response.Body)
				if err != nil {
					return 500, nil, err, nil
				} else {
					if response.StatusCode == 404 {
						responseBodyText := string(responseBody)
						return 404, nil, nil, &responseBodyText
					} else {
						err := json.Unmarshal(responseBody, &returnedStrucutre)
						if err != nil {
							responseBodyText := string(responseBody)
							return response.StatusCode, nil, err, &responseBodyText
						} else {
							responseBodyText := string(responseBody)
							return response.StatusCode, returnedStrucutre, nil, &responseBodyText
						}
					}
				}
			}
		}
	}
}

func RequestGetWithStructure(url string, returnedStrucutre interface{}) (interface{}, error) {
	statusCode, data, err, responseBody := RequestWithStructure("GET", url, nil, returnedStrucutre)
	if err != nil {
		return data, err
	} else if statusCode == 200 || statusCode == 204 {
		return data, nil
	} else {
		if responseBody == nil {
			return data, errors.New("Status code: " + strconv.Itoa(statusCode))
		} else {
			return data, errors.New("Status code: " + strconv.Itoa(statusCode) +
				" Body: " + *responseBody + " url: " + url)
		}
	}
}

func RequestPostWithStructure(url string, body interface{}, returnedStrucutre interface{}) (interface{}, error) {
	statusCode, data, err, responseBody := RequestWithStructure("POST", url, body, returnedStrucutre)
	if err != nil {
		return data, err
	} else if statusCode == 200 || statusCode == 201 || statusCode == 202 {
		return data, nil
	} else {
		if responseBody == nil {
			return data, errors.New("Status code: " + strconv.Itoa(statusCode))
		} else {
			return data, errors.New("Status code: " + strconv.Itoa(statusCode) +
				" Body: " + *responseBody + " url: " + url)
		}
	}
}

func RequestPutWithStructure(url string, body interface{}, returnedStrucutre interface{}) (interface{}, error) {
	statusCode, data, err, responseBody := RequestWithStructure("PUT", url, body, returnedStrucutre)
	if err != nil {
		return data, err
	} else if statusCode == 200 || statusCode == 202 || statusCode == 204 {
		return data, nil
	} else {
		if responseBody == nil {
			return data, errors.New("Status code: " + strconv.Itoa(statusCode))
		} else {
			return data, errors.New("Status code: " + strconv.Itoa(statusCode) +
				" Body: " + *responseBody + " url: " + url)
		}
	}
}

func RequestDeleteWithStructure(url string, body interface{}, returnedStrucutre interface{}) (interface{}, error) {
	statusCode, data, err, responseBody := RequestWithStructure("DELETE", url, body, returnedStrucutre)
	if err != nil {
		return data, err
	} else if statusCode == 200 || statusCode == 202 || statusCode == 204 {
		return data, nil
	} else {
		if responseBody == nil {
			return data, errors.New("Status code: " + strconv.Itoa(statusCode))
		} else {
			return data, errors.New("Status code: " + strconv.Itoa(statusCode) +
				" Body: " + *responseBody + " url: " + url)
		}
	}
}

func RequestByteSliceResult(method string, url string, body map[string]interface{}) (returnedStatusCode int, returnedByteSlice []byte, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			returnedStatusCode = 500
			returnedByteSlice = nil
			returnedError = err.(error)
		}
	}()

	byteSlice, err := json.Marshal(body)
	if err != nil {
		return 500, nil, err
	} else {
		request, err := http.NewRequest(method, url, bytes.NewReader(byteSlice))
		request.Header.Add("Content-Type", "application/json")
		if err != nil {
			return 500, nil, err
		} else {
			response, err := GetInsecureHTTPSClient().Do(request)
			if err != nil {
				return 500, nil, err
			} else {
				defer response.Body.Close()
				if response.ContentLength == 0 {
					return response.StatusCode, nil, nil
				}
				responseBody, err := ioutil.ReadAll(response.Body)
				if err != nil {
					return 500, nil, err
				} else {
					return response.StatusCode, responseBody, nil
				}
			}
		}
	}
}

func RequestGetByteSliceResult(url string) ([]byte, error) {
	statusCode, byteSlice, err := RequestByteSliceResult("GET", url, nil)
	if err != nil {
		return byteSlice, err
	} else if statusCode == 200 || statusCode == 204 {
		return byteSlice, nil
	} else {
		return byteSlice, errors.New("Status code: " + strconv.Itoa(statusCode) +
			" byteSlice: " + fmt.Sprintf("%s", byteSlice) + " url: " + url)
	}
}

func RequestPostByteSliceResult(url string, body map[string]interface{}) ([]byte, error) {
	statusCode, byteSlice, err := RequestByteSliceResult("POST", url, body)
	if err != nil {
		return byteSlice, err
	} else if statusCode == 200 || statusCode == 201 || statusCode == 202 {
		return byteSlice, nil
	} else {
		return byteSlice, errors.New("Status code: " + strconv.Itoa(statusCode) +
			" byteSlice: " + fmt.Sprintf("%s", byteSlice) + " url: " + url)
	}
}

func RequestPutByteSliceResult(url string, body map[string]interface{}) ([]byte, error) {
	statusCode, byteSlice, err := RequestByteSliceResult("PUT", url, body)
	if err != nil {
		return byteSlice, err
	} else if statusCode == 200 || statusCode == 202 || statusCode == 204 {
		return byteSlice, nil
	} else {
		return byteSlice, errors.New("Status code: " + strconv.Itoa(statusCode) +
			" byteSlice: " + fmt.Sprintf("%s", byteSlice) + " url: " + url)
	}
}

func RequestDeleteByteSliceResult(url string, body map[string]interface{}) ([]byte, error) {
	statusCode, byteSlice, err := RequestByteSliceResult("DELETE", url, body)
	if err != nil {
		return byteSlice, err
	} else if statusCode == 200 || statusCode == 202 || statusCode == 204 {
		return byteSlice, nil
	} else {
		return byteSlice, errors.New("Status code: " + strconv.Itoa(statusCode) +
			" byteSlice: " + fmt.Sprintf("%s", byteSlice) + " url: " + url)
	}
}
