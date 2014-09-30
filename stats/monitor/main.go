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
	"flag"
	"fmt"
	"io/ioutil"
	"runtime/debug"
	"time"

	"github.com/nsf/termbox-go"
)

var (
	host     = flag.String("host", "http://localhost:5080", "Host to connect to for stats")
	interval = flag.Int("interval", 5, "Interval to poll stats")
)

func main() {
	flag.Parse()

	defer func() {
		if e := recover(); e != nil {
			termbox.Close()
			trace := fmt.Sprintf("%s: %s", e, debug.Stack()) // line 20
			ioutil.WriteFile("trace.txt", []byte(trace), 0644)
		}
	}()

	go startCli()
	go startPoll()

	for {
		time.Sleep(time.Millisecond * 100)
		if quit == true {
			break
		}
	}

	termbox.Close()
}
