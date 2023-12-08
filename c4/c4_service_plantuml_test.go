package c4

import (
	"bytes"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestNewPlantUMLExporter(t *testing.T) {
	newPlantUMLExporter()
}

func TestExportToPlantUMLContext(t *testing.T) {
	// Arrange
	is := is.New(t)
	e := newPlantUMLExporter()

	sw := bytes.NewBufferString("")
	m := &C4DiagramModel{}
	m.AddSystem(&System{
		Title: "my-system",
	})

	// Act
	err := e.ExportToPlantUMLContext(m, sw)
	puml := sw.String()

	// Assert
	is.NoErr(err)
	is.True(strings.Contains(puml, "my-system"))
}
