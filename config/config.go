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
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	log "code.google.com/p/log4go"
	"github.com/BurntSushi/toml"
)

var (
	cfg  *Configuration
	lock = new(sync.RWMutex)
)

type ApiConfig struct {
	Port int
	Bind string
}

type StorageConfig struct {
	DSN            string
	MaxIdle        int
	MaxConnections int
}

type InfluxDbConfig struct {
	User     string
	Password string
	Host     string
	Database string
}

type LoggingConfig struct {
	File  string
	Level string
}

type NameServerConfig struct {
	Bind          string
	Port          int
	Secret        string
	LogQuery      bool
	CompressQuery bool
}

type TomlConfiguration struct {
	Api        ApiConfig
	Storage    StorageConfig
	Influx     InfluxDbConfig
	Logging    LoggingConfig
	NameServer NameServerConfig
}

type Configuration struct {
	ApiServerBind string
	ApiServerPort int

	StorageDSN            string
	StorageMaxIdle        int
	StorageMaxConnections int

	InfluxUser     string
	InfluxPassword string
	InfluxHost     string
	InfluxDb       string

	LogLevel string
	LogFile  string

	NameServerBind   string
	NameServerPort   int
	NameServerSecret string
	LogQuery         bool
	CompressQuery    bool
}

func LoadConfiguration(fileName string) (*Configuration, error) {
	log.Info("Loading configuration file %s", fileName)

	config, err := parseTomlConfiguration(fileName)
	if err != nil {
		fmt.Println("Couldn't parse configuration file: " + fileName)
		fmt.Println(err)
		return nil, err
	}

	lock.Lock()
	cfg = config
	lock.Unlock()
	return cfg, nil
}

func GetConfig() *Configuration {
	lock.RLock()
	defer lock.RUnlock()
	return cfg
}

func parseTomlConfiguration(filename string) (*Configuration, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	tomlConfiguration := &TomlConfiguration{}
	_, err = toml.Decode(string(body), tomlConfiguration)
	if err != nil {
		return nil, err
	}

	config := &Configuration{
		ApiServerBind: tomlConfiguration.Api.Bind,
		ApiServerPort: tomlConfiguration.Api.Port,

		StorageDSN:            tomlConfiguration.Storage.DSN,
		StorageMaxIdle:        tomlConfiguration.Storage.MaxIdle,
		StorageMaxConnections: tomlConfiguration.Storage.MaxConnections,

		InfluxUser:     tomlConfiguration.Influx.User,
		InfluxPassword: tomlConfiguration.Influx.Password,
		InfluxHost:     tomlConfiguration.Influx.Host,
		InfluxDb:       tomlConfiguration.Influx.Database,

		LogFile:  tomlConfiguration.Logging.File,
		LogLevel: tomlConfiguration.Logging.Level,

		NameServerBind:   tomlConfiguration.NameServer.Bind,
		NameServerPort:   tomlConfiguration.NameServer.Port,
		NameServerSecret: tomlConfiguration.NameServer.Secret,
		LogQuery:         tomlConfiguration.NameServer.LogQuery,
		CompressQuery:    tomlConfiguration.NameServer.CompressQuery,
	}
	return config, err
}

func parseJsonConfiguration(fileName string) (*Configuration, error) {
	log.Info("Loading Config from " + fileName)
	config := &Configuration{}

	data, err := ioutil.ReadFile(fileName)
	if err == nil {
		err = json.Unmarshal(data, config)
		if err != nil {
			return nil, err
		}
	} else {
		log.Error("Couldn't load configuration file: " + fileName)
		panic(err)
	}

	return config, nil
}

func (self *Configuration) ApiServerListen() string {
	if self.ApiServerPort <= 0 {
		return ":5080"
	}

	return fmt.Sprintf("%s:%d", self.ApiServerBind, self.ApiServerPort)
}

func (self *Configuration) NameServerListen() string {
	if self.NameServerPort <= 0 {
		return ":5053"
	}

	return fmt.Sprintf("%s:%d", self.NameServerBind, self.NameServerPort)
}
