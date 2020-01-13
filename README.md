# scanPort
ScanPort 端口扫描工具是一个可以检测服务器或是指定ip段的端口开放情况。

### 安装使用
方法1. 你可以直接下载已编译好的文件直接运行

https://github.com/xs25cn/scanPort/bin

方法2. 你可以使用go命名进行下载，下载完成后会自动安装到你的$GOPATH/bin目录中

```
go get -v github.com/xs25cn/scanPort

```

### 帮助信息
````
scanPort -h 
Options:
  -h    帮助信息
  -ip string
        ip地址 例如:-ip=192.168.0.1-255 或直接输入域名 xs25.cn (default "127.0.0.1")
  -n int
        进程数 例如:-n=10 (default 100)
  -p string
        端口号范围 例如:-p=80,81,88-1000 (default "80")
  -path string
        日志地址 例如:-path=log (default "log")
  -t int
        超时时长(毫秒) 例如:-t=200 (default 200)

````
#### 指定IP范围扫描

scanPort -ip=192.168.0.1-255

#### 指定端口号扫描，如我们要扫描xs25.cn这台服务的开放端口，使用1000个协程进行

scanport -p=80,81,88-3306 -ip=xs25.cn -n=1000 


