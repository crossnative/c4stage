package c4

import (
	"context"
	"fmt"
	"slices"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
)

var _ C4Repository = (*C4EntityNeo4j)(nil)

type C4EntityNeo4j struct {
	Driver neo4j.DriverWithContext
}

func (r *C4EntityNeo4j) ContainerDiagram(
	ctx context.Context,
	name string,
) (*C4DiagramModel, error) {
	result, err := neo4j.ExecuteQuery(ctx, r.Driver,
		`
		MATCH (s:System{name: $name})-[r:CONTAINS]-(c:Component)
		OPTIONAL MATCH (s)-[:CONTAINS]-(d:Component)-[otherSystemDep:DEPENDS_ON]-(otherSystem:System)
		OPTIONAL MATCH (d)-[extSystemDep:DEPENDS_ON]-(extSystem:System{type:"external"}) 
		OPTIONAL MATCH (c)-[personDep:DEPENDS_ON]-(person:System{type:"person"}) 
		OPTIONAL MATCH (c)-[containerDep:DEPENDS_ON]-(e:Component{system: $name})
		RETURN s,r,c,d,containerDep,e,otherSystemDep,otherSystem,extSystemDep,extSystem,personDep,person
		`,
		map[string]any{
			"name": name,
		}, neo4j.EagerResultTransformer)
	if err != nil {
		return nil, err
	}

	c4Model := &C4DiagramModel{}

	for _, record := range result.Records {
		for _, recordValue := range record.Values {
			switch node := recordValue.(type) {
			case dbtype.Relationship:
				relation := readRelation(node)
				if node.Type != "CONTAINS" {
					c4Model.AddRelation(relation)
				}
			case dbtype.Node:
				if slices.Contains(node.Labels, "System") {
					system := readSystem(node)

					if name == system.Label {
						c4Model.AddSystem(system)
					} else {
						c4Model.AddExternalSystem(system)
					}
				} else if slices.Contains(node.Labels, "Component") {
					container := readContainer(node)
					c4Model.AddContainer(container)
				}
			}
		}
	}

	c4Model.PostProcess()

	return c4Model, nil
}

func (r *C4EntityNeo4j) SystemLandscapeContainerDiagram(
	ctx context.Context,
) (*C4DiagramModel, error) {
	result, err := neo4j.ExecuteQuery(ctx, r.Driver,
		`		
		MATCH (e:Component)-[r:DEPENDS_ON]-(m:Component)
		MATCH (c1:Component)-[l:CONTAINS]-(s1:System)
		OPTIONAL MATCH (c1:Component)-[extSystemDep:DEPENDS_ON]-(extSystem:System{type:"external"})
		OPTIONAL MATCH (c1:Component)-[personDep:DEPENDS_ON]-(person:System{type:"person"})
		RETURN e,r,m,c1,l,s1,extSystemDep,extSystem,personDep,person
		LIMIT 10000
		`,
		map[string]any{}, neo4j.EagerResultTransformer)

	if err != nil {
		return nil, err
	}

	c4Model := &C4DiagramModel{}

	for _, record := range result.Records {
		for _, recordValue := range record.Values {
			switch node := recordValue.(type) {
			case dbtype.Relationship:
				relation := readRelation(node)
				if node.Type != "CONTAINS" {
					c4Model.AddRelation(relation)
				}
			case dbtype.Node:
				if slices.Contains(node.Labels, "System") {
					system := readSystem(node)
					c4Model.AddSystem(system)
				} else if slices.Contains(node.Labels, "Component") {
					container := readContainer(node)
					c4Model.AddContainer(container)
				}
			}
		}
	}

	c4Model.PostProcess()

	return c4Model, nil
}

func (r *C4EntityNeo4j) SystemLandscapeDiagram(
	ctx context.Context,
) (*C4DiagramModel, error) {
	result, err := neo4j.ExecuteQuery(ctx, r.Driver,
		`
		MATCH (s:System) 
		OPTIONAL MATCH (:System)-[r]-(:System)
		RETURN s,r LIMIT 3000
		`,
		map[string]any{}, neo4j.EagerResultTransformer)
	if err != nil {
		return nil, err
	}

	c4Model := &C4DiagramModel{}

	for _, record := range result.Records {
		for _, recordValue := range record.Values {
			switch node := recordValue.(type) {
			case dbtype.Relationship:
				relation := readRelation(node)
				c4Model.AddRelation(relation)
			case dbtype.Node:
				if slices.Contains(node.Labels, "System") {
					system := readSystem(node)
					c4Model.AddSystem(system)
				}
			}
		}
	}

	return c4Model, nil
}

func readSystem(node dbtype.Node) *System {
	system := &System{
		ID:          AsID(node.ElementId),
		Label:       fmt.Sprintf("%v", node.Props["name"]),
		Title:       fmt.Sprintf("%v", node.Props["title"]),
		Description: fmt.Sprintf("%v", node.Props["description"]),
		Type:        fmt.Sprintf("%v", node.Props["type"]),
	}

	if system.Title == "" {
		system.Title = system.Label
	}

	lifecycle := fmt.Sprintf("%v", node.Props["lifecycle"])
	system.AddTag(lifecycle)

	return system
}

func readContainer(node dbtype.Node) *Container {
	container := &Container{
		ID:          AsID(node.ElementId),
		Type:        fmt.Sprintf("%v", node.Props["type"]),
		Label:       fmt.Sprintf("%v", node.Props["name"]),
		Title:       fmt.Sprintf("%v", node.Props["title"]),
		Description: fmt.Sprintf("%v", node.Props["description"]),
		System:      fmt.Sprintf("%v", node.Props["system"]),
	}

	if node.Props["tags"] != nil {
		tagsRaw := node.Props["tags"].([]any)

		var tags []string
		for _, tagRaw := range tagsRaw {
			tags = append(tags, fmt.Sprintf("%v", tagRaw))
		}

		container.Tags = tags
	}

	if container.Title == "" {
		container.Title = container.Label
	}

	lifecycle := fmt.Sprintf("%v", node.Props["lifecycle"])
	container.AddTag(lifecycle)

	return container
}

func readRelation(node dbtype.Relationship) Relation {
	relation := Relation{
		SourceID: AsID(node.StartElementId),
		TargetID: AsID(node.EndElementId),
		Label:    AsRelation(node.Type),
	}
	return relation
}
