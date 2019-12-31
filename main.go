package main

import (
	"gowget/wget"
	"os"
	"runtime"
)



func main()  {

	runtime.GOMAXPROCS(runtime.NumCPU())
	w := wget.New()
	w.Run(os.Args[1:])
}
