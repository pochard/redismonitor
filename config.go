package main

import (
	"github.com/pochard/zkutils"
	"encoding/json"
)

type cookieOption struct {
	QueryParam       string `json:"query_param"`
	CookieExpireDays int    `json:"cookie_expire_days"`
}

type config struct {
	HostName   string                  `json:"hostName"`
	Debug      bool                    `json:"debug"`
	Port       int                     `json:"port"`
	AutoTest   bool                    `json:"autoTest"`
	CookieName string                  `json:"cookieName"`
	Logdir     string                  `json:"logdir"`
	Partner    map[string]cookieOption `json:"partner"`
}

type redisOptions struct {
	Addrs        []string `json:"Addrs"`
	DialTimeout  int      `json:"DialTimeout"`
	ReadTimeout  int      `json:"ReadTimeout"`
	WriteTimeout int      `json:"WriteTimeout"`
	PoolTimeout  int      `json:"PoolTimeout"`
}

func GetConfig(zkconf *zkutils.ZkConf, k string) (*config, error) {
	var err error
	var val []byte
	val, err = zkconf.Get(k)
	if err != nil {
		return nil, err
	}
	v := config{}
	err = json.Unmarshal(val, &v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func GetRedisOptions(zkconf *zkutils.ZkConf, k string) (*redisOptions, error) {
	var err error
	var val []byte
	val, err = zkconf.Get(k)
	if err != nil {
		return nil, err
	}
	v := redisOptions{}
	err = json.Unmarshal(val, &v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}
