package shared

import (
	"strings"
)

type Config struct {
	BindPort string `envconfig:"PORT" default:"8080"`
	Env      string `default:"dev"`

	Db         string `default:"neo4j://localhost"`
	DbUser     string `default:"neo4j"`
	DbPassword string `default:"c4stage12345!"`

	PlantUMLServer  string `default:"http://localhost:9090"`
	BackstageServer string `default:"http://localhost:7007"`
}

func (c Config) IsProduction() bool {
	return strings.ToLower(c.Env) == "production"
}
