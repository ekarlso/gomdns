/*
 * Copyright (c) 2014 Hewlett-Packard Development Company, L.P.
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ekarlso/gomdns/stats"
)

var meters map[string]stats.Meter

func startPoll() {
	timer := time.Tick(time.Second * 2)
	for {
		select {
		case <-timer:
			PollStats()
		}
	}
}

func PollStats() {
	url := *host + "/stats"

	response, err := http.Get(url)
	if err != nil {
		panic(err)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)

		if err != nil {
			fmt.Println(err)
		}

		meters = map[string]stats.Meter{}

		err = json.Unmarshal(contents, &meters)
		if err != nil {
			fmt.Println(err)
		}
	}
}
