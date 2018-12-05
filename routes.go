package main

import (
	"github.com/go-chi/chi"
)

func (s *server) addRoutes() {
	r := chi.NewRouter()
	r.NotFound(s.Handlers.Use("404")) // A route for 404s
	r.Post("/build-complete", s.Handlers.Use("BuildComplete"))
	r.Post("/slack-interactions", s.Handlers.Use("SlackInteractions"))
	s.Router = r
}
