package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var SvcConfig = &Config{}

type Config struct {
	DB         DatabaseConfig   `yaml:"db"`
	ClickHouse ClickHouseConfig `yaml:"clickhouse"`
	Solana     SolanaConfig     `yaml:"solana"`
	RpcCall    RpcCallConfig    `yaml:"rpc_call"`
	Env        string           `yaml:"env"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DbName   string `yaml:"db_name"`
}

type ClickHouseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DbName   string `yaml:"db_name"`
}

type SolanaConfig struct {
	RpcUrl string `yaml:"rpc_url"`
}

type RpcCallConfig struct {
	Url string `yaml:"url"`
}

func LoadSvcConfig() error {
	cf, err := os.Open("./config-yaml/config.yaml")
	if err != nil {
		fmt.Println("failed to open config file", err)
		os.Exit(1)
	}

	decoderCf := yaml.NewDecoder(cf)
	if err = decoderCf.Decode(&SvcConfig); err != nil {
		return err
	}

	fmt.Printf("svcConfig: %+v\n", SvcConfig)
	return nil
}

func LoadSvcConfigFromPath() error {
	cf, err := os.Open("/Users/a11111/Desktop/code/golang/smartx/solana-dex-parser-golang/config-yaml/config.yaml")
	if err != nil {
		fmt.Println("failed to open config file", err)
		os.Exit(1)
	}

	decoderCf := yaml.NewDecoder(cf)
	if err = decoderCf.Decode(&SvcConfig); err != nil {
		return err
	}

	fmt.Printf("svcConfig: %+v\n", SvcConfig)
	return nil
}
