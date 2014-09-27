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
package nameserver

import (
	"strconv"
	"strings"
)

// Extract weight, port and dname from a srv string.
func extractSrv(srvString string) (uint16, uint16, string) {
	data := strings.Split(srvString, " ")

	var (
		w uint64
		p uint64
		d string
	)

	w, _ = strconv.ParseUint(data[0], 10, 16)
	p, _ = strconv.ParseUint(data[1], 10, 16)
	d = data[2]

	return uint16(w), uint16(p), d
}

func formatEmail(emailString string) string {
	return strings.Replace(emailString, "@", ".", -1) + "."
}
