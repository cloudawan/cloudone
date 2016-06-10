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
	"github.com/coreos/etcd/client"
	"time"
)

const (
	LockDefaultTimeout       = time.Hour
	ReleaseLockRetryAmount   = 3
	ReleaseLockRetryInterval = time.Second * 10
)

type Lock struct {
	Name        string
	Timeout     time.Duration
	CreatedTime time.Time
	ExpiredTime time.Time
	Deleted     bool
}

func getLockName(kind string, name string) string {
	return kind + "." + name
}

func LockAvailable(kind string, name string) bool {
	lockName := getLockName(kind, name)
	currentTime := time.Now()

	oldLock, err := GetStorage().loadLock(lockName)
	if err != nil {
		etcdError, _ := err.(client.Error)
		if etcdError.Code != client.ErrorCodeKeyNotFound {
			log.Error(err)
			return false
		}
	}
	if oldLock != nil {
		if oldLock.Deleted == false {
			if currentTime.Before(oldLock.ExpiredTime) {
				return false
			}
		}
	}

	return true
}

// timeout 0 means default time out
func AcquireLock(kind string, name string, timeout time.Duration) bool {
	lockName := getLockName(kind, name)
	currentTime := time.Now()

	oldLock, err := GetStorage().loadLock(lockName)
	if err != nil {
		etcdError, _ := err.(client.Error)
		if etcdError.Code != client.ErrorCodeKeyNotFound {
			log.Error(err)
			return false
		}
	}
	if oldLock != nil {
		if oldLock.Deleted == false {
			if currentTime.Before(oldLock.ExpiredTime) {
				return false
			}
		}
	}

	if timeout == 0 {
		timeout = LockDefaultTimeout
	}

	// Acquire
	lock := &Lock{
		lockName,
		timeout,
		currentTime,
		currentTime.Add(timeout),
		false,
	}
	err = GetStorage().saveLock(lock)
	if err != nil {
		log.Error(err)
		return false
	} else {
		return true
	}
}

func ReleaseLock(kind string, name string) error {
	lockName := getLockName(kind, name)

	currentTime := time.Now()
	lock := &Lock{
		lockName,
		LockDefaultTimeout,
		currentTime,
		currentTime.Add(LockDefaultTimeout),
		true,
	}

	var err error = nil
	for i := 0; i < ReleaseLockRetryAmount; i++ {
		time.Sleep(ReleaseLockRetryInterval * time.Duration(i))

		err = GetStorage().saveLock(lock)
		if err == nil {
			return nil
		} else {
			log.Error("The %d time retry to release lock with error %s", i, err)
		}
	}

	return err
}
