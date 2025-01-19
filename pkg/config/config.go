	package config

	import (
	    "fmt"
	    "github.com/spf13/viper"
	    "time"
	)

	type Config struct {
	    Environment string `mapstructure:"ENVIRONMENT"`
	    
	    HTTP struct {
	        Host string `mapstructure:"HTTP_HOST"`
	        Port int    `mapstructure:"HTTP_PORT"`
	    }
	    
	    Database struct {
	        Host     string `mapstructure:"DB_HOST"`
	        Port     int    `mapstructure:"DB_PORT"`
	        User     string `mapstructure:"DB_USER"`
	        Password string `mapstructure:"DB_PASSWORD"`
	        Name     string `mapstructure:"DB_NAME"`
	        SSLMode  string `mapstructure:"DB_SSL_MODE"`
	    }
	    
	    JWT struct {
	        Secret        string        `mapstructure:"JWT_SECRET"`
	        ExpireMinutes time.Duration `mapstructure:"JWT_EXPIRE_MINUTES"`
	    }
	    
	    Metrics struct {
	        Enabled bool   `mapstructure:"METRICS_ENABLED"`
	        Path    string `mapstructure:"METRICS_PATH"`
	    }
	}

	func LoadConfig(path string) (*Config, error) {
	    v := viper.New()
	    
	    v.SetDefault("ENVIRONMENT", "development")
	    v.SetDefault("HTTP_HOST", "0.0.0.0")
	    v.SetDefault("HTTP_PORT", 8080)
	    v.SetDefault("DB_SSL_MODE", "disable")
	    v.SetDefault("JWT_EXPIRE_MINUTES", 60)
	    v.SetDefault("METRICS_ENABLED", true)
	    v.SetDefault("METRICS_PATH", "/metrics")
	    
	    v.SetConfigName("config")
	    v.SetConfigType("yaml")
	    v.AddConfigPath(path)
	    v.AddConfigPath(".")
	    
	    v.AutomaticEnv()
	    
	    if err := v.ReadInConfig(); err != nil {
	        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
	            return nil, fmt.Errorf("error reading config file: %v", err)
	        }
	    }

	    var config Config
	    if err := v.Unmarshal(&config); err != nil {
	        return nil, fmt.Errorf("error unmarshaling config: %v", err)
	    }

	    return &config, nil
	}

