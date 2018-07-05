package server

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/uc-cdis/arborist/arborist"
)

func handleEngineSerialize(engine *arborist.Engine) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := engine.HandleEngineSerialize()
		err := response.Write(w, wantPrettyJSON(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func addEngineRouter(mainRouter *mux.Router, engine *arborist.Engine) {
	engineRouter := mainRouter.PathPrefix("/engine").Subrouter()
	engineRouter.Handle("/", handleEngineSerialize(engine)).Methods("GET")
}
