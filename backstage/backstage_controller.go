package backstage

import (
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.io/remast/c4stage/catalog"
	"github.io/remast/c4stage/shared"
)

type ImportController struct {
	Config            *shared.Config
	CatalogRepository catalog.CatalogRepository
}

func (c *ImportController) RegisterProtected(router chi.Router) {
}

func (c *ImportController) RegisterOpen(router chi.Router) {
	r := chi.NewRouter()
	router.Mount("/backstage", r)

	r.Get("/import", c.HandleImportFromBackstage())
}

func (c *ImportController) HandleImportFromBackstage() http.HandlerFunc {
	backstageImportService := BackstageImporter{
		Config:     c.Config,
		Repository: c.CatalogRepository,
	}
	isProduction := c.Config.IsProduction()

	return func(w http.ResponseWriter, r *http.Request) {
		err := c.CatalogRepository.Reset(r.Context())
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		log.Println("Importing Backstage Catalog ...")
		io.WriteString(w, "Importing Backstage Catalog ...")
		err = backstageImportService.ImportBackstageCatalog()
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
		}
		log.Println("Successfully imported Backstage Catalog.")
		io.WriteString(w, "Successfully imported Backstage Catalog.")
	}
}
