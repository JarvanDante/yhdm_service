package config

import (
	"github.com/spf13/viper"
)

// Config 是服务的全部配置。
type Config struct {
	App   AppConfig   `mapstructure:"app"`
	Mongo MongoConfig `mapstructure:"mongo"`
	Redis RedisConfig `mapstructure:"redis"`
	JWT   JWTConfig   `mapstructure:"jwt"`
	Log   LogConfig   `mapstructure:"log"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Dev  bool   `mapstructure:"dev"`
	Addr string `mapstructure:"addr"`
}

type MongoConfig struct {
	URI       string `mapstructure:"uri"`
	Database  string `mapstructure:"database"`
	TimeoutMs int    `mapstructure:"timeoutMs"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	ExpireSeconds int64  `mapstructure:"expireSeconds"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load 从指定 yaml 文件读取配置。环境变量以 YHDM_ 为前缀覆盖，
// 例如 YHDM_MONGO_URI、YHDM_JWT_SECRET。
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("YHDM")
	v.AutomaticEnv()

	// 合理默认值
	v.SetDefault("app.addr", ":8100")
	v.SetDefault("mongo.uri", "mongodb://127.0.0.1:27017")
	v.SetDefault("mongo.database", "yhdm")
	v.SetDefault("mongo.timeoutMs", 5000)
	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("jwt.expireSeconds", 36000)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
