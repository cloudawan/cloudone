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

package lock

import (
	"testing"
)

func TestLock(t *testing.T) {
	for i := 0; i < 10000; i++ {
		if AcquireLock("build", "aaa", 0) == false {
			t.Errorf("Iteration %d. The first time acquire lock should not be false", i)
		}

		if AcquireLock("build", "aaa", 0) == true {
			t.Errorf("Iteration %d. The second time acquire lock should not be true", i)
		}

		if ReleaseLock("build", "aaa") != nil {
			t.Errorf("Iteration %d. The first time release lock should be nil", i)
		}
	}
}
