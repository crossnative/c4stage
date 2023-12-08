package shared

import (
	"testing"

	"github.com/matryer/is"
)

func TestGetVersion(t *testing.T) {
	is := is.New(t)

	t.Run("get version", func(t *testing.T) {
		version := GetVersion()
		is.Equal(version.Version, "unknown")
	})
}

func TestLogVersion(t *testing.T) {

	t.Run("log version", func(t *testing.T) {
		LogVersion()
	})
}
