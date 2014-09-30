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
package stats

// Representation of a Meter
type Meter struct {
	Count    int64   `json:"count"`
	Rate1    float64 `json:"1m.rate"`
	Rate5    float64 `json:"5m.rate"`
	Rate15   float64 `json:"15m.rate"`
	RateMean float64 `json:"mean.rate"`
}

func (s Meter) IsValid() bool { return s.Count != 0 }
