package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

type Build struct {
	Project Project
	Target  string
	Image   string
	Type    string
}

// We use viper here to load configuration from a config.yml file
func setupConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err)
		return
	}
}

func init() {
	setupConfig()
}

func main() {

	s, err := NewServer()
	if err != nil {
		panic(err)
	}

	http.Handle("/", s.Router)
	fmt.Println("listening on port 80")

	err = http.ListenAndServe(":80", s.Router)
	if err != nil {
		panic(err)
	}
}
