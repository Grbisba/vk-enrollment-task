package config

type Controller struct {
	Host           string `config:"host" json:"host" yaml:"host" toml:"host"`
	Port           int    `config:"port" json:"port" yaml:"port" toml:"port"`
	TimeoutSeconds int    `config:"timeout-seconds" json:"timeout_seconds" yaml:"timeout_seconds" toml:"timeout_seconds"`
}
