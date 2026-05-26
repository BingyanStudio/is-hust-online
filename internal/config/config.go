package config

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/spf13/viper"
)

type CaptchaConfig struct {
	CaptchaId  string `mapstructure:"captcha_id" json:"captcha_id" yaml:"captcha_id"`
	CaptchaKey string `mapstructure:"captcha_key" json:"captcha_key" yaml:"captcha_key"`
	ApiServer  string `mapstructure:"api_server" json:"api_server" yaml:"api_server"`
}

type MongoConfig struct {
	URI      string `mapstructure:"uri" json:"uri" yaml:"uri"`
	Database string `mapstructure:"database" json:"database" yaml:"database"`
}

type GRPCConfig struct {
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	CertFile string `mapstructure:"cert_file" json:"cert_file" yaml:"cert_file"`
	KeyFile  string `mapstructure:"key_file" json:"key_file" yaml:"key_file"`
	Insecure bool   `mapstructure:"insecure" json:"insecure" yaml:"insecure"`
}

type AppConfig struct {
	Captcha           CaptchaConfig `mapstructure:"captcha" json:"captcha" yaml:"captcha"`
	Mongo             MongoConfig   `mapstructure:"mongo" json:"mongo" yaml:"mongo"`
	Debug             bool          `mapstructure:"debug" json:"debug" yaml:"debug"`
	Port              int           `mapstructure:"port" json:"port" yaml:"port"`
	GRPC              GRPCConfig    `mapstructure:"grpc" json:"grpc" yaml:"grpc"`
	BasicAuthUser     string        `mapstructure:"basic_auth_user" json:"basic_auth_user" yaml:"basic_auth_user"`
	BasicAuthPassword string        `mapstructure:"basic_auth_password" json:"basic_auth_password" yaml:"basic_auth_password"`
}

var C AppConfig

func LoadConfig() error {
	v := viper.New()
	if os.Getenv("POD_NAME") != "" {
		url := fmt.Sprintf("http://cc-server.config-center/%s/config/raw", os.Getenv("APP_NAME"))
		data, err := readConfigCenter(url)
		if err != nil {
			return fmt.Errorf("failed to read config from config center: %w", err)
		}
		v.SetConfigType("yaml")
		if err := v.ReadConfig(bytes.NewBuffer(data)); err != nil {
			return fmt.Errorf("failed to parse config from config center: %w", err)
		}
	} else {
		slog.Info("本地调试")
		v.SetConfigFile("config/default.yaml")
		v.SetConfigType("yaml")
		if err := v.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := v.Unmarshal(&C); err != nil {
		return err
	}
	return nil
}

func readConfigCenter(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic("Get config error: " + resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	return data, err
}
