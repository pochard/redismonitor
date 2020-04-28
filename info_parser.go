package main

import (
	_"fmt"
	"regexp"
	"strings"
	"strconv"
)

//parse the response of redis.info() into a map
func ParseRedisInfo(info string) map[string]string {
	kvmap := make(map[string]string)
	info = strings.Replace(info, "\r", "", -1)
	lines := strings.Split(info, "\n")
	re := regexp.MustCompile(`^db[0-9]*$`)
	for _, line := range lines {
		line = strings.Trim(line, " ")
		if line != "" && !strings.HasPrefix(line, "#") {
			fields := strings.Split(line, ":")
			if len(fields) > 1 {
				kvmap[fields[0]] = fields[1]
				if re.MatchString(fields[0]) {
					subfields := strings.Split(fields[1], ",")
					for _, sf := range subfields {
						kv := strings.Split(sf, "=")
						if len(kv) > 1 {
							kvmap[strings.Join([]string{fields[0], kv[0]}, "_")] = kv[1]
						}
					}
				}
			}
		}
	}

	return kvmap
}

type Node struct{
	Id string
	Addr string
	Cpport int
	Flags []string
	MasterId string
	PingSent int64
	PongRecv int64
	ConfigEpoch int
	LinkState string
	Slots string
}
//parse the response of redis.info() into a map
func ParseNodes(reply string) map[string]Node {
	arr := strings.Split(reply,"\n")
	var nodes = make(map[string]Node)
	for _,line := range arr{
		if line != ""{
		fields := strings.Split(line," ")
		n := Node{}
		n.Id = fields[0]
		subfields := strings.Split(fields[1],"@")
		n.Addr = subfields[0]
		n.Cpport,_ = strconv.Atoi(subfields[1])
		n.Flags = strings.Split(fields[2],",")
		n.MasterId = fields[3]
		n.PingSent,_ = strconv.ParseInt(fields[4], 10, 64)
		n.PongRecv,_ =  strconv.ParseInt(fields[5], 10, 64)
		n.ConfigEpoch,_ =  strconv.Atoi(fields[6])
		n.LinkState =  fields[7]

		 		if len(fields) > 7{
         			n.Slots = strings.Join(fields[8:],",")
         		}
         		nodes[n.Addr] = n
		}

	}
	return nodes
}

func GetClusterMap(nodes map[string]Node)map[string]string{
	m := make(map[string]string)
	r := make(map[string]string)
	for _, node := range nodes{
		m[node.Id] = node.Addr
	}
	for _, node := range nodes{
		if node.MasterId != "-"{
			masterAddr,_ := m[node.MasterId]
			r[masterAddr] = node.Addr
		}
	}
	return r
}
