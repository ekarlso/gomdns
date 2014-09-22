package config

import (
	log "code.google.com/p/log4go"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type StorageConfig struct {
	DSN     string
	MaxIdle int
}

type ServerConfig struct {
	Address string
	Port    int
}

type TomlConfiguration struct {
	Storage StorageConfig
	Server  ServerConfig
}

type Configuration struct {
	StorageDSN     string
	StorageMaxIdle int
	Bind           string
}

func LoadConfiguration(fileName string) (*Configuration, error) {
	log.Info("Loading configuration file %s", fileName)
	config, err := parseTomlConfiguration(fileName)
	if err != nil {
		fmt.Println("Couldn't parse configuration file: " + fileName)
		fmt.Println(err)
		return nil, err
	}

	return config, nil
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
		StorageDSN:     tomlConfiguration.Storage.DSN,
		StorageMaxIdle: tomlConfiguration.Storage.MaxIdle,
		Bind:           tomlConfiguration.Server.Address,
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
