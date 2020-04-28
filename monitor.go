package main

import(
	"fmt"
	"sync"
	"sort"
	"strings"
	"time"
)

type RGMap map[string]RedisGroup

func (m RGMap)Put(key string, info Info){
	if info.Role == "master"{
		group,ok := m[key]
		if ok{
			group.Master = info
			m[key] = group
			//fmt.Printf("k=%v group.Master=%v  group.Slave=%v\n", k,  info, group.Slave)
		}else{
			m[key] = RedisGroup{Master:info}
		}
	}else{
		group,ok := m[info.Master]
		if ok{
			group.Slave = info
			m[info.Master] = group
		}else{
			m[info.Master] = RedisGroup{Slave:info}
		}
	}
}

func (m RGMap)getValues()[]RedisGroup{
	values := make([]RedisGroup,len(m))
	i := 0
	for _,v := range m{
		values[i] = v
		i++
	}
	sort.Slice(values, func(i,j int)bool{
		return (values[i].Master.Id < values[j].Master.Id)
	})
	return values
}

type RedisGroup struct{
	Master Info
	Slave  Info
}

type Info struct{
	Id  string
	Mem string
	Ops string
	Role string
	Master string
	Keys string
	Slots string
	Err  string
}

func (i Info)String()string{
	return strings.Join([]string{"id=", i.Id,",used_memory_human=", i.Mem,",instantaneous_ops_per_sec=",i.Ops,",role=",i.Role,",keys=",i.Keys,",err=",i.Err},"")
}


type Results struct {
  mutex sync.Mutex
  results map[string]Info
}

func NewResults() *Results{
	r := Results{}
	r.results = make(map[string]Info)
	return &r
}

func (r *Results)Set(key string ,value Info){
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.results[key] = value
}

func GetInfo() []RedisGroup{
	var err error
	var infostr string
	var results = NewResults()
	var wg sync.WaitGroup

	totalStart := time.Now()
	for index, client := range gClients{
		wg.Add(1)
		go func(index int, rclient *redisClient){
			start := time.Now()
			fmt.Printf("start: %d %v\n",index, start)
			infostr, err = rclient.Info()
			if err != nil{
				results.Set(rclient.Id, Info{Id: rclient.Id, Err:err.Error()})
			}else{
				kvmap := ParseRedisInfo(infostr)
				info := Info{}
				info.Id = rclient.Id
				info.Mem = kvmap["used_memory_human"]
				info.Ops = kvmap["instantaneous_ops_per_sec"]
				info.Role = kvmap["role"]
				if info.Role == "slave"{
					info.Master = strings.Join([]string{kvmap["master_host"],kvmap["master_port"]},":")
				}
				info.Keys = kvmap["db0_keys"]
				fmt.Printf("%s\n",info.String())
				results.Set(rclient.Id, info)
			}

			fmt.Printf("end: %d %v\n",index, time.Now().Sub(start))
			wg.Add(-1)
		}(index, client)
	}
	wg.Wait()
	fmt.Printf("total end: %v\n", time.Now().Sub(totalStart))

	val, _ := gClient.Do("cluster","nodes").String()
	fmt.Printf("%v\n",val)

	nodeMap := ParseNodes(val)
	clusterMap := GetClusterMap(nodeMap)
	groupMap := make(RGMap)
	for masterId,slaveId := range clusterMap{
		masterInfo := results.results[masterId]
		node, ok := nodeMap[masterId]
		if ok{
			masterInfo.Slots = node.Slots
		}
		slaveInfo := results.results[slaveId]
		groupMap[masterId] = RedisGroup{masterInfo, slaveInfo}
	}

	for _,node := range nodeMap{
		if node.MasterId == "-"{
			_,ok := clusterMap[node.Addr]
			if !ok{
				masterInfo := results.results[node.Addr]
				node, ok := nodeMap[node.Addr]
				if ok{
					masterInfo.Slots = node.Slots
				}
				slaveInfo := Info{Err:"no slave"}
				groupMap[node.Addr] = RedisGroup{masterInfo, slaveInfo}
			}
		}
	}

	return groupMap.getValues()
}

