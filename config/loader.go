package config

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type content struct {
	// 用于从go-cqhttp接收消息
	HttpServerConfig struct {
		IsEnable bool   `yaml:"isEnable" json:"isEnable"`
		Address  string `yaml:"address" json:"address"`
		Port     int    `yaml:"port" json:"port"`
	} `yaml:"http" json:"http"`

	// 用于发送消息
	BotServerConfig struct {
		Address string `yaml:"address" json:"address"`
		Port    int    `yaml:"port" json:"port"`
		QQ      int    `yaml:"qq"`
	} `yaml:"bot"`
}

// Content 配置文件内容
var Content = new(content)

// LoadConfig 加载配置文件
func LoadConfig() {
	g.Cfg().GetAdapter().(*gcfg.AdapterFile).SetPath(fmt.Sprintf("../"))
	g.Cfg().GetAdapter().(*gcfg.AdapterFile).SetFileName("config.yaml")

	configFile, err := os.ReadFile("../config.yaml")
	if err != nil {
		log.Fatal(fmt.Sprintf("ERROR: failed to load config, err(%s).", err))
	}

	err = yaml.Unmarshal(configFile, Content)
	if err != nil {
		log.Fatal(fmt.Sprintf("ERROR: failed to unmarshal config, err(%s).", err))
	}
}
