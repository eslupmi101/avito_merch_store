package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	SecretKey  string
	Env        string `yaml:"env" env-default:"local"`
	HTTPServer `yaml:"http_server"`
	Database   Database `yaml:"database"`
}

type HTTPServer struct {
	Address      string        `yaml:"address" env-default:"localhost:8080"`
	Timeout      time.Duration `yaml:"timeout" env-default:"4s"`
	Idle_timeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Database struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	Name     string `yaml:"name" env-required:"true"`
	SSLMode  string `yaml:"sslmode" env-default:"disable"`
}

var (
	cfgInstance     *Config
	cfgInstanceOnce sync.Once
)

func GetCfgInstance() *Config {
	cfgInstanceOnce.Do(func() {
		cfgInstance = nil
	})
	return cfgInstance
}

func (cfg *Config) BuildPGConnString() (string, error) {
	db := cfg.Database
	if db.User == "" || db.Password == "" || db.Name == "" {
		return "", fmt.Errorf("database user, password, and name must be provided")
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode,
	)
	return connStr, nil
}

func NewConfig() Config {
	var cfg Config

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatalf("SECRET_KEY is not set")
	}

	cfg.SecretKey = secretKey

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatalf("CONFIG_PATH is not set")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}
	absConfigPath := filepath.Join(workingDir, configPath)

	if _, err := os.Stat(absConfigPath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist: %s", absConfigPath)
	}

	cfg.Database = NewDatabase()

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Cannot read config file: %s", err)
	}

	cfgInstance = &cfg

	return cfg
}

func NewDatabase() Database {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("Invalid DB_PORT: %v", err)
	}

	db := Database{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Name:     os.Getenv("POSTGRES_DB"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	if db.Host == "" || db.User == "" || db.Password == "" || db.Name == "" {
		log.Fatalf("Database configuration is incomplete")
	}

	return db
}
