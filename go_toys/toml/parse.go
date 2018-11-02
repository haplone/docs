package main

import (
	"log"
	"time"
	"github.com/BurntSushi/toml"
)

// https://www.cnblogs.com/CraryPrimitiveMan/p/7928647.html
type tomlConfig struct {
	Title   string
	Owner   ownerInfo
	DB      database `toml:"database"`
	Servers map[string]server
	Clients clients
}

type ownerInfo struct {
	Name string
	Org  string `toml:"organization"`
	Bio  string
	DOB  time.Time
}

type database struct {
	Server  string
	Ports   []int
	ConnMax int `toml:"connection_max"`
	Enabled bool
}

type server struct {
	IP string
	DC string
}

type clients struct {
	Data  [][]interface{}
	Hosts []string
}

func main() {
	parse()
}

func parse() {
	var config tomlConfig
	filePath := "first.toml"
	c, err := toml.DecodeFile(filePath, &config)
	if err != nil {
		panic(err)
	}
	log.Println(c)
	log.Println(config)
	log.Println(config.Servers)
}
