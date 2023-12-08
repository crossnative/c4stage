package catalog

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.io/remast/c4stage/shared"
	"github.io/remast/c4stage/shared/paged"
)

type CatalogController struct {
	Config     *shared.Config
	Repository CatalogRepository
}

type systemsModel struct {
	Data        []System `json:"data"`
	*paged.Page `json:"page"`
}

func (c *CatalogController) RegisterProtected(router chi.Router) {
}

func (c *CatalogController) RegisterOpen(router chi.Router) {
	r := chi.NewRouter()
	router.Mount("/catalog", r)

	r.Get("/", c.HandleGetSystems())
}

func (c *CatalogController) HandleGetSystems() http.HandlerFunc {
	isProduction := c.Config.IsProduction()

	return func(w http.ResponseWriter, r *http.Request) {
		pageParams := paged.PageParamsOf(r)
		systems, page, err := c.Repository.FindSystems(r.Context(), pageParams)
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		systemsModel := systemsModel{
			Data: systems,
			Page: page,
		}

		shared.RenderJSON(w, systemsModel)
	}
}
