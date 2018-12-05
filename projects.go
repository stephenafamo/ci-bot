package main

import (
	"log"

	"github.com/spf13/viper"
)

type Project struct {
	ID      string
	Name    string
	URL     string
	Channel string
	QA      []string
	Owners  []string
}

// This loads projectd defined in the config to the server
func (s *server) addProjects() {
	var projects []Project

	err := viper.UnmarshalKey("projects", &projects)
	if err != nil {
		log.Println(err)
		return
	}

	for _, p := range projects {
		s.Projects[p.ID] = p
	}
}
