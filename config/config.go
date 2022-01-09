package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type URL struct {
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
	Path     string `json:"path"`

	Authenticated bool `json:"authenticated"`
}

type Config struct {
	Token   string `json:"token"`
	APIkey  string `json:"apikey"`
	Prefix  string `json:"prefix"`
	EmojiID string `json:"emoji_id"`
}

type Statistics struct {
	ScannedMessages int `json:"scanned_messages"`
	ScannedURLs     int `json:"scanned_urls"`
	UnsafeURLs      int `json:"unsafe_urls"`
	SafeURLs        int `json:"safe_urls"`
}

var statsFile string
var Cache = make(map[string]*URL)

func New(filepath string) (*Config, error) {

	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	config := Config{}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadCache(filepath string) error {

	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &Cache)
	if err != nil {
		return err
	}

	return nil
}

func UpdateCache(filepath, key string, value *URL) {

	Cache[key] = value

	bytes, err := json.MarshalIndent(Cache, "", "    ")
	if err != nil {
		fmt.Println(err)
		return
	}

	_ = ioutil.WriteFile(filepath, bytes, 0666)
}

func NewStatisticsWatcher(filepath string) (*Statistics, error) {

	statsFile = filepath

	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	statistics := Statistics{}
	err = json.Unmarshal(file, &statistics)
	if err != nil {
		return nil, err
	}

	return &statistics, nil
}

func (s *Statistics) Update() {

	bytes, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		fmt.Println(err)
		return
	}

	_ = ioutil.WriteFile(statsFile, bytes, 0666)
}