package shared

import (
	"context"
	"log"
	"runtime"
)

type contextKey int

var (
	version        = "unknown"
	buildTime      = "unknown"
	buildGoVersion = "unknown"
)

const (
	ContextKeyPrincipal contextKey = 0
	ContextKeyTx        contextKey = 1
)

type RepositoryTxer interface {
	InTx(ctx context.Context, txFuncs ...func(ctxWithTx context.Context) error) error
}

type Version struct {
	Version        string
	BuildTime      string
	BuildGoVersion string
}

func GetVersion() Version {
	return Version{
		Version:        version,
		BuildTime:      buildTime,
		BuildGoVersion: buildGoVersion,
	}
}

func LogVersion() {
	log.Printf("Running version=%s (BuildTime=%s) BuildWith=(%s) RunOn=%s/%s\n",
		version, buildTime, buildGoVersion, runtime.GOOS, runtime.GOARCH)
}
