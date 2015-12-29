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

package healthcheck

import (
	"fmt"
	"testing"
)

func TestGetAllNodeStatus(t *testing.T) {
	jsonMap, err := GetAllNodeStatus()
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(jsonMap)
	}
	fmt.Println(jsonMap["kubernetes"])
	a, ok := jsonMap["kubernetes"].(map[string]interface{})
	fmt.Println(ok)
	fmt.Println(a["192.168.0.31"])
	b, ok := a["192.168.0.31"].(string)
	fmt.Println(ok)
	fmt.Println(b)
}
