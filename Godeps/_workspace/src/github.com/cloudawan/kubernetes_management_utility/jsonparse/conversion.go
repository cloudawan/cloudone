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

package jsonparse

import (
	"encoding/json"
	"strconv"
)

func ConvertToInt64(field interface{}) (int64, bool) {
	switch number := field.(type) {
	case float64:
		return int64(number), true
	case json.Number:
		if value, err := number.Int64(); err != nil {
			if _, ok := err.(*strconv.NumError); ok {
				if value, err := number.Float64(); err != nil {
					return 0, false
				} else {
					return int64(value), true
				}
			} else {
				return 0, false
			}
		} else {
			return value, true
		}
	default:
		return 0, false
	}
}

func ConvertToFloat64(field interface{}) (float64, bool) {
	switch number := field.(type) {
	case float64:
		return number, true
	case json.Number:
		if value, err := number.Float64(); err != nil {
			return 0, false
		} else {
			return value, true
		}
	default:
		return 0, false
	}
}
