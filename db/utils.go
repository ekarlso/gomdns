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

	"github.com/ekarlso/gomdns/config"
)

var Database *sqlx.DB

// Connect and return a connection
func Setup(cfg *config.Configuration) (err error) {
	log.Info("Connecting to %s", cfg.StorageDSN)

	db, err := sqlx.Open("mysql", cfg.StorageDSN)
	if err != nil {
		log.Error(err)
		return err
	}

	err = db.Ping()
	if err != nil {
		log.Error("Error connecting to db %s", err)
		return err
	}

	db.SetMaxOpenConns(cfg.StorageMaxConnections)
	db.SetMaxIdleConns(cfg.StorageMaxIdle)

	Database = db
	return nil
}

// Check that the Database is valid.
func CheckDB(dsn string) bool {
	_, err := Database.Exec("SELECT * FROM domains")
	if err != nil {
		log.Error(err)
		return false
	}

	return true
}
