// forward 部分 和 handlePost 中的 Forward 逻辑重合，可以考虑合并
// 建立类似于如下的 ws/wss 连接
// 10.211.55.54 --> 10.211.55.57 --> 10.211.55.58 --> 10.211.55.59 --> 10.211.55.60 --> 10.211.55.61 --> 10.211.55.52
// client -- HTTP POST --> 10.211.55.54 --> [...] --> 10.211.55.52 --> hanele data
package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var (
	nodeID   = flag.String("id", "", "node id, length is 5") // TODO: just for test
	wsAddr   = flag.String("wsAddr", "0.0.0.0:9002", "ws service address")
	httpAddr = flag.String("httpAddr", "0.0.0.0:9003", "http service address")
	useSSL   = flag.Bool("ssl", false, "是否使用 ssl 加密通信")
	upgrader = websocket.Upgrader{} // use default options
)

// set log
type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format("2006-01-02 15:04:05") + " " + string(bytes))
}

// default route map
// 54(f) 57(f) 58(f) 59(f) 60(f) 52(h)

const (
	DefaultTaskID = "0000000000"
	HelloTaskID   = "0000000001"
	ForwardTask   = iota
	HandleTask    // default is store
)

type Task struct {
	ID         string
	NodeID     string
	Type       int
	NextNodeID string
}

type NodeInfo struct {
	ID     string
	IP     string
	WSPort string
}

type NodeConnection struct {
	ID           string
	ConnectNodes map[string]NodeInfo
	Connections  map[string]*websocket.Conn
}

var Nodes = map[string]NodeInfo{
	"00054": {ID: "00054", IP: "10.211.55.54", WSPort: "9002"},
	"00057": {ID: "00057", IP: "10.211.55.57", WSPort: "9002"},
	"00058": {ID: "00058", IP: "10.211.55.58", WSPort: "9002"},
	"00059": {ID: "00059", IP: "10.211.55.59", WSPort: "9002"},
	"00060": {ID: "00060", IP: "10.211.55.60", WSPort: "9002"},
	"00061": {ID: "00060", IP: "10.211.55.61", WSPort: "9002"},
	"00052": {ID: "00052", IP: "10.211.55.52", WSPort: "9002"},
}

// var Connections = map[string]NodeConnection{
// 	// "00054": {"00054", map[string]NodeInfo{"00052": Nodes["00052"]}, make(map[string]*websocket.Conn)},
// 	"00054": {"00054", map[string]NodeInfo{"00057": Nodes["00057"]}, make(map[string]*websocket.Conn)},
// 	"00057": {"00057", map[string]NodeInfo{"00052": Nodes["00052"]}, make(map[string]*websocket.Conn)},
// 	"00052": {"00052", map[string]NodeInfo{}, make(map[string]*websocket.Conn)},
// }

// var TaskList = []Task{
// 	// {ID: DefaultTaskID, NodeID: "00054", Type: ForwardTask, NextNodeID: "00052"},
// 	{ID: DefaultTaskID, NodeID: "00054", Type: ForwardTask, NextNodeID: "00057"},
// 	{ID: DefaultTaskID, NodeID: "00057", Type: ForwardTask, NextNodeID: "00052"},
// 	{ID: DefaultTaskID, NodeID: "00052", Type: HandleTask, NextNodeID: ""},
// }

var Connections = map[string]NodeConnection{
	"00054": {"00054", map[string]NodeInfo{"00057": Nodes["00057"]}, make(map[string]*websocket.Conn)},
	"00057": {"00057", map[string]NodeInfo{"00058": Nodes["00058"]}, make(map[string]*websocket.Conn)},
	"00058": {"00058", map[string]NodeInfo{"00059": Nodes["00059"]}, make(map[string]*websocket.Conn)},
	"00059": {"00059", map[string]NodeInfo{"00060": Nodes["00060"]}, make(map[string]*websocket.Conn)},
	"00060": {"00060", map[string]NodeInfo{"00061": Nodes["00061"]}, make(map[string]*websocket.Conn)},
	"00061": {"00061", map[string]NodeInfo{"00052": Nodes["00052"]}, make(map[string]*websocket.Conn)},
	"00052": {"00052", map[string]NodeInfo{}, make(map[string]*websocket.Conn)},
}

