package config

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	configuration *Configuration
	configFileExt = ".yml"
	configType    = "yaml"
)

// Configuration ...
type Configuration struct {
	Server   ServerConfiguration
	Database DatabaseConfiguration
	Email    EmailConfiguration
}

// ServerConfiguration is the required parameters to set up a server
type ServerConfiguration struct {
	Env                        string `default:"dev"` // dev, prod
	Port                       string `default:"3625"`
	Domain                     string `default:"https://vault.passwall.io"`
	Dir                        string `default:"/app/config"`
	Passphrase                 string `default:"passphrase-for-encrypting-passwords-do-not-forget"`
	Secret                     string `default:"secret-key-for-JWT-TOKEN"`
	Timeout                    int    `default:"24"`
	GeneratedPasswordLength    int    `default:"16"`
	AccessTokenExpireDuration  string `default:"30m"`
	RefreshTokenExpireDuration string `default:"15d"`
	APIKey                     string `default:"my-secret-api-key"`
}

// DatabaseConfiguration is the required parameters to set up a DB instance
type DatabaseConfiguration struct {
	Name     string `default:"passwall"`
	Username string `default:"user"`
	Password string `default:"password"`
	Host     string `default:"localhost"`
	Port     string `default:"5432"`
	LogMode  bool   `default:"false"`
	SSLMode  string `default:"disable"`
}

// EmailConfiguration is the required parameters to send emails
type EmailConfiguration struct {
	Host     string `default:"smtp.passwall.io"`
	Port     string `default:"25"`
	Username string `default:"hello@passwall.io"`
	Password string `default:"password"`
	From     string `default:"hello@passwall.io"`
	Admin    string `default:"hello@passwall.io"`
}

// Init initializes the configuration manager
func Init(configPath, configName string) (*Configuration, error) {

	configFilePath := filepath.Join(configPath, configName) + configFileExt

	// initialize viper configuration
	initializeConfig(configPath, configName)

	// Bind environment variables
	bindEnvs()

	// Set default values
	setDefaults()

	// Read or create configuration file
	if err := readConfiguration(configFilePath); err != nil {
		return nil, err
	}

	// Auto read env variables
	viper.AutomaticEnv()

	// Unmarshal config file to struct
	if err := viper.Unmarshal(&configuration); err != nil {
		return nil, err
	}

	return configuration, nil
}

// read configuration from file
func readConfiguration(configFilePath string) error {
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		// if file does not exist, simply create one
		if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
			os.Create(configFilePath)
		} else {
			return err
		}
		// let's write defaults
		if err := viper.WriteConfig(); err != nil {
			return err
		}
	}
	return nil
}

// initialize the configuration manager
func initializeConfig(configPath, configName string) {
	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
}

func bindEnvs() {
	viper.BindEnv("server.env", "PW_ENV")
	viper.BindEnv("server.port", "PORT")
	viper.BindEnv("server.domain", "DOMAIN")
	viper.BindEnv("server.passphrase", "PW_SERVER_PASSPHRASE")
	viper.BindEnv("server.secret", "PW_SERVER_SECRET")
	viper.BindEnv("server.timeout", "PW_SERVER_TIMEOUT")

	viper.BindEnv("server.generatedPasswordLength", "PW_SERVER_GENERATED_PASSWORD_LENGTH")
	viper.BindEnv("server.accessTokenExpireDuration", "PW_SERVER_ACCESS_TOKEN_EXPIRE_DURATION")
	viper.BindEnv("server.refreshTokenExpireDuration", "PW_SERVER_REFRESH_TOKEN_EXPIRE_DURATION")

	viper.BindEnv("server.apiKey", "PW_SERVER_API_KEY")
	viper.BindEnv("server.recaptcha", "PW_SERVER_RECAPTCHA")

	viper.BindEnv("database.name", "PW_DB_NAME")
	viper.BindEnv("database.username", "PW_DB_USERNAME")
	viper.BindEnv("database.password", "PW_DB_PASSWORD")
	viper.BindEnv("database.host", "PW_DB_HOST")
	viper.BindEnv("database.port", "PW_DB_PORT")
	viper.BindEnv("database.logmode", "PW_DB_LOG_MODE")

	// "require", "verify-full", "verify-ca", "disable" supported for postgres
	viper.BindEnv("database.sslmode", "PW_DB_SSL_MODE")

	viper.BindEnv("email.host", "PW_EMAIL_HOST")
	viper.BindEnv("email.port", "PW_EMAIL_PORT")
	viper.BindEnv("email.username", "PW_EMAIL_USERNAME")
	viper.BindEnv("email.password", "PW_EMAIL_PASSWORD")
	viper.BindEnv("email.fromEmail", "PW_EMAIL_FROM_EMAIL")
	viper.BindEnv("email.fromName", "PW_EMAIL_FROM_NAME")
	viper.BindEnv("email.apiKey", "PW_EMAIL_API_KEY")
}

func setDefaults() {

	// Server defaults
	viper.SetDefault("server.env", "prod")
	viper.SetDefault("server.port", "3625")
	viper.SetDefault("server.domain", "https://vault.passwall.io")
	viper.SetDefault("server.passphrase", generateKey())
	viper.SetDefault("server.secret", generateKey())
	viper.SetDefault("server.timeout", 24)
	viper.SetDefault("server.generatedPasswordLength", 16)
	viper.SetDefault("server.accessTokenExpireDuration", "30m")
	viper.SetDefault("server.refreshTokenExpireDuration", "15d")
	viper.SetDefault("server.apiKey", generateKey())
	viper.SetDefault("server.recaptcha", "GoogleRecaptchaSecret")

	// Database defaults
	viper.SetDefault("database.name", "passwall")
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.logmode", false)

	// "require", "verify-full", "verify-ca", "disable" supported for postgres
	viper.SetDefault("database.sslmode", "disable")

	// Email defaults
	viper.SetDefault("email.host", "smtp.passwall.io")
	viper.SetDefault("email.port", "25")
	viper.SetDefault("email.username", "hello@passwall.io")
	viper.SetDefault("email.password", "password")
	viper.SetDefault("email.fromName", "Passwall")
	viper.SetDefault("email.fromEmail", "hello@passwall.io")
	viper.SetDefault("email.apiKey", "apiKey")
}

func generateKey() string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "add-your-key-to-here"
	}
	keyEnc := base64.StdEncoding.EncodeToString(key)
	return keyEnc
}
