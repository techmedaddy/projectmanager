package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	minPort         = 1
	maxPort         = 65535
	minJWTSecretLen = 16
	minBcryptCost   = 12
)

// Config holds validated application configuration sourced from environment
// variables.
type Config struct {
	AppPort        int
	DatabaseURL    string
	JWTSecret      string
	JWTExpiryHours int
	BcryptCost     int
	AutoSeed       bool
}

// Load reads required environment variables and validates them for application
// startup.
func Load() (Config, error) {
	var cfg Config
	var validationErrs []error

	appPort, err := requiredInt("APP_PORT")
	if err != nil {
		validationErrs = append(validationErrs, err)
	} else if appPort < minPort || appPort > maxPort {
		validationErrs = append(validationErrs, fmt.Errorf("APP_PORT must be between %d and %d", minPort, maxPort))
	} else {
		cfg.AppPort = appPort
	}

	databaseURL, err := requiredString("DATABASE_URL")
	if err != nil {
		validationErrs = append(validationErrs, err)
	} else if err := validateDatabaseURL(databaseURL); err != nil {
		validationErrs = append(validationErrs, err)
	} else {
		cfg.DatabaseURL = databaseURL
	}

	jwtSecret, err := requiredString("JWT_SECRET")
	if err != nil {
		validationErrs = append(validationErrs, err)
	} else if len(jwtSecret) < minJWTSecretLen {
		validationErrs = append(validationErrs, fmt.Errorf("JWT_SECRET must be at least %d characters long", minJWTSecretLen))
	} else {
		cfg.JWTSecret = jwtSecret
	}

	jwtExpiryHours, err := requiredInt("JWT_EXPIRY_HOURS")
	if err != nil {
		validationErrs = append(validationErrs, err)
	} else if jwtExpiryHours <= 0 {
		validationErrs = append(validationErrs, fmt.Errorf("JWT_EXPIRY_HOURS must be greater than 0"))
	} else {
		cfg.JWTExpiryHours = jwtExpiryHours
	}

	bcryptCost, err := requiredInt("BCRYPT_COST")
	if err != nil {
		validationErrs = append(validationErrs, err)
	} else if bcryptCost < minBcryptCost {
		validationErrs = append(validationErrs, fmt.Errorf("BCRYPT_COST must be at least %d", minBcryptCost))
	} else {
		cfg.BcryptCost = bcryptCost
	}

	autoSeed, err := optionalBool("AUTO_SEED", true)
	if err != nil {
		validationErrs = append(validationErrs, err)
	} else {
		cfg.AutoSeed = autoSeed
	}

	if len(validationErrs) > 0 {
		return Config{}, errors.Join(validationErrs...)
	}

	return cfg, nil
}

func requiredString(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required", key)
	}

	return strings.TrimSpace(value), nil
}

func requiredInt(key string) (int, error) {
	rawValue, err := requiredString(key)
	if err != nil {
		return 0, err
	}

	value, convErr := strconv.Atoi(rawValue)
	if convErr != nil {
		return 0, fmt.Errorf("%s must be a valid integer", key)
	}

	return value, nil
}

func optionalBool(key string, defaultValue bool) (bool, error) {
	rawValue, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(rawValue) == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(rawValue))
	if err != nil {
		return false, fmt.Errorf("%s must be a valid boolean", key)
	}

	return parsed, nil
}

func validateDatabaseURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("DATABASE_URL must be a valid URL")
	}

	if parsed.Scheme == "" {
		return fmt.Errorf("DATABASE_URL must include a scheme")
	}

	if parsed.Host == "" {
		return fmt.Errorf("DATABASE_URL must include a host")
	}

	return nil
}
