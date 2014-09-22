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
	log "code.google.com/p/log4go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Connect and return a connection
func Connect(dsn string) *sqlx.DB {
	log.Info("Connecting to %s", dsn)
	conn, err := sqlx.Open("mysql", dsn)

	if err != nil {
		log.Crash(err)
	}

	return conn
}

// Check that the DB is valid.
func CheckDB(dsn string) bool {
	conn := Connect(dsn)

	_, err := conn.Exec("SELECT * FROM domains")
	if err != nil {
		log.Crash(err)
	}

	return true
}
