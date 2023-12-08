package catalog

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.io/remast/c4stage/shared/paged"
)

var _ CatalogRepository = (*CatalogRepositoryNeo4j)(nil)

var dependsOnWhitelist = map[string]string{
	"system":    "System",
	"component": "Component",
	"api":       "API",
}

type CatalogRepositoryNeo4j struct {
	Driver neo4j.DriverWithContext
}

func (r *CatalogRepositoryNeo4j) Reset(ctx context.Context) error {
	resetStatements := []string{
		"MATCH (n) DETACH DELETE n",
		"DROP CONSTRAINT system_name_idx IF EXISTS",
		`
		CREATE CONSTRAINT system_name_idx
		FOR (s:System) REQUIRE s.name IS UNIQUE`,
		"DROP CONSTRAINT component_name_idx IF EXISTS",
		`
		CREATE CONSTRAINT component_name_idx
		FOR (c:Component) REQUIRE c.name IS UNIQUE`,
	}

	for _, resetStatement := range resetStatements {
		_, err := neo4j.ExecuteQuery(ctx, r.Driver,
			resetStatement,
			map[string]any{}, neo4j.EagerResultTransformer)
		if err != nil {
			return err
		}

	}
	return nil
}

func (r *CatalogRepositoryNeo4j) FindSystems(
	ctx context.Context,
	pageParams *paged.PageParams,
) ([]System, *paged.Page, error) {
	result, err := neo4j.ExecuteQuery(ctx, r.Driver,
		`
		MATCH (s:System)
		WHERE s.type <> "person" AND s.type <> "external"
		RETURN s
		LIMIT $limit
		`,
		map[string]any{
			"limit": pageParams.Size,
		}, neo4j.EagerResultTransformer)

	if err != nil {
		return []System{}, nil, err
	}

	var systems []System

	for _, record := range result.Records {
		for _, recordValue := range record.Values {
			switch node := recordValue.(type) {
			case dbtype.Node:
				if slices.Contains(node.Labels, "System") {
					system := readSystem(node)
					systems = append(systems, *system)
				}
			}
		}
	}

	return systems, pageParams.PageOfTotal(len(systems)), nil
}

