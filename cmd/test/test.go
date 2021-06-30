package main

import (
	"fmt"
	"time"
)

func main()  {
	a()
	time.Sleep(10*time.Second)
}

func a()  {
	go b()
}

func b()  {
	for {fmt.Println("2")}
}