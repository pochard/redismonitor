package main

import(
	"context"
	"encoding/json"
	_"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/wonderivan/logger"
	"github.com/pochard/zkutils"
	"github.com/samuel/go-zookeeper/zk"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const ZK_CONFIG_DIR = "/hades/configs/main/cms"

var gEnv env
var gRedisOptions *redisOptions

var gClients []*redisClient
var gClient *redis.ClusterClient

type IpAddr struct {
	IP   string
	Port int
}

func (ia *IpAddr) String() string {
	return ia.IP + ":" + strconv.Itoa(ia.Port)
}


func apiHandler(w http.ResponseWriter, r *http.Request) {
	results := GetInfo()
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(200)

	b, err_json := json.MarshalIndent(results,"","  ")
	if err_json != nil {
		logger.Error("%v\n", err_json.Error())
		_, err := w.Write([]byte{})
		if err != nil {
			logger.Error("%v\n", err.Error())
		}
		return
	}
	_, err := w.Write(b)
	if err != nil {
		logger.Error("%v\n", err.Error())
	}
}

func initZkConf(conn *zk.Conn){
	var err error
	var zkconf *zkutils.ZkConf
	zkconf, err = zkutils.NewZkConf(conn, ZK_CONFIG_DIR)
	handleInitErr(err)
	gRedisOptions, err = GetRedisOptions(zkconf, "redis")
	handleInitErr(err)
}

func handleExitSignal(srv *http.Server, idleConnsClosed chan struct{}){
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	data := <-sigint
	logger.Info("received signal: " + data.String())
	logger.Info("start to shutdown...")

	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Error("HTTP server Shutdown: %v", err)
	}
	close(idleConnsClosed)
}

func handleInitErr(err error) {
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
}

func load(tplname string) *template.Template{
 	return template.Must(template.ParseFiles("./template/" + tplname))
}
func createRedisClient(opt *redisOptions) *redis.ClusterClient {
	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"127.0.0.1:7000","127.0.0.1:7001","127.0.0.1:7002","127.0.0.1:7003","127.0.0.1:7004","127.0.0.1:7005","127.0.0.1:7006","127.0.0.1:7007"},
	})
}
func main(){
	var err error
	var conn *zk.Conn

	gEnv = readLocalEnv("./conf/env.json")
	//zk initiation
	conn, _, err = zk.Connect(gEnv.ZkServer,
		time.Second*5,
		zk.WithLogInfo(false),
	)
	handleInitErr(err)
	defer conn.Close()

	initZkConf(conn)
	gClient = createRedisClient(gRedisOptions)
	defer gClient.Close()

	val, _ := gClient.Do("cluster","nodes").String()
	fmt.Printf("%v\n",val)
	nodes := ParseNodes(val)
	var addrs []string
	for _,node := range nodes{
		addrs = append(addrs,node.Addr)
	}
	fmt.Printf("%v\n",addrs)
	gClients = make([]*redisClient,len(addrs))
	for index, addr := range addrs{
		rclient := redis.NewClient(&redis.Options{
			Addr: addr,
        })
        gClients[index] = &redisClient{addr,rclient}
        defer rclient.Close()
	}

	fs := http.FileServer(http.Dir("resources/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

//	tmpl := load("dashboard.tpl")
//	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
//		results := GetInfo()
//		tmpl.Execute(w, results)
//	})
	//start http server
	ipAddr := &IpAddr{"", 8888}
	srv := http.Server{
		Addr:    ipAddr.String(),
	}
	http.HandleFunc("/api", apiHandler)

	logger.Info("listen on " + ipAddr.String())

	idleConnsClosed := make(chan struct{})

	go handleExitSignal(&srv, idleConnsClosed)

	err_http := srv.ListenAndServe()

	if err_http != nil && err_http != http.ErrServerClosed {
		panic(err_http)
	}

	<-idleConnsClosed

}