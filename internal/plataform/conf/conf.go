package conf

import (
	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

const configFileName = "restql.yml"
const defaultConfigPath = "./internal/plataform/conf/defaults.yaml"

type requestIdConf struct {
	Header   string `yaml:"header"`
	Strategy string `yaml:"strategy"`
}

type timeoutConf struct {
	Duration string `yaml:"duration"`
}

type Config struct {
	Web struct {
		Server struct {
			ApiAddr                 string        `env:"PORT,required"`
			ApiHealthAddr           string        `env:"HEALTH_PORT,required"`
			DebugAddr               string        `env:"DEBUG_PORT"`
			Env                     string        `env:"ENV"`
			GracefulShutdownTimeout time.Duration `yaml:"gracefulShutdownTimeout"`
			ReadTimeout             time.Duration `yaml:"readTimeout"`
			Middlewares             struct {
				RequestId *requestIdConf `yaml:"requestId"`
				Timeout   *timeoutConf   `yaml:"timeout"`
			} `yaml:"middlewares"`
		} `yaml:"server"`

		Client struct {
			ReadTimeout                time.Duration `yaml:"readTimeout"`
			WriteTimeout               time.Duration `yaml:"writeTimeout"`
			MaxConnDuration            time.Duration `yaml:"maxConnectionsDuration"`
			MaxIdleConnDuration        time.Duration `yaml:"maxIdleConnectionDuration"`
			MaxIdleConnectionsPerHosts int           `yaml:"maxIdleConnectionsPerHost"`
		} `yaml:"client"`
	} `yaml:"web"`

	Logging struct {
		Enable    bool   `yaml:"enable"`
		Timestamp bool   `yaml:"timestamp"`
		Level     string `yaml:"level"`
		Format    string `yaml:"format"`
	} `yaml:"logging"`

	Tenant               string        `env:"TENANT"`
	GlobalQueryTimeout   time.Duration `env:"QUERY_GLOBAL_TIMEOUT" envDefault:"30s"`
	QueryResourceTimeout time.Duration `env:"QUERY_RESOURCE_TIMEOUT" envDefault:"5s"`

	Mappings map[string]string `yaml:"mappings"`

	Queries map[string]map[string][]string `yaml:"queries"`

	Env EnvSource

	Build string
}

func Load(build string) (*Config, error) {
	cfg := Config{}
	err := yaml.Unmarshal(readDefaultFile(), &cfg)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(readConfigFile(), &cfg)
	if err != nil {
		return nil, err
	}

	err = env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	cfg.Build = build
	cfg.Env = EnvSource{}

	return &cfg, nil
}

func readDefaultFile() []byte {
	data, err := ioutil.ReadFile(defaultConfigPath)
	if err != nil {
		log.Printf("[ERROR] could not load default file at %s", defaultConfigPath)
		return nil
	}

	return data
}

func readConfigFile() []byte {
	path := getConfigFilepath()
	if path == "" {
		log.Printf("[WARN] no config file present")
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("[ERROR] could not load file at %s", path)
		return nil
	}

	return data
}

func getConfigFilepath() string {
	fileAtRoot := filepath.Join(".", configFileName)
	if doesFileExist(fileAtRoot) {
		return fileAtRoot
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("[DEBUG] failed to find home directory: %v", err)
		return ""
	}

	fileAtHome := filepath.Join(homeDir, configFileName)
	if doesFileExist(fileAtHome) {
		return fileAtHome
	}

	return ""
}

func doesFileExist(filepath string) bool {
	info, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

type EnvSource struct{}

func (e EnvSource) GetString(key string) string {
	return os.Getenv(key)
}
