package backstage

import (
	"fmt"

	"github.io/remast/c4stage/catalog"
)

type RawEntity struct {
	APIVersion string      `json:"apiVersion" yaml:"apiVersion"`
	Kind       string      `json:"kind" yaml:"kind"`
	Namespace  string      `json:"namespace" yaml:"namespace"`
	Metadata   RawMetadata `json:"metadata" yaml:"metadata"`
	Spec       RawSpec     `json:"spec" yaml:"spec"`
	Links      []Link      `json:"links" yaml:"links"`
}

type RawMetadata struct {
	Name        string   `json:"name" yaml:"name"`
	Title       string   `json:"title" yaml:"title"`
	Description string   `json:"description" yaml:"description"`
	Owner       string   `json:"owner" yaml:"owner"`
	Domain      string   `json:"domain" yaml:"domain"`
	Tags        []string `json:"tags" yaml:"tags"`
}

type RawSpec struct {
	Type         string   `json:"type" yaml:"type"`
	System       string   `json:"system" yaml:"system"`
	ConsumesAPIs []string `json:"consumesApis" yaml:"consumesApis"`
	ProvidesAPIs []string `json:"providesApis" yaml:"providesApis"`
	Definition   string   `json:"definition" yaml:"definition"`
	DependsOn    []string `json:"dependsOn" yaml:"dependsOn"`
	Lifecycle    string   `json:"lifecycle" yaml:"lifecycle"`
}

type Link struct {
}

func (rawEntity RawEntity) EntityEnvelopeFromRaw() catalog.EntityEnvelope {
	envelope := catalog.EntityEnvelope{
		Name:        rawEntity.Metadata.Name,
		Title:       rawEntity.Metadata.Title,
		Description: rawEntity.Metadata.Description,
		Kind:        rawEntity.Kind,
		Lifecycle:   rawEntity.Spec.Lifecycle,
		Tags:        rawEntity.Metadata.Tags,
		Type:        rawEntity.Spec.Type,
	}
	return envelope
}

func (rawEntity RawEntity) FromRaw() (any, error) {
	var entity any
	envelope := rawEntity.EntityEnvelopeFromRaw()
	switch rawEntity.Kind {
	case "System":
		system := catalog.System{
			EntityEnvelope: envelope,
			DependsOn:      rawEntity.Spec.DependsOn,
		}
		entity = system
	case "Component":
		if rawEntity.Spec.Type != "service" &&
			rawEntity.Spec.Type != "database" {
			return nil, fmt.Errorf("could not convert component of type %s", rawEntity.Spec.Type)
		}
		container := catalog.Container{
			EntityEnvelope: envelope,
			ConsumesAPIs:   rawEntity.Spec.ConsumesAPIs,
			DependsOn:      rawEntity.Spec.DependsOn,
			System:         rawEntity.Spec.System,
		}
		entity = container
	case "API":
		api := catalog.API{
			EntityEnvelope: envelope,
			System:         rawEntity.Spec.System,
		}
		entity = api
	default:
		return nil, fmt.Errorf("could not convert kind %s", rawEntity.Kind)
	}

	return entity, nil
}
