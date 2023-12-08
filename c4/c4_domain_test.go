package c4

import (
	"testing"

	"github.com/matryer/is"
)

func TestC4ModelIsEmptyWithEmptyModel(t *testing.T) {
	is := is.New(t)

	m := C4DiagramModel{}

	is.True(m.IsEmpty())
}

func TestC4ModelIsEmptyWithOneSystem(t *testing.T) {
	is := is.New(t)

	m := C4DiagramModel{}
	m.AddSystem(&System{})

	is.True(!m.IsEmpty())
}

func TestC4ModelIsEmptyWithOneExtSystem(t *testing.T) {
	is := is.New(t)

	m := C4DiagramModel{}
	m.AddExternalSystem(&System{})

	is.True(!m.IsEmpty())
}

func TestC4ModelIsEmptyWithOneContainer(t *testing.T) {
	is := is.New(t)

	m := C4DiagramModel{}
	m.AddContainer(&Container{})

	is.True(!m.IsEmpty())
}

func TestSystemIsExternalWithExternalSystem(t *testing.T) {
	is := is.New(t)

	is.True(System{Type: "external"}.IsExternal())
}

func TestSystemIsExternalWithOtherSystem(t *testing.T) {
	is := is.New(t)

	is.True(!System{Type: "other"}.IsExternal())
}

func TestSpriteWithWhitelistedSprite(t *testing.T) {
	is := is.New(t)

	c := &Container{
		Tags: []string{"java"},
	}

	is.Equal(c.Sprite(), "java")
}

func TestSpriteWithNonWhitelistedSprite(t *testing.T) {
	is := is.New(t)

	c := &Container{
		Tags: []string{"other"},
	}

	is.Equal(c.Sprite(), "")
}

func TestAsTagsWithWhitelistedSprite(t *testing.T) {
	is := is.New(t)

	c := &Container{
		Tags: []string{"java", "deprecated"},
	}

	is.Equal(c.AsTags(), "deprecated")
}

func TestAsTagsWithNonWhitelistedTag(t *testing.T) {
	is := is.New(t)

	c := &Container{
		Tags: []string{"other", "java"},
	}

	is.Equal(c.AsTags(), "")
}
