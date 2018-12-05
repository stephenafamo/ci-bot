package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type Handlers map[string]func() http.HandlerFunc

// The is used to load a handler by name
// It sends a http 500 error if the handler does not exist
func (h Handlers) Use(name string) http.HandlerFunc {
	method, ok := h[name]
	if ok == false {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Handler not found", 500)
		}
	}
	return method()
}

func (s *server) addHandlers() {

	s.Handlers["404"] = func() http.HandlerFunc {
		// This handles 404 errors
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "404: Page not found", 404)
		}
	}

	s.Handlers["BuildComplete"] = func() http.HandlerFunc {
		// This handles build notifications and sends them
		// into the server Builds channel
		return func(w http.ResponseWriter, r *http.Request) {

			projectName := r.FormValue("project")
			project, ok := s.Projects[projectName]
			if !ok {
				log.Println(errors.New("Project " + projectName + " not found"))
				http.Error(w, "Error encountered", 500)
				return
			}

			build := Build{}
			build.Project = project
			build.Image = r.FormValue("image")   //docker image
			build.Target = r.FormValue("target") // name of the branch or tag
			build.Type = r.FormValue("type")     // branch or tag

			s.Builds <- build
			w.Write([]byte("Received successfully"))
		}
	}

	s.Handlers["SlackInteractions"] = func() http.HandlerFunc {
		// This handles slack interactions and sends them
		// into the server Interactions channel
		return func(w http.ResponseWriter, r *http.Request) {

			var theResp SlackInteraction
			bodyBytes := []byte(r.FormValue("payload"))

			err := json.Unmarshal(bodyBytes, &theResp)
			if err != nil {
				log.Println(err)
				return
			}

			switch theResp.Type {
			case "interactive_message":
				s.Interactions <- theResp
				return
			default:
				fmt.Println("Unknown Interaction", string(bodyBytes))
			}
		}
	}
}
