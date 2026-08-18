package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port     int           `yaml:"port"`
	DataDir  string        `yaml:"data_dir"`
	LogDir   string        `yaml:"log_dir"`
	Database string        `yaml:"database"`
	TLS      TLSConfig     `yaml:"tls"`
	Modules  ModulesConfig `yaml:"modules"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type ModulesConfig struct {
	Minecraft bool `yaml:"minecraft"`
	Docker    bool `yaml:"docker"`
	Apps      bool `yaml:"apps"`
	Websites  bool `yaml:"websites"`
	Databases bool `yaml:"databases"`
}

var C Config

func Default() Config {
	return Config{
		Port:     8080,
		DataDir:  "/var/lib/rockpanel",
		LogDir:   "/var/log/rockpanel",
		Database: "sqlite",
		TLS:      TLSConfig{Enabled: false},
		Modules: ModulesConfig{
			Minecraft: true,
			Docker:    true,
			Apps:      true,
			Websites:  true,
			Databases: true,
		},
	}
}

func Load(path string) error {
	if path == "" {
		etcPath := "/etc/rockpanel/config.yaml"
		homePath := filepath.Join(os.Getenv("HOME"), ".config/rockpanel/config.yaml")
		if _, err := os.Stat(etcPath); err == nil {
			path = etcPath
		} else {
			path = homePath
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			C = Default()
			return nil
		}
		return err
	}
	return yaml.Unmarshal(data, &C)
}

func Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	out, err := yaml.Marshal(&C)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

func EnsureDirs() error {
	dirs := []string{
		C.DataDir,
		C.LogDir,
		filepath.Join(C.DataDir, "servers"),
		filepath.Join(C.DataDir, "apps"),
		filepath.Join(C.DataDir, "minecraft"),
		filepath.Join(C.DataDir, "backups"),
		filepath.Join(C.DataDir, "uploads"),
		filepath.Join(C.DataDir, "websites"),
		filepath.Join(C.DataDir, "webui"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}