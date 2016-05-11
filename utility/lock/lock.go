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
	"time"
)

const (
	LockDefaultTimeout = time.Hour
)

type Lock struct {
	Name        string
	CreatedTime time.Time
	ExpiredTime time.Time
}

func getLockName(kind string, name string) string {
	return kind + "." + name
}

// timeout 0 means default time out
func AcquireLock(kind string, name string, timeout time.Duration) bool {
	lockName := getLockName(kind, name)

	currentTime := time.Now()
	oldLock, _ := GetStorage().loadLock(lockName)
	if oldLock != nil {
		if currentTime.Before(oldLock.ExpiredTime) {
			return false
		}
	}

	if timeout == 0 {
		timeout = LockDefaultTimeout
	}

	// Acquire
	lock := &Lock{
		lockName,
		currentTime,
		currentTime.Add(timeout),
	}
	err := GetStorage().saveLock(lock)
	if err != nil {
		log.Error(err)
		return false
	} else {
		return true
	}
}

func ReleaseLock(kind string, name string) error {
	lockName := getLockName(kind, name)

	err := GetStorage().deleteLock(lockName)
	if err != nil {
		log.Error(err)
		return err
	} else {
		return nil
	}
}
