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

package execute

import (
	"time"
)

var quitChannel = make(chan struct{})

func Close() {
	close(quitChannel)
}

func init() {
	loop(1*time.Second, loopAutoScaler)
	loop(1*time.Second, loopNotifier)
}

type functionLoop func(ticker *time.Ticker, checkingInterval time.Duration)

func loop(checkingInterval time.Duration, function functionLoop) {
	ticker := time.NewTicker(checkingInterval)
	go function(ticker, checkingInterval)
}
