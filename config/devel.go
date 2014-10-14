// +build !heroku

package config

import (
	"fmt"
	"github.com/go-martini/martini"
	"os"
)

func IsHeroku() bool {
	return false
}

func PostgresArgs() string {
	return "dbname=postgres sslmode=disable"
}

func Url() string {
	host, port := os.Getenv("HOST"), os.Getenv("PORT")
	if len(host) <= 0 {
		host = "localhost"
	}
	if len(port) <= 0 {
		port = "3000"
	}
	return fmt.Sprintf("%s:%s", host, port)
}

func Initialize(m *martini.ClassicMartini) {
	fmt.Println("Running in debug environment")
}
