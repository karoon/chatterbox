package boxconfig

import (
	"encoding/json"
	"log"
	// "fmt"
	"os"
	"path"
	"runtime"
)

var configuration *Configuration

type Configuration struct {
	ProjectPath string
	Debug       bool
	Auth        AuthConfiguration
}

func init() {
	makeConfiguration()
}

func NewConfigHandler() *Configuration {
	if configuration != nil {
		return configuration
	}
	makeConfiguration()

	return configuration
}

func makeConfiguration() {
	file := getConfigFile()
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&configuration)
	if err != nil {
		panic(err)
	}

	_, filename, _, _ := runtime.Caller(0)
	rootDir := path.Dir(path.Dir(filename))
	log.Println(rootDir)
	configuration.ProjectPath = rootDir

	defer file.Close()
}

func getConfigFile() *os.File {
	_, filename, _, _ := runtime.Caller(0)
	file, err := os.Open(path.Join(path.Dir(filename), "conf.json"))
	if err != nil {
		panic(err)
	}
	return file
}
