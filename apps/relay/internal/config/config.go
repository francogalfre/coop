package config

import "os"

type Config struct {
	Addr string
}

const defaultAddr = ":8787"

func Load() Config {
	addr := os.Getenv("COOP_RELAY_ADDR")

	if addr == "" {
		addr = defaultAddr
	}

	return Config{Addr: addr}
}
