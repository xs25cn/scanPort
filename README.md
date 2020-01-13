# scanPort
ScanPort 端口扫描工具是一个可以检测服务器或是指定ip段的端口开放情况。

功能：可以快速扫描指定端口范围，ip地址范围。将扫描结果保存到本地！

先来体验一下运行后的效果：

![image.png](https://upload-images.jianshu.io/upload_images/19018717-e8826aa18d92aa57.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

![image.png](https://upload-images.jianshu.io/upload_images/19018717-f62194125b0a4a5b.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)
### 安装使用
方法1. 你可以直接下载已编译好的文件直接运行

https://github.com/xs25cn/scanPort/tree/master/bin

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
        ip地址 例如:-ip 192.168.0.1-255 或直接输入域名 xs25.cn (default "127.0.0.1")
  -n int
        进程数 例如:-n 10 (default 100)
  -p string
        端口号范围 例如:-p 80,81,88-1000 (default "80")
  -path string
        日志地址 例如:-path log (default "log")
  -t int
        超时时长(毫秒) 例如:-t 200 (default 200)

````

#### 例1：指定端口号扫描，如我们要扫描xs25.cn这台服务的开放端口，使用1000个协程进行

scanport -p 80,81,88-3306 -ip xs25.cn -n 1000 

#### 例2：指定IP范围扫描,如我们扫描 192.168.0.1-255 网段的端口 80-10000

scanPort -ip 192.168.0.1-255 -p 80-10000


##### 注：程序扫描完后开放端口放在log目录中，如想更改目录名请加 -path 参数来指定



