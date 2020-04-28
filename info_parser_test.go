package main

import (
	"io/ioutil"
	"testing"
)

func checkstring(t *testing.T, expected string, actual string) {
	if actual != expected {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", expected, actual)
	}
}

func TestHello(t *testing.T) {
	bytes, _ := ioutil.ReadFile("./redisinfo.txt")
	kvmap := ParseRedisInfo(string(bytes))

	checkstring(t, "788.61K", kvmap["used_memory_human"])
	checkstring(t, "239", kvmap["instantaneous_ops_per_sec"])
	checkstring(t, "master", kvmap["role"])
	checkstring(t, "3678", kvmap["db0_keys"])
	checkstring(t, "122", kvmap["db0_expires"])
}

func testNodes(t *testing.T) {
	bytes, _ := ioutil.ReadFile("./redis_nodes.txt")
	nodes := ParseNodes(string(bytes))
	for _,node := range nodes{
		t.Errorf("%v", node)
	}
	t.Errorf("%v", GetClusterMap(nodes))

}

func TestNodes2(t *testing.T) {
	bytes, _ := ioutil.ReadFile("./redis_nodes2.txt")
	nodes := ParseNodes(string(bytes))
	for _,node := range nodes{
		t.Errorf("%v", node)
	}
	t.Errorf("%v", GetClusterMap(nodes))

}