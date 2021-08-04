package scan

import (
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//ip 扫描
type ScanIp struct {
	debug   bool
	timeout int
	process int
}

func NewScanIp(timeout int,process int,debug bool) *ScanIp {
	return &ScanIp{
		debug:debug,
		timeout:timeout,
		process:process,
	}
}

//获取开放端口号
func (s *ScanIp) GetIpOpenPort(ip string, port string) []int {
	var (
		total     int
		pageCount int
		num       int
		openPorts []int
		mutex     sync.Mutex
	)
	ports, _ := s.getAllPort(port)
	total = len(ports)
	if total < s.process {
		pageCount = total
	} else {
		pageCount = s.process
	}
	num = int(math.Ceil(float64(total) / float64(pageCount)))

	s.sendLog(fmt.Sprintf("%v 【%v】需要扫描端口总数:%v 个，总协程:%v 个，每个协程处理:%v 个，超时时间:%v毫秒", time.Now().Format("2006-01-02 15:04:05"), ip, total, pageCount, num, s.timeout))
	start := time.Now()
	all := map[int][]int{}
	for i := 1; i <= pageCount; i++ {
		for j := 0; j < num; j++ {
			tmp := (i-1)*num + j
			if tmp < total {
				all[i] = append(all[i], ports[tmp])
			}
		}
	}

	wg := sync.WaitGroup{}
	for k, v := range all {
		wg.Add(1)
		go func(value []int, key int) {
			defer wg.Done()
			var tmpPorts []int
			for i := 0; i < len(value); i++ {
				opened := s.isOpen(ip, value[i])
				if opened {
					tmpPorts = append(tmpPorts, value[i])
				}
			}
			mutex.Lock()
			openPorts = append(openPorts, tmpPorts...)
			mutex.Unlock()
			if len(tmpPorts) > 0 {
				s.sendLog(fmt.Sprintf("%v 【%v】协程%v 执行完成，时长： %.3fs，开放端口： %v", time.Now().Format("2006-01-02 15:04:05"), ip, key, time.Since(start).Seconds(), tmpPorts))
			}
		}(v, k)
	}
	wg.Wait()
	s.sendLog(fmt.Sprintf("%v 【%v】扫描结束，执行时长%.3fs , 所有开放的端口:%v", time.Now().Format("2006-01-02 15:04:05"), ip, time.Since(start).Seconds(), openPorts))
	return openPorts
}

//获取所有ip
func (s *ScanIp) GetAllIp(ip string) ([]string, error) {
	var (
		ips []string
	)

	ipTmp := strings.Split(ip, "-")
	firstIp, err := net.ResolveIPAddr("ip", ipTmp[0])
	if err != nil {
		return ips, errors.New(ipTmp[0] + "域名解析失败" + err.Error())
	}
	if net.ParseIP(firstIp.String()) == nil {
		return ips, errors.New(ipTmp[0] + " ip地址有误~")
	}
	//域名转化成ip再塞回去
	ipTmp[0] = firstIp.String()
	ips = append(ips, ipTmp[0]) //最少有一个ip地址

	if len(ipTmp) == 2 {
		//以切割第一段ip取到最后一位
		ipTmp2 := strings.Split(ipTmp[0], ".")
		startIp, _ := strconv.Atoi(ipTmp2[3])
		endIp, err := strconv.Atoi(ipTmp[1])
		if err != nil || endIp < startIp {
			endIp = startIp
		}
		if endIp > 255 {
			endIp = 255
		}
		totalIp := endIp - startIp + 1
		for i := 1; i < totalIp; i++ {
			ips = append(ips, fmt.Sprintf("%s.%s.%s.%d", ipTmp2[0], ipTmp2[1], ipTmp2[2], startIp+i))
		}
	}
	return ips, nil
}



//记录日志
func (s *ScanIp) sendLog(str string) {
	if s.debug == true {
		fmt.Println(str)
	}
}


//获取所有端口
func (s *ScanIp) getAllPort(port string) ([]int, error) {
	var ports []int
	//处理 ","号 如 80,81,88 或 80,88-100
	portArr := strings.Split(strings.Trim(port, ","), ",")
	for _, v := range portArr {
		portArr2 := strings.Split(strings.Trim(v, "-"), "-")
		startPort, err := s.filterPort(portArr2[0])
		if err != nil {
			continue
		}
		//第一个端口先添加
		ports = append(ports, startPort)
		if len(portArr2) > 1 {
			//添加第一个后面的所有端口
			endPort, _ := s.filterPort(portArr2[1])
			if endPort > startPort {
				for i := 1; i <= endPort-startPort; i++ {
					ports = append(ports, startPort+i)
				}
			}
		}
	}
	//去重复
	ports = s.arrayUnique(ports)

	return ports, nil
}

//端口合法性过滤
func (s *ScanIp) filterPort(str string) (int, error) {
	port, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	if port < 1 || port > 65535 {
		return 0, errors.New("端口号范围超出")
	}
	return port, nil
}



//查看端口号是否打开
func (s *ScanIp) isOpen(ip string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), time.Millisecond*time.Duration(s.timeout))
	if err != nil {
		if strings.Contains(err.Error(),"too many open files"){
			fmt.Println("连接数超出系统限制！"+err.Error())
			os.Exit(1)
		}
		return false
	}
	_ = conn.Close()
	return true
}

//数组去重
func (s *ScanIp) arrayUnique(arr []int) []int {
	var newArr []int
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return newArr
}
