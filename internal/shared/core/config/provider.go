package config

import (
	"log"
	"sync"
)

var (
	instance *Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {

		LoadEnv()
		instance = NewConfig()

		log.Println("Config initialized successfully")
	})

	return instance
}
