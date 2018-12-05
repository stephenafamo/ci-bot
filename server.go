package main

import (
	"net/http"
)

type server struct {
	Router       http.Handler
	Handlers     Handlers
	Projects     map[string]Project
	Builds       chan Build
	Interactions chan SlackInteraction
}

func NewServer() (*server, error) {
	s := &server{}

	s.Builds = make(chan Build, 5)
	s.Interactions = make(chan SlackInteraction, 5)
	s.Projects = make(map[string]Project)
	s.Handlers = make(map[string]func() http.HandlerFunc)
	s.load()

	return s, nil
}

func (s *server) load() {
	s.startProcessors() // to read from the channels
	s.addProjects()     // all our projects
	s.addHandlers()     // the handlers for our routes
	s.addRoutes()       // Setting up the routes
}
