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
package db

import (
	"database/sql"
)

// Designate Zone
type Zone struct {
	Id      string `db:"id"`
	Version int
	Name    string
	Email   string
	Ttl     uint32
	Serial  int
	Refresh int
	Retry   int
	Expire  int
	Minimum int
}

// Records interface and helpers
type Records []*Record

func (s Records) Len() int      { return len(s) }
func (s Records) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByPriority struct{ Records }

func (s ByPriority) Less(i, j int) bool {
	var (
		x int64
		y int64
	)

	x, y = 0, 0

	if s.Records[i].Priority.Valid {
		x = s.Records[i].Priority.Int64
	}
	if s.Records[j].Priority.Valid {
		y = s.Records[j].Priority.Int64
	}

	return x < y
}

type RecordSet struct {
	Id       string
	DomainId string `db:"domain_id"`
	Name     string
	Type     string
	Ttl      sql.NullInt64
	Records  Records
}

type Record struct {
	Id          string
	DomainId    string `db:"domain_id"`
	RecordSetId string `db:"recordset_id"`
	Data        string
	Priority    sql.NullInt64
	Hash        string
}
