package shared

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestHandleGetOffers(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	c := &VersionController{}

	r, _ := http.NewRequest("GET", "/api/version", nil)

	c.HandleGetVersion()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	versionModel := &versionModel{}
	err := json.NewDecoder(httpRec.Body).Decode(versionModel)
	is.NoErr(err)
	is.Equal(versionModel.Version, "unknown")
}
func TestRenderJSON(t *testing.T) {
	is := is.New(t)

	w := httptest.NewRecorder()

	c := struct {
		Name string
		Type string
	}{
		Name: "Sammy",
		Type: "Shark",
	}

	RenderJSON(w, c)

	is.True(strings.Contains(w.Body.String(), "Shark"))
}

func TestRenderProblemJSONInProduction(t *testing.T) {
	is := is.New(t)

	isProduction := true
	err := errors.New("my error")
	w := httptest.NewRecorder()

	RenderProblemJSON(w, isProduction, err)

	is.True(!strings.Contains(w.Body.String(), "my error"))
}

func TestRenderProblemJSON(t *testing.T) {
	is := is.New(t)

	isProduction := false
	err := errors.New("my error")
	w := httptest.NewRecorder()

	RenderProblemJSON(w, isProduction, err)

	is.True(strings.Contains(w.Body.String(), "my error"))
}
