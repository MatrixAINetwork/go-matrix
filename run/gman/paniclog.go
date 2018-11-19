//+build linux, darwin

package main

import (
	"fmt"
	"os"
	"syscall"
)

const panicFile = "/tmp/panic.log"
var globalFile *os.File

func initPanicFile() {
	file, err := os.OpenFile(panicFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Create panic file err", err)
	}
	globalFile = file
	err = syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd()))
	if err != nil {

		fmt.Println("dup2 failed", err)
	}
}
