package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	PublicHost             string
	Port                   string
	DBUser                 string
	DBPassword             string
	DBAddress              string
	DBName                 string
	JWTExpirationInSeconds int64
	JWTSecret              string
	APIkey                 string
}

var Envs = initConfig()

func initConfig() Config {

	// godotenv.Load()
	return Config{
		PublicHost:             getEnv("Public_Host", "127.0.0.1"),
		Port:                   getEnv("Port", "8080"),
		DBUser:                 getEnv("DB_User", "root"),
		DBPassword:             getEnv("DB_Password", "S@ro#i6674"),
		DBAddress:              fmt.Sprintf("%s:%s", getEnv("DB_Host", "127.0.0.1"), getEnv("Port", "3306")),
		DBName:                 getEnv("DB_Name", "vendor"),
		JWTExpirationInSeconds: getEnvAsInt("JWT_Exp", 3600*24*7),
		JWTSecret:              getEnv("JWT_Secret", "Secret is not secret anymore"),
		APIkey:                 getEnv("APIkey", "qMi4HV2NcrowG63SFfs01EkdbBTlRxgUZIJaYumXKePvn58CtLZ7Ixd2zTBCUcwHO8b61oMpEhuXRqGa"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}
