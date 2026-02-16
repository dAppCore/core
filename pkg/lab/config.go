package lab

import (
	"os"
	"strconv"
)

type Config struct {
	Addr string

	PrometheusURL      string
	PrometheusInterval int

	ForgeURL      string
	ForgeToken    string
	ForgeInterval int

	HFAuthor   string
	HFInterval int

	M3Host     string
	M3User     string
	M3SSHKey   string
	M3APIURL   string
	M3Interval int

	TrainingDataDir  string
	TrainingInterval int

	DockerInterval int

	InfluxURL      string
	InfluxToken    string
	InfluxDB       string
	InfluxInterval int
}

func LoadConfig() *Config {
	return &Config{
		Addr: env("ADDR", ":8080"),

		PrometheusURL:      env("PROMETHEUS_URL", "http://prometheus:9090"),
		PrometheusInterval: envInt("PROMETHEUS_INTERVAL", 15),

		ForgeURL:      env("FORGE_URL", "https://forge.lthn.io"),
		ForgeToken:    env("FORGE_TOKEN", ""),
		ForgeInterval: envInt("FORGE_INTERVAL", 60),

		HFAuthor:   env("HF_AUTHOR", "lthn"),
		HFInterval: envInt("HF_INTERVAL", 300),

		M3Host:     env("M3_HOST", "10.69.69.108"),
		M3User:     env("M3_USER", "claude"),
		M3SSHKey:   env("M3_SSH_KEY", "/root/.ssh/id_ed25519"),
		M3APIURL:   env("M3_API_URL", "http://10.69.69.108:9800"),
		M3Interval: envInt("M3_INTERVAL", 30),

		TrainingDataDir:  env("TRAINING_DATA_DIR", "/data/training"),
		TrainingInterval: envInt("TRAINING_INTERVAL", 60),

		DockerInterval: envInt("DOCKER_INTERVAL", 30),

		InfluxURL:      env("INFLUX_URL", "http://localhost:8181"),
		InfluxToken:    env("INFLUX_TOKEN", ""),
		InfluxDB:       env("INFLUX_DB", "training"),
		InfluxInterval: envInt("INFLUX_INTERVAL", 60),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
