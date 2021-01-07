package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"github.com/astaxie/beego/logs"
)

type Cfg struct {
	Agent    *Agent `yaml:"agent"`
	Http     string `yaml:"http"`
	Protocol string `yaml:"protocol"`
	Web      string `yaml:"web"`
}

type Agent struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
}

func init() {
	buf, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		logs.Error("Config read error:", err)
		return
	}
	err = yaml.Unmarshal(buf, &g_cfg)
	if err != nil {
		logs.Error("Config parse error:", err)
		return
	}
}

func Get() *Cfg {
	return g_cfg
}

var g_cfg *Cfg
