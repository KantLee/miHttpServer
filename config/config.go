package config

// 全局变量，用于存储解析后的配置信息
var Configs Config

type Config struct {
	Redis      RedisConfig      `yaml:"redis"`
	MySQL      MysqlConfig      `yaml:"mysql"`
	Server     ServerConfig     `yaml:"server"`
	Code       CodeConfig       `yaml:"code"`
	Lock       LockConfig       `yaml:"lock"`
	LocalCache LocalCacheConfig `yaml:"localCache"`
}

// redis的配置项
type RedisConfig struct {
	Address      string `yaml:"address"`
	Protocal     string `yaml:"protocal"`
	Password     string `yaml:"password"`
	Database     int    `yaml:"database"`
	Prefix       string `yaml:"prefix"`
	MaxIdle      int    `yaml:"maxIdle"`
	MaxActive    int    `yaml:"maxActive"`
	Expire       int    `yaml:"expire"`
	Timeout      int    `yaml:"timeout"`
	ReadTimeout  int    `yaml:"readTimeout"`
	WriteTimeout int    `yaml:"writeTimeout"`
}

// mysql的配置项
type MysqlConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
}

// http服务的配置项
type ServerConfig struct {
	Port int `yaml:"port"`
}

// Code状态码的配置项
type CodeConfig struct {
	Success      int `yaml:"success"`
	RequestError int `yaml:"requestError"`
	ServerError  int `yaml:"serverError"`
}

// 分布式锁相关的配置项
type LockConfig struct {
	ExpireSec uint64 `yaml:"expireSec"`
	WaitSec   int    `yaml:"waitSec"`
}

// 本地缓存配置项
type LocalCacheConfig struct {
	Capacity  int `yaml:"capacity"`
	ExpireSec int `yaml:"expireSec"`
}
