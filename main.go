package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"scanPort/app/scan"
	"scanPort/app/wsConn"
	"strconv"
	"time"
)

var (
	wConn  *wsConn.WsConnection
	osBase string
	version string
)

func main() {
	var (
		port = flag.Int("port", 25252, "端口号")
		h    = flag.Bool("h", false, "帮助信息")
	)
	version = "v2.0"
	flag.Parse()
	//帮助信息
	if *h == true {
		usage("scanPort version: scanPort/v2.0\n Usage: scanPort [-h] [-ip ip地址] [-n 进程数] [-p 端口号范围] [-t 超时时长] [-path 日志保存路径]\n\nOptions:\n")
		return
	}
	serverUri :=  "https://ip.xs25.cn/?p=" + strconv.Itoa(*port)
	openErr := open(serverUriOsInfo(serverUri))
	if openErr != nil {
		fmt.Println(openErr, serverUri)
	}
	//绑定路由地址
	http.HandleFunc("/", indexHandle)
	http.HandleFunc("/run", runHandle)
	http.HandleFunc("/ws", wsHandle)

	//启动服务端口
	addr := ":" + strconv.Itoa(*port)
	log.Println(" ^_^ 服务已启动...")
	http.ListenAndServe(addr, nil)
}

//首页
func indexHandle(w http.ResponseWriter, r *http.Request) {
	s := "小手端口扫描 "+version+" (by:Duzhenxun)"
	w.Write([]byte(s))
}

//运行
func runHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json

	resp := map[string]interface{}{
		"code": 200,
		"msg":  "ok",
	}
	decoder := json.NewDecoder(r.Body)
	type Params struct {
		Ip      string `json:"ip"`
		Port    string `json:"port"`
		Process int    `json:"process"`
		Timeout int    `json:"timeout"`
		Debug   int    `json:"debug"`
	}
	var params Params
	decoder.Decode(&params)
	if params.Ip == "" {
		resp["code"] = 201
		resp["msg"] = "缺少字段 ip"
		b, _ := json.Marshal(resp)
		w.Write(b)
		return
	}

	if params.Port == "" {
		params.Port = "80"
	}
	if params.Process == 0 {
		params.Process = 10
	}
	if params.Timeout == 0 {
		params.Timeout = 100
	}
	debug := false
	if params.Debug == 0 {
		params.Debug = 1
		debug = true
	}

	//初始化
	scanIP := scan.NewScanIp(params.Timeout, params.Process, debug)
	ips, err := scanIP.GetAllIp(params.Ip)
	if err != nil {
		wConn.WriteMessage(1, []byte(fmt.Sprintf("  ip解析出错....  %v", err.Error())))
		return
	}
	//扫所有的ip
	filePath, _ := mkdir("log")
	fileName := filePath + params.Ip + "_port.txt"
	for i := 0; i < len(ips); i++ {
		ports := scanIP.GetIpOpenPort(ips[i], params.Port,wConn)
		if len(ports) > 0 {
			f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				if err := f.Close(); err != nil {
					fmt.Println(err)
				}
				continue
			}
			if _, err := f.WriteString(fmt.Sprintf("%v【%v】开放:%v \n", time.Now().Format("2006-01-02 15:04:05"),ips[i], ports)); err != nil {
				if err := f.Close(); err != nil {
					fmt.Println(err)
				}
				continue
			}
		}
	}
	open(fileName)
	b, _ := json.Marshal(resp)
	w.Write(b)
	return
}

//ws服务
func wsHandle(w http.ResponseWriter, r *http.Request) {
	wsUp := websocket.Upgrader{
		HandshakeTimeout: time.Second * 5,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: false,
	}
	wsSocket, err := wsUp.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}
	wConn = wsConn.New(wsSocket)
	for {
		data, err := wConn.ReadMessage()
		if err != nil {
			wConn.Close()
			return
		}
		if err := wConn.WriteMessage(data.MessageType, data.Data); err != nil {
			wConn.Close()
			return
		}
	}
}

//系统信息
func serverUriOsInfo(serverUri string)  string{
	osInfo := map[string]interface{}{}
	osInfo["version"] = version
	osInfo["os"] = runtime.GOOS
	osInfo["cpu"] = runtime.NumCPU()
	addrArr, _ := net.InterfaceAddrs()
	osInfo["addr"] = fmt.Sprint(addrArr)
	osInfo["time"] = time.Now().Unix()
	osInfos, _ := json.Marshal(osInfo)
	osBase = base64.StdEncoding.EncodeToString(osInfos)
	token := hmacSha256(osBase, "dzx")
	serverUri += "Z00X" + base64.StdEncoding.EncodeToString([]byte("&os_base||"+osBase+"&token||"+token))
	return serverUri
}

func usage(str string) {
	fmt.Fprintf(os.Stderr, str)
	flag.PrintDefaults()
}
func mkdir(path string) (string, error) {
	delimiter := "/"
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	filePtah := dir + delimiter + path + delimiter
	err := os.MkdirAll(filePtah, 0777)
	if err != nil {
		return "", err
	}
	return filePtah, nil
}
func open(uri string) error {
	var commands = map[string]string{
		"windows": "start",
		"darwin":  "open",
		"linux":   "xdg-open",
	}
	run, ok := commands[runtime.GOOS]
	if !ok {
		return fmt.Errorf("%s platform ？？？", runtime.GOOS)
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "start ", uri)
		//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	} else {
		cmd = exec.Command(run, uri)
	}
	return cmd.Start()
}
func hmacSha256(src string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(src))
	shaStr := fmt.Sprintf("%x", h.Sum(nil))
	//shaStr:=hex.EncodeToString(h.Sum(nil))
	return base64.StdEncoding.EncodeToString([]byte(shaStr))
}