func (r *CatalogRepositoryNeo4j) CreateAll(
	ctx context.Context,
	entities []any,
) error {
	for _, entity := range entities {
		var err error
		switch e := entity.(type) {
		case System:
			_, err = neo4j.ExecuteQuery(ctx, r.Driver,
				`
				MERGE (s:System { name: $name })
				ON MATCH
					SET s.titleTmp = $title
					SET s.title = $title
					SET s.description = $description
					SET s.type = $type
					SET s.lifecycle = $lifecycle
				RETURN s
				`,
				map[string]any{
					"name":        e.Name,
					"title":       e.Title,
					"description": e.Description,
					"type":        e.Type,
					"lifecycle":   e.Lifecycle,
				}, neo4j.EagerResultTransformer)

			if len(e.DependsOn) > 0 {
				for _, dependsOn := range e.DependsOn {
					dependsOnKind, dependsOnName, err := parseDependsOn(dependsOn)
					if err != nil {
						return err
					}

					_, err = neo4j.ExecuteQuery(ctx, r.Driver,
						fmt.Sprintf("MERGE (n:%s { name: $name }) RETURN n", dependsOnKind),
						map[string]any{
							"name": dependsOnName,
						}, neo4j.EagerResultTransformer)
					if err != nil {
						return err
					}

					_, err = neo4j.ExecuteQuery(ctx, r.Driver,
						fmt.Sprintf(`
									MATCH (c:System{ name: $name})
									MATCH (d:%s{ name: $dependsOnName })
									MERGE (c)-[:DEPENDS_ON]->(d)`, dependsOnKind),
						map[string]any{
							"name":          e.Name,
							"dependsOnName": dependsOnName,
						}, neo4j.EagerResultTransformer)
					if err != nil {
						return err
					}
				}
			}
		case Container:
			_, err = neo4j.ExecuteQuery(ctx, r.Driver,
				`
				MERGE (c:Component { name: $name })
				ON MATCH
					SET c.titleTmp = $title
					SET c.title = $title
					SET c.description = $description
					SET c.system = $system
					SET c.type = $type
					SET c.lifecycle = $lifecycle
					SET c.tags = $tags
				RETURN c
				`,
				map[string]any{
					"name":        e.Name,
					"title":       e.Title,
					"system":      e.System,
					"description": e.Description,
					"type":        e.Type,
					"lifecycle":   e.Lifecycle,
					"tags":        e.Tags,
				}, neo4j.EagerResultTransformer)
			if err != nil {
				return err
			}
			if e.System != "" {
				_, err = neo4j.ExecuteQuery(ctx, r.Driver,
					"MERGE (n:System { name: $name }) RETURN n",
					map[string]any{
						"name": e.System,
					}, neo4j.EagerResultTransformer)
				if err != nil {
					return err
				}

				_, err = neo4j.ExecuteQuery(ctx, r.Driver,
					"MERGE (c:Component { name: $name }) RETURN c",
					map[string]any{
						"name": e.Name,
					}, neo4j.EagerResultTransformer)
				if err != nil {
					return err
				}
				_, err = neo4j.ExecuteQuery(ctx, r.Driver,
					`
					MATCH (s:System{ name: $system })
					MATCH (c:Component{ name: $componentName })
					MERGE (s)-[:CONTAINS]->(c)
					`,
					map[string]any{
						"componentName": e.Name,
						"system":        e.System,
					}, neo4j.EagerResultTransformer)
				if err != nil {
					return err
				}
			}
			if len(e.ConsumesAPIs) > 0 {
				for _, consumedAPI := range e.ConsumesAPIs {
					_, err = neo4j.ExecuteQuery(ctx, r.Driver,
						"MERGE (n:Component { name: $name }) RETURN n",
						map[string]any{
							"name": e.Name,
						}, neo4j.EagerResultTransformer)
					if err != nil {
						return err
					}

					_, err = neo4j.ExecuteQuery(ctx, r.Driver,
						"MERGE (a:API { name: $name }) RETURN a",
						map[string]any{
							"name": consumedAPI,
						}, neo4j.EagerResultTransformer)
					if err != nil {
						return err
					}

					_, err = neo4j.ExecuteQuery(ctx, r.Driver,
						`
						MATCH (c:Component{ name: $name})
						MATCH (a:API{ name: $apiName })
						MERGE (c)-[:CONSUMES]->(a)`,
						map[string]any{
							"name":    e.Name,
							"apiName": consumedAPI,
						}, neo4j.EagerResultTransformer)
					if err != nil {
						return err
					}
				}

			}

			if len(e.DependsOn) > 0 {
				for _, dependsOn := range e.DependsOn {
					dependsOnKind, dependsOnName, err := parseDependsOn(dependsOn)
					if err != nil {
						return err
					}

					_, err = neo4j.ExecuteQuery(ctx, r.Driver,
						fmt.Sprintf("MERGE (n:%s { name: $name }) RETURN n", dependsOnKind),
						map[string]any{
							"name": dependsOnName,
						}, neo4j.EagerResultTransformer)
					if err != nil {
						return err
					}

					_, err = neo4j.ExecuteQuery(ctx, r.Driver,
						fmt.Sprintf(`
								MATCH (c:Component{ name: $name})
								MATCH (d:%s{ name: $dependsOnName })
								MERGE (c)-[:DEPENDS_ON]->(d)`, dependsOnKind),
						map[string]any{
							"name":          e.Name,
							"dependsOnName": dependsOnName,
						}, neo4j.EagerResultTransformer)
					if err != nil {
						return err
					}
				}
			}

		case API:
			_, err = neo4j.ExecuteQuery(ctx, r.Driver,
				`
				MERGE (a:API { name: $name })
				ON MATCH
					SET a.system = $system
					SET a.lifecycle = $lifecycle
				RETURN a
				`,
				map[string]any{
					"name":      e.Name,
					"system":    e.System,
					"lifecycle": e.Lifecycle,
				}, neo4j.EagerResultTransformer)
			if err != nil {
				return err
			}
			if e.System != "" {
				_, err = neo4j.ExecuteQuery(ctx, r.Driver,
					"MERGE (n:System { name: $name }) RETURN n",
					map[string]any{
						"name": e.System,
					}, neo4j.EagerResultTransformer)
				if err != nil {
					return err
				}

				_, err = neo4j.ExecuteQuery(ctx, r.Driver,
					"MERGE (a:API { name: $name }) RETURN a",
					map[string]any{
						"name": e.Name,
					}, neo4j.EagerResultTransformer)
				if err != nil {
					return err
				}

				_, err = neo4j.ExecuteQuery(ctx, r.Driver,
					`
					MATCH (s:System{ name: $system})
					MATCH (a:API{ name: $apiName })
					MERGE (s)-[:PROVIDES]->(a)`,
					map[string]any{
						"apiName": e.Name,
						"system":  e.System,
					}, neo4j.EagerResultTransformer)
				if err != nil {
					return err
				}
			}
		}

		if err != nil {
			return err
		}
	}

	// Link API with Systems and Containers
	_, err := neo4j.ExecuteQuery(ctx, r.Driver,
		`
		MATCH (c:Component)-[:CONSUMES]-(a:API)
		MATCH (sourceSystem:System{name: c.system})
		MATCH (targetSystem:System{name: a.system})
		WHERE c.system <> a.system
		CREATE (sourceSystem)-[:DEPENDS_ON{apiName: a.name}]->(targetSystem)
		CREATE (c)-[:DEPENDS_ON{apiName: a.name}]->(targetSystem)
		RETURN c, a, sourceSystem
		`,
		map[string]any{}, neo4j.EagerResultTransformer)
	if err != nil {
		return err
	}

	// Create relations from Components to Systems
	_, err = neo4j.ExecuteQuery(ctx, r.Driver,
		`
		MATCH (sourceContainer:Component)-[:DEPENDS_ON]->(targetContainer:Component)
		MATCH (sourceSystem:System{name: sourceContainer.system})
		MATCH (targetSystem:System{name: targetContainer.system})
		WHERE sourceContainer.system <> targetContainer.system
		MERGE (sourceSystem)-[:DEPENDS_ON]->(targetContainer)
		MERGE (sourceContainer)-[:DEPENDS_ON]->(targetSystem)
		MERGE (sourceSystem)-[:DEPENDS_ON]->(targetSystem)
		RETURN targetContainer, sourceSystem
			`,
		map[string]any{}, neo4j.EagerResultTransformer)
	if err != nil {
		return err
	}

	// Link System and Container with External Systems
	_, err = neo4j.ExecuteQuery(ctx, r.Driver,
		`
		MATCH (sourceContainer:Component)-[:DEPENDS_ON]->(targetSystem:System)
		MATCH (sourceSystem:System{name: sourceContainer.system})
		WHERE sourceContainer.system <> targetSystem.name
		MERGE (sourceSystem)-[:DEPENDS_ON]->(targetSystem)
		RETURN sourceContainer, targetSystem
			`,
		map[string]any{}, neo4j.EagerResultTransformer)
	if err != nil {
		return err
	}
	_, err = neo4j.ExecuteQuery(ctx, r.Driver,
		`
			MATCH (sourceContainer:Component)<-[:DEPENDS_ON]-(targetSystem:System)
			MATCH (sourceSystem:System{name: sourceContainer.system})
			WHERE sourceContainer.system <> targetSystem.name
			MERGE (sourceSystem)<-[:DEPENDS_ON]-(targetSystem)
			RETURN sourceContainer, targetSystem
				`,
		map[string]any{}, neo4j.EagerResultTransformer)

	return err
}

