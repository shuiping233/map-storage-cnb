package model

import (
	"encoding/json"
	"log"
)

type ServiceConfig struct {
	Host string `default:"0.0.0.0"`
	Port string `default:"8080"`
}

type LocalStorageConfig struct {
	Path string `default:"./uploads"`
}

type GitStorageConfig struct {
	RepoInitFileName    string `default:"Placeholder"`
	RepoInitFileContent string `default:"Placeholder"`
	RemoteGitRepoUrl    string `default:""`
	GitWorkSpaceDir     string `default:"./git_storage_repo"`
	WriteFileWorkers    uint   `default:"4"`
	MaxPushFileAtOnce   uint   `default:"50"`
	FileBatchWindow     uint   `default:"5"`
	CommitAuthor        string `default:"RepoBot"`
	CommitEmail         string `default:"RepoBot@example.com"`
}

type StorageDBConfig struct {
	URL      string `default:"data.db"`
	Username string `default:""`
	Password string `default:""`
}

type StorageConfig struct {
	Type         StorageType        `default:"LocalStorage"`
	DB           StorageDBConfig    `default:""`
	GitStorage   GitStorageConfig   `default:""`
	LocalStorage LocalStorageConfig `default:""`
}

type Config struct {
	Service ServiceConfig `default:""`
	Storage StorageConfig `default:""`
}

// 默认4空格缩进
func (c *Config) Print(indent string) {
	result, _ := c.JsonDump(indent)
	log.Printf("config : %s", result)
}

// 默认4空格缩进
func (c *Config) JsonDump(indent string) (string, error) {
	if indent == "" {
		indent = "    "
	}
	result, err := json.MarshalIndent(c, "", indent)
	return string(result), err
}

// 默认4空格缩进
func (c *Config) JsonDumpBytes(indent string) ([]byte, error) {
	if indent == "" {
		indent = "    "
	}
	result, err := json.MarshalIndent(c, "", indent)
	return result, err
}
