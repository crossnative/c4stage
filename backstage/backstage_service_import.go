package backstage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.io/remast/c4stage/catalog"
	"github.io/remast/c4stage/shared"
	"gopkg.in/yaml.v3"
)

type BackstageImporter struct {
	Config     *shared.Config
	Repository catalog.CatalogRepository
}

func (i BackstageImporter) ImportBackstageCatalog() error {
	backstageURL := fmt.Sprintf(
		"%v/api/catalog/entities?offset=0&limit=500&filter=kind=component&filter=kind=system&filter=kind=api",
		i.Config.BackstageServer,
	)
	response, err := http.Get(backstageURL)
	if err != nil {
		return err
	}

	var rawEntities []RawEntity
	err = json.NewDecoder(response.Body).Decode(&rawEntities)
	if err != nil {
		return err
	}

	var entities []any
	for _, rawEntity := range rawEntities {
		entity, _ := rawEntity.FromRaw()
		entities = append(entities, entity)
	}

	return i.Repository.CreateAll(context.Background(), entities)
}

func (i BackstageImporter) ImportYamlFiles() error {
	files, _ := filepath.Glob("/Users/jsta/Documents/projects/backstage/examples/sldo/**/catalog-info.yaml")

	var entities []any

	for _, f := range files {
		fileBytes, _ := os.ReadFile(f)
		decoder := yaml.NewDecoder(bytes.NewBuffer(fileBytes))
		for {
			var d RawEntity
			err := decoder.Decode(&d)
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("document decode failed: %w", err)
			}

			entity, err := d.FromRaw()
			if err != nil {
				return err
			}

			entities = append(entities, entity)
		}
	}

	return i.Repository.CreateAll(context.Background(), entities)
}
