package xconf

import (
	"os"
	"testing"
	"time"

	"github.com/go-viper/mapstructure/v2"

	"github.com/ivy-mobile/odin/encoding/yaml"
)

type Config struct {
	Nacos struct {
		IpAddr string `yaml:"ip_addr" json:"ip_addr"`
		Port   int    `yaml:"port" json:"port"`
	} `yaml:"nacos" json:"nacos"`

	Application struct {
		GameID int           `yaml:"game_id" json:"game_id"`
		State  time.Duration `yaml:"state" json:"state"`
	} `yaml:"application" json:"application"`
}

func TestYaml(t *testing.T) {
	var c Config
	if err := LoadConfigFromFile("./etc.yaml", &c, true); err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", c)
}

func TestJSON(t *testing.T) {
	var c Config
	if err := LoadConfigFromFile("./etc.json", &c, true); err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", c)
}

func TestJsonMapStructure(t *testing.T) {
	bytes, err := os.ReadFile("./etc.yaml")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%v", string(bytes))
	var raw map[string]interface{}
	if err = yaml.Unmarshal(bytes, &raw); err != nil {
		t.Error(err)
		return
	}

	var cfg Config

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result: &cfg,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	})
	if err != nil {
		t.Error(err)
		return
	}
	if err = decoder.Decode(raw); err != nil {
		t.Error(err)
		return
	}
	//if err = mapstructure.Decode(raw, &cfg); err != nil {
	//	t.Error(err)
	//	return
	//}
	t.Logf("%+v", cfg)
}
