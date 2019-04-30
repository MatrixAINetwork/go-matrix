//+build linux darwin

package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

const panicFile = "/tmp/panic.log"

var globalFile *os.File

func initPanicFile() {
	file, err := os.OpenFile(panicFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Create panic file err", err)
	}
	test := time.Now()
	timestr := test.Format("2006-01-02 15:04:05\r\n")
	file.Write([]byte(timestr))
	globalFile = file
	err = syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd()))
	if err != nil {

		fmt.Println("dup2 failed", err)
	}
}