var TaskList = []Task{
	{ID: DefaultTaskID, NodeID: "00054", Type: ForwardTask, NextNodeID: "00057"},
	{ID: DefaultTaskID, NodeID: "00057", Type: ForwardTask, NextNodeID: "00058"},
	{ID: DefaultTaskID, NodeID: "00058", Type: ForwardTask, NextNodeID: "00059"},
	{ID: DefaultTaskID, NodeID: "00059", Type: ForwardTask, NextNodeID: "00060"},
	{ID: DefaultTaskID, NodeID: "00060", Type: ForwardTask, NextNodeID: "00061"},
	{ID: DefaultTaskID, NodeID: "00061", Type: ForwardTask, NextNodeID: "00052"},
	{ID: DefaultTaskID, NodeID: "00052", Type: HandleTask, NextNodeID: ""},
}

func GetTaskDefault(nodeID string) Task {
	for i := range TaskList {
		if TaskList[i].NodeID == nodeID {
			return TaskList[i]
		}
	}
	panic("cannot find task for node " + nodeID)
}

// Fake usage above
// ============

func init() {
	// set log
	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	flag.Parse()
	if *nodeID == "" {
		panic("must have node ID")
	}

	// 这里先跟相应的 ws server 建立连接
	// 当然，也可以懒连接，在使用到的时候再连

	for _, node := range Connections[*nodeID].ConnectNodes {
		conn, err := GetWSConnection(node)
		if err != nil {
			log.Println("currently node " + node.ID + " is not available")
			continue
		}
		Connections[*nodeID].Connections[node.ID] = conn
	}
}

func handleDefaultPost(resp http.ResponseWriter, req *http.Request) {
	file, _, err := req.FormFile("upload")
	if err != nil {
		log.Fatal("err when handle post data", err)
	}
	defer file.Close()

	// get connection
	task := GetTaskDefault(*nodeID)
	// forward Task
	nextNodeID := task.NextNodeID

	// get the connection, if not, just connect to it
	connWithNextNode, ok := Connections[*nodeID].Connections[nextNodeID]
	if !ok {
		nextNode, nodeExists := Nodes[nextNodeID]
		if !nodeExists {
			log.Fatal("the node is not exists: " + nextNodeID)
		}
		if connWithNextNode, err = GetWSConnection(nextNode); err != nil {
			panic(err)
		}
		Connections[*nodeID].Connections[nextNodeID] = connWithNextNode
	}
	nextWriter, err := connWithNextNode.NextWriter(websocket.BinaryMessage)
	if err != nil {
		log.Fatal(err)
	}
	_, _ = nextWriter.Write([]byte(DefaultTaskID + *nodeID))
	if _, err = io.Copy(nextWriter, file); err != nil {
		log.Fatal("error when forwarding msg", err)
	}
	nextWriter.Close()

	// response
	_, nextReader, _ := connWithNextNode.NextReader()
	if _, err = io.Copy(resp, nextReader); err != nil {
		log.Fatal("err when response to client", err)
	}
}

func httpServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleDefaultPost)
	server := http.Server{
		Addr:    *httpAddr,
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}

