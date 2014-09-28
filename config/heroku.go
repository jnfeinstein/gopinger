// +build heroku

package config

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/gorelic"
	"os"
)

const APPNAME = "GOPINGER"

func IsHeroku() bool {
	return true
}

func Url() string {
	return os.Getenv("URL")
}

func Initialize(m *martini.ClassicMartini) {
	fmt.Println("Running in production environment")

	NEW_RELIC_LICENSE_KEY := os.Getenv("NEW_RELIC_LICENSE_KEY")
	if len(NEW_RELIC_LICENSE_KEY) > 0 {
		gorelic.InitNewrelicAgent(NEW_RELIC_LICENSE_KEY, APPNAME, true)
		m.Use(gorelic.Handler)
	}
}