func readSystem(node dbtype.Node) *System {
	system := &System{}
	system.EntityEnvelope = EntityEnvelope{
		ID:          node.ElementId,
		Name:        fmt.Sprintf("%v", node.Props["name"]),
		Title:       fmt.Sprintf("%v", node.Props["title"]),
		Description: fmt.Sprintf("%v", node.Props["description"]),
		Lifecycle:   fmt.Sprintf("%v", node.Props["lifecycle"]),
	}

	if system.Title == "" {
		system.Title = system.Name
	}

	return system
}

func parseDependsOn(dependsOn string) (string, string, error) {
	dependsOnKind := "Component"
	dependsOnName := dependsOn
	if strings.Contains(dependsOn, ":") {
		dependsOnKindRaw := dependsOnName[:strings.Index(dependsOn, ":")]
		_, ok := dependsOnWhitelist[dependsOnKindRaw]
		if !ok {
			return dependsOnKind, dependsOnName, fmt.Errorf("unknown depends on %v", dependsOnKindRaw)
		}
		dependsOnKind = dependsOnWhitelist[dependsOnKindRaw]
		dependsOnName = dependsOnName[strings.Index(dependsOn, ":")+1:]
		return dependsOnKind, dependsOnName, nil
	}

	return dependsOnKind, dependsOnName, nil
}