func handleWSConn(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		t, rawReader, err := c.NextReader()
		if err != nil {
			log.Println("getnextreader err: ", err)
			break
		}
		// 前十位为信息的ID，后五位为来源主机的信息
		// 000000000 为测试任务
		// 000000001 为 hello 信号，建立第一次连接
		id := make([]byte, 15)
		if _, err = rawReader.Read(id); err != nil {
			log.Fatal("error when reading the header msg", err)
		}
		taskID := string(id[:10])
		fromNodeID := string(id[10:])
		if t == websocket.BinaryMessage {
			fmt.Println("get task:", string(taskID), ", from node ", fromNodeID)

			// 如果是 hello task，则不作处理
			if taskID == HelloTaskID {
				continue
			}

			task := GetTaskDefault(*nodeID)
			if task.Type == ForwardTask {
				// forward Task
				nextNodeID := task.NextNodeID

				// get the connection, if not, just connect to it
				connWithNextNode, ok := Connections[*nodeID].Connections[nextNodeID]
				if !ok {
					nextNode, nodeExists := Nodes[nextNodeID]
					if !nodeExists {
						log.Fatal("the node is not exists: " + nextNodeID)
					}
					if connWithNextNode, err = GetWSConnection(nextNode); err != nil {
						panic(err)
					}
					Connections[*nodeID].Connections[nextNodeID] = connWithNextNode
				}
				nextWriter, err := connWithNextNode.NextWriter(websocket.BinaryMessage)
				if err != nil {
					log.Fatal(err)
				}
				_, _ = nextWriter.Write([]byte(taskID + *nodeID))
				_, err = io.Copy(nextWriter, rawReader)
				if err != nil {
					log.Println("error happens when forwarding data", err)
				}
				_ = nextWriter.Close()
				_, data, err := connWithNextNode.ReadMessage()
				if err != nil {
					log.Fatal("error when get response", err)
				}

				// response to previous
				// warning!! this type is TextMessage
				writer, _ := c.NextWriter(websocket.TextMessage)
				_, _ = writer.Write(data)
				writer.Close()
			} else if task.Type == HandleTask {
				// TODO: default handle type, save to file
				saveFileName := "data.iso"
				func() {
					f, err := os.Create(saveFileName)
					if err != nil {
						log.Fatal("err when create saving file", err)
					}
					defer f.Close()
					if _, err := io.Copy(f, rawReader); err != nil {
						log.Fatal("err when reading data to save", err)
					}
				}()

				writer, _ := c.NextWriter(websocket.TextMessage)
				_, _ = writer.Write([]byte("ok"))
				writer.Close()
			}
		}

	}
}

func wsServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleWSConn)
	server := http.Server{
		Addr:    *wsAddr,
		Handler: mux,
	}
	if !*useSSL {
		log.Fatal(server.ListenAndServe())
	} else {
		log.Println("using ssl to secure :)")
		log.Fatal(server.ListenAndServeTLS("server-ssl/server.crt", "server-ssl/server.key"))
	}
}

func loadCA(caFile string) *x509.CertPool {
	pool := x509.NewCertPool()
	if ca, e := ioutil.ReadFile(caFile); e != nil {
		log.Fatal("ReadFile: ", e)
	} else {
		pool.AppendCertsFromPEM(ca)
	}
	return pool
}

func GetWSConnection(nextNode NodeInfo) (*websocket.Conn, error) {
	var (
		conn *websocket.Conn
		u    url.URL
		err  error
	)
	if !*useSSL {
		u = url.URL{Scheme: "ws", Host: nextNode.IP + ":" + nextNode.WSPort, Path: ""}
	} else {
		log.Println("using ssl to secure :)")
		u = url.URL{Scheme: "wss", Host: nextNode.IP + ":" + nextNode.WSPort, Path: ""}
		websocket.DefaultDialer.TLSClientConfig = &tls.Config{RootCAs: loadCA("server-ssl/ca.crt")}
	}
	if conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil); err != nil {
		log.Println("error when connection to node "+u.String(), err)
		return nil, err
	}
	writer, _ := conn.NextWriter(websocket.BinaryMessage)
	_, _ = writer.Write([]byte(HelloTaskID + *nodeID))
	writer.Close()
	log.Println("establish with " + u.String() + " successfully")
	return conn, nil
}

func main() {
	go httpServer()
	go wsServer()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
