package c4

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.io/remast/c4stage/shared"
	"schneider.vip/problem"
)

type C4Controller struct {
	Config     *shared.Config
	Repository C4Repository
}

func (c *C4Controller) RegisterProtected(router chi.Router) {
}

func (c *C4Controller) RegisterOpen(router chi.Router) {
	r := chi.NewRouter()
	router.Mount("/c4", r)

	r.Get("/context", c.HandleGetSystemLandscapeDiagram())
	r.Get("/container", c.HandleGetSystemLandscapeContainerDiagram())
	r.Get("/{name}/container", c.HandleGetContainerDiagram())
}

func (c *C4Controller) HandleGetSystemLandscapeDiagram() http.HandlerFunc {
	e := newPlantUMLExporter()
	plantUMLServer := c.Config.PlantUMLServer
	isProduction := c.Config.IsProduction()

	return func(w http.ResponseWriter, r *http.Request) {
		c4Model, err := c.Repository.SystemLandscapeDiagram(r.Context())
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		// adjust scale
		scale, _ := strconv.ParseFloat(r.URL.Query().Get("scale"), 64)
		c4Model.Scale = scale

		// adjust image format
		imageFormat := r.URL.Query().Get("format")

		sw := bytes.NewBufferString("")

		err = e.ExportToPlantUMLContext(c4Model, sw)
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		renderPlantUML(
			w,
			sw.String(),
			imageFormat,
			plantUMLServer,
			isProduction,
		)
	}
}

func (c *C4Controller) HandleGetSystemLandscapeContainerDiagram() http.HandlerFunc {
	e := newPlantUMLExporter()
	plantUMLServer := c.Config.PlantUMLServer
	isProduction := c.Config.IsProduction()

	return func(w http.ResponseWriter, r *http.Request) {
		c4Model, err := c.Repository.SystemLandscapeContainerDiagram(r.Context())
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		// adjust scale
		scale, _ := strconv.ParseFloat(r.URL.Query().Get("scale"), 64)
		c4Model.Scale = scale

		// adjust image format
		imageFormat := r.URL.Query().Get("format")

		sw := bytes.NewBufferString("")

		err = e.ExportToPlantUMLContainer(c4Model, sw)
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		renderPlantUML(
			w,
			sw.String(),
			imageFormat,
			plantUMLServer,
			isProduction,
		)
	}
}

func (c *C4Controller) HandleGetContainerDiagram() http.HandlerFunc {
	e := newPlantUMLExporter()
	plantUMLServer := c.Config.PlantUMLServer
	isProduction := c.Config.IsProduction()

	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		c4Model, err := c.Repository.ContainerDiagram(r.Context(), name)
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		if c4Model.IsEmpty() {
			message := fmt.Sprintf("system %v not found", name)
			http.Error(w, problem.New(problem.Title(message)).JSONString(), http.StatusNotFound)
			return
		}

		// adjust scale
		scale, _ := strconv.ParseFloat(r.URL.Query().Get("scale"), 64)
		c4Model.Scale = scale

		// adjust image format
		imageFormat := r.URL.Query().Get("format")

		sw := bytes.NewBufferString("")

		err = e.ExportToPlantUMLContainer(c4Model, sw)
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		renderPlantUML(
			w,
			sw.String(),
			imageFormat,
			plantUMLServer,
			isProduction,
		)
	}
}

func renderPlantUML(w http.ResponseWriter, plantUMLBody string, imageFormat string, plantUMLServer string, isProduction bool) {
	outputFormat := "png"
	if imageFormat == "svg" {
		outputFormat = "svg"
	} else if imageFormat == "puml" {
		w.Write([]byte(plantUMLBody))
		return
	}

	response, err := http.Post(
		fmt.Sprintf("%v/%v", plantUMLServer, outputFormat),
		"text/plain; charset=UTF-8",
		strings.NewReader(plantUMLBody),
	)
	if err != nil {
		shared.RenderProblemJSON(w, isProduction, err)
		return
	}

	_, err = io.Copy(w, response.Body)
	if err != nil {
		shared.RenderProblemJSON(w, isProduction, err)
		return
	}
}
