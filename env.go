package main

import (
	"encoding/json"
	"io/ioutil"
)

type env struct {
	Ip string `json:"ip"`
	ZkServer []string `json:"zkServer"`
}

func readLocalEnv(filename string) env {
	byteArr, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	e := env{}
	err = json.Unmarshal(byteArr, &e)
	if err != nil {
		panic(err)
	}
	return e
}
