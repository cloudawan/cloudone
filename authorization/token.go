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

package authorization

import (
	"errors"
	"github.com/cloudawan/cloudone_utility/rbac"
	jwt "github.com/dgrijalva/jwt-go"
	"time"
)

const (
	signingKey         = "userTokenKey"
	cacheCheckInterval = time.Minute
	cacheTTL           = cacheCheckInterval * 60
)

var SystemAdminToken string = ""

func init() {
	createDefaultUser()
	createSystemUserInMemory()
	periodicallyCleanCache()
}

func createDefaultUser() {
	user, _ := GetStorage().LoadUser("admin")
	if user == nil {
		permission := &rbac.Permission{"all", "*", "*", "*"}
		permissionSlice := make([]*rbac.Permission, 0)
		permissionSlice = append(permissionSlice, permission)
		role := &rbac.Role{"admin", permissionSlice, "admin"}
		roleSlice := make([]*rbac.Role, 0)
		roleSlice = append(roleSlice, role)
		resource := &rbac.Resource{"all", "*", "*"}
		resourceSlice := make([]*rbac.Resource, 0)
		resourceSlice = append(resourceSlice, resource)
		user := rbac.CreateUser("admin", "password", roleSlice, resourceSlice, "admin")

		if err := GetStorage().SaveRole(role); err != nil {
			log.Critical(err)
		}

		if err := GetStorage().SaveUser(user); err != nil {
			log.Critical(err)
		}
	}
}

func createSystemUserInMemory() {
	permission := &rbac.Permission{"system-all", "*", "*", "*"}
	permissionSlice := make([]*rbac.Permission, 0)
	permissionSlice = append(permissionSlice, permission)
	role := &rbac.Role{"system-admin", permissionSlice, "system-admin"}
	roleSlice := make([]*rbac.Role, 0)
	roleSlice = append(roleSlice, role)
	resource := &rbac.Resource{"system-all", "*", "*"}
	resourceSlice := make([]*rbac.Resource, 0)
	resourceSlice = append(resourceSlice, resource)
	// Use time as password and have it encrypted so no one other than system could use
	user := rbac.CreateUser("system", time.Now().String(), roleSlice, resourceSlice, "system-admin")

	token, err := generateToken(user)
	if err != nil {
		log.Critical(err)
		return
	}

	// Set the maximum duration
	rbac.SetCache(token, user, time.Duration(1<<63-1))
	SystemAdminToken = token
}

var closed bool = false

func Close() {
	closed = true
}

func periodicallyCleanCache() {
	go func() {
		for {
			if closed {
				break
			}

			rbac.CheckCacheTimeout()

			time.Sleep(cacheCheckInterval)
		}
	}()
}

func GetUserFromToken(token string) (*rbac.User, error) {
	_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}

		expiredText, _ := token.Claims["expired"].(string)

		expiredTime, err := time.Parse(time.RFC3339, expiredText)
		if err != nil {
			log.Error("Fail to parse expired time. Token %v error %s", token, err)
			return nil, err
		}

		if expiredTime.Before(time.Now()) {
			log.Debug("Token is expired. Token %v ", token)
			return nil, errors.New("Token is expired")
		}
		return []byte(signingKey), nil
	})

	if err != nil {
		return nil, err
	} else {
		user := rbac.GetCache(token)
		if user != nil {
			return user, nil
		} else {
			log.Error("User not in the cache. Token %v", token)
			return nil, errors.New("User not in the cache")
		}
	}
}

func CreateToken(name string, password string) (string, error) {
	user, err := GetStorage().LoadUser(name)
	if err != nil {
		log.Error(err)
		return "", err
	}

	if user.CheckPassword(password) == false {
		log.Error("Incorrect User %s or Password %s", name, password)
		return "", errors.New("Incorrect User or Password")
	}

	return generateToken(user)
}

func generateToken(user *rbac.User) (string, error) {
	// Create the token
	token := jwt.New(jwt.SigningMethodHS512)
	// Set some claims
	token.Claims["username"] = user.Name
	token.Claims["expired"] = time.Now().Add(cacheTTL).Format(time.RFC3339)
	// Sign
	signedToken, err := token.SignedString([]byte(signingKey))
	if err != nil {
		log.Error(err)
		return "", err
	}

	rbac.SetCache(signedToken, user, cacheTTL)

	// Sign and get the complete encoded token as a string
	return signedToken, nil
}

func GetAllTokenExpiredTime() map[string]time.Time {
	return rbac.GetAllTokenExpiredTime()
}
