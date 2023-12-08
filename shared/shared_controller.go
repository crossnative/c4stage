package shared

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"schneider.vip/problem"
)

type DomainHandler interface {
	RegisterProtected(router chi.Router)
	RegisterOpen(router chi.Router)
}

type versionModel struct {
	Version        string `json:"version"`
	BuildTime      string `json:"buildTime"`
	BuildGoVersion string `json:"buildGoVersion"`
}

type VersionController struct {
}

func (a *VersionController) RegisterProtected(router chi.Router) {
}

func (a *VersionController) RegisterOpen(router chi.Router) {
	router.Get("/version", a.HandleGetVersion())
}

// HandleGetVersion reads the current version
func (a *VersionController) HandleGetVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		version := GetVersion()

		versionModel := &versionModel{
			Version:        version.Version,
			BuildTime:      version.BuildTime,
			BuildGoVersion: version.BuildGoVersion,
		}

		RenderJSON(w, versionModel)
	}
}

func RenderJSON(w http.ResponseWriter, jsonModel interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(jsonModel)
	if err != nil {
		http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusInternalServerError)
	}
}

func RenderProblemJSON(w http.ResponseWriter, isProduction bool, err error) {
	log.Printf("internal server error: %s", err)

	if !isProduction {
		http.Error(w, problem.New(problem.Title("internal server error"), problem.Wrap(err)).JSONString(), http.StatusInternalServerError)
		return
	}

	http.Error(w, problem.New(problem.Title("internal server error")).JSONString(), http.StatusInternalServerError)
}
