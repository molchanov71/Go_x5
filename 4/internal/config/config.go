package config

import "os"

func Address() string {
	return os.Getenv("HTTP_ADDR")
}
