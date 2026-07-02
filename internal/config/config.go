package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
	Posts    PostsConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	URL string
}

type KafkaConfig struct {
	Brokers []string
}

type PostsConfig struct {
	ServiceURL string
}

func Load() (*Config, error) {
	viper.SetDefault("SERVER_PORT", "8084")
	viper.SetDefault("KAFKA_BROKERS", "localhost:29092")
	viper.SetDefault("POST_SERVICE_URL", "http://localhost:8082")

	viper.AutomaticEnv()

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
		},
		Database: DatabaseConfig{
			URL: viper.GetString("DATABASE_URL"),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(viper.GetString("KAFKA_BROKERS"), ","),
		},
		Posts: PostsConfig{
			ServiceURL: viper.GetString("POST_SERVICE_URL"),
		},
	}

	if cfg.Database.URL == "" {
		cfg.Database.URL = "postgres://snaply:snaply_secret@localhost:5432/likes?sslmode=disable"
	}

	return cfg, nil
}
