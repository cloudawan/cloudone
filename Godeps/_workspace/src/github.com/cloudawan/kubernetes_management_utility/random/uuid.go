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

package random

import (
	"crypto/rand"
	"fmt"
	"io"
)

func UUID() string {
	// RFC4122
	uuid := make([]byte, 16)
	io.ReadFull(rand.Reader, uuid)
	// Section 4.1.1 variant bits
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// Section 4.1.3 version 4 (pseudo-random);
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
