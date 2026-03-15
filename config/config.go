package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Repository RepositoryConfig `mapstructure:"repository"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type RepositoryConfig struct {
	Type         string `mapstructure:"type"` // "memory" or "mysql"
	SegmentCount int    `mapstructure:"segment_count"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/wallet_service/")

	// Set default values
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("repository.type", "memory")
	viper.SetDefault("repository.segment_count", 64)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.user", "root")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "wallet")

	// Support environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("WALLET")
	viper.BindEnv("server.port", "PORT")
	viper.BindEnv("repository.type")
	viper.BindEnv("repository.segment_count")
	viper.BindEnv("database.host")
	viper.BindEnv("database.port")
	viper.BindEnv("database.user")
	viper.BindEnv("database.password")
	viper.BindEnv("database.dbname")

	if err := viper.ReadInConfig(); err != nil {
		// Ignore config file not found, use defaults
		_, ok := err.(viper.ConfigFileNotFoundError)
		if !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
