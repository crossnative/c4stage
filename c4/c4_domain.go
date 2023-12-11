package c4

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

type C4DiagramModel struct {
	Systems         []*System
	ExternalSystems []*System
	Containers      []*Container
	Relations       []Relation
	Persons         []*System
	Scale           float64
}

type System struct {
	ID          string
	Label       string
	Title       string
	Description string
	Type        string
	Technology  string
	Containers  []*Container
	Tags        []string
}

type Container struct {
	ID          string
	Label       string
	Title       string
	Description string
	Technology  string
	Type        string
	System      string
	Tags        []string
}

type Relation struct {
	SourceID    string
	TargetID    string
	Label       string
	Title       string
	Description string
	Technology  string
}

type C4Repository interface {
	ContainerDiagram(
		ctx context.Context,
		name string,
	) (*C4DiagramModel, error)

	SystemLandscapeContainerDiagram(
		ctx context.Context,
	) (*C4DiagramModel, error)

	SystemLandscapeDiagram(
		ctx context.Context,
	) (*C4DiagramModel, error)
}

var spriteWhitelist = map[string]string{
	"spring":  "spring",
	"c":       "c",
	"java":    "java",
	"angular": "angular",
	"oracle":  "oracle_original",
}
var tagWhitelist = []string{"deprecated", "experimental"}

func (m C4DiagramModel) IsEmpty() bool {
	return len(m.Systems) == 0 && len(m.Containers) == 0 && len(m.ExternalSystems) == 0
}

func (s System) IsExternal() bool {
	return s.Type == "external"
}

func (s System) IsPerson() bool {
	return s.Type == "person"
}

func (s *System) AddTag(toAdd string) {
	s.Tags = append(s.Tags, toAdd)
}

func (m *C4DiagramModel) ScaleFormatted() string {
	if m.Scale == 0 {
		return "1.0"
	}
	return fmt.Sprintf("%.2f", m.Scale)
}

func (s System) AsTags() string {
	var tags []string
	for _, tag := range s.Tags {
		if !slices.Contains(tagWhitelist, tag) {
			continue
		}
		tags = append(tags, tag)
	}
	return strings.Join(tags, ",")
}

func (c *Container) AddTag(toAdd string) {
	c.Tags = append(c.Tags, toAdd)
}

func (c Container) IsDatabase() bool {
	return c.Type == "database"
}

func (c Container) IsPerson() bool {
	return c.Type == "person"
}

func (c4Model *C4DiagramModel) PostProcess() {
	// Add all Containers to their System
	for _, system := range c4Model.Systems {
		for _, container := range c4Model.Containers {
			if container.System == system.Label {
				system.Containers = append(system.Containers, container)
			}
		}
	}
}

func (c4Model *C4DiagramModel) AddRelation(toAdd Relation) {
	for _, relation := range c4Model.Relations {
		if relation.SourceID == toAdd.SourceID && relation.TargetID == toAdd.TargetID {
			return
		}
	}
	c4Model.Relations = append(c4Model.Relations, toAdd)
}

func (c4Model *C4DiagramModel) AddSystem(toAdd *System) {
	if toAdd.IsPerson() {
		c4Model.AddPerson(toAdd)
		return
	}
	if toAdd.IsExternal() {
		c4Model.AddExternalSystem(toAdd)
		return
	}

	for _, system := range c4Model.Systems {
		if system.ID == toAdd.ID {
			return
		}
	}
	c4Model.Systems = append(c4Model.Systems, toAdd)
}

func (c4Model *C4DiagramModel) AddPerson(toAdd *System) {
	for _, system := range c4Model.Persons {
		if system.ID == toAdd.ID {
			return
		}
	}
	c4Model.Persons = append(c4Model.Persons, toAdd)
}

func (c4Model *C4DiagramModel) AddExternalSystem(toAdd *System) {
	if toAdd.IsPerson() {
		c4Model.AddPerson(toAdd)
		return
	}

	for _, system := range c4Model.ExternalSystems {
		if system.ID == toAdd.ID {
			return
		}
	}
	c4Model.ExternalSystems = append(c4Model.ExternalSystems, toAdd)
}

func (c4Model *C4DiagramModel) AddContainer(toAdd *Container) {
	for _, container := range c4Model.Containers {
		if container.ID == toAdd.ID {
			return
		}
	}
	c4Model.Containers = append(c4Model.Containers, toAdd)
}

func (c Container) Sprite() string {
	for _, tag := range c.Tags {
		spriteName, ok := spriteWhitelist[tag]
		if ok {
			return spriteName
		}
	}
	return ""
}

func (c Container) AsTags() string {
	var tags []string
	for _, tag := range c.Tags {
		if !slices.Contains(tagWhitelist, tag) {
			continue
		}
		tags = append(tags, tag)
	}
	return strings.Join(tags, ",")
}

func AsID(elementID string) string {
	res := strings.ReplaceAll(elementID, ":", "")
	res = strings.ReplaceAll(res, "-", "")
	return res
}

func AsRelation(relation string) string {
	switch relation {
	case "DEPENDS_ON":
		return "depends on"
	case "CONTAINS":
		return "contains"
	default:
		return relation
	}
}
