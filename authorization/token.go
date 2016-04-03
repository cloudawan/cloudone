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

func init() {
	periodicallyCleanCache()
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
			log.Error("Fail to parse expired time. Token %s error %s", token, err)
			return nil, err
		}

		if expiredTime.Before(time.Now()) {
			log.Error("Token is expired. Token %s ", token)
			return nil, errors.New("Token is expired")
		}
		return []byte(signingKey), nil
	})

	if err != nil {
		log.Error(err)
		return nil, err
	} else {
		user := rbac.GetCache(token)
		if user != nil {
			return user, nil
		} else {
			log.Error("User not in the cache. Token %s", token)
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

	// Create the token
	token := jwt.New(jwt.SigningMethodHS512)
	// Set some claims
	token.Claims["username"] = name
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
