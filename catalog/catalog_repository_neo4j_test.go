package catalog

import (
	"testing"

	"github.com/matryer/is"
)

func TestParseDependsOnSystem(t *testing.T) {
	is := is.New(t)

	dependsOnKind, dependsOnName, err := parseDependsOn("system:my-system")

	is.NoErr(err)
	is.Equal(dependsOnKind, "System")
	is.Equal(dependsOnName, "my-system")
}

func TestParseDependsOnComponent(t *testing.T) {
	is := is.New(t)

	dependsOnKind, dependsOnName, err := parseDependsOn("component:my-component")

	is.NoErr(err)
	is.Equal(dependsOnKind, "Component")
	is.Equal(dependsOnName, "my-component")
}

func TestParseDependsOnAny(t *testing.T) {
	is := is.New(t)

	dependsOnKind, dependsOnName, err := parseDependsOn("my-component")

	is.NoErr(err)
	is.Equal(dependsOnKind, "Component")
	is.Equal(dependsOnName, "my-component")
}

func TestParseDependsOnAPI(t *testing.T) {
	is := is.New(t)

	dependsOnKind, dependsOnName, err := parseDependsOn("api:my-api")

	is.NoErr(err)
	is.Equal(dependsOnKind, "API")
	is.Equal(dependsOnName, "my-api")
}

func TestParseDependsOnInvalid(t *testing.T) {
	is := is.New(t)

	_, _, err := parseDependsOn("invalid:my-api")

	is.True(err != nil)
}
