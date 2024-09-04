package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func LoadConfig(path string) {
	viper.SetConfigFile(path)

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name)
		reloadConfig()
	})
}

func reloadConfig() {
	viper.ReadInConfig()
	log.Println("Configuration reloaded")
}

func GetConfig() *viper.Viper {
	return viper.GetViper()
}

func Init() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	configPath := dir + "/../config/config.yaml"
	LoadConfig(configPath)
}
