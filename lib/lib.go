package lib

import (
	"flag"
	"fmt"
	"os"
)

func Usage(str string) {
	fmt.Fprintf(os.Stderr, str)
	flag.PrintDefaults()
}

func Mkdir(path string){
	f, err := os.Stat(path)
	if err != nil || f.IsDir() == false {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			fmt.Println("创建目录失败！", err)
			return
		}
	}
}