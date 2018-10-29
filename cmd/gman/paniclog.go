//+build linux

package main

import (
	"fmt"
	"os"
	"syscall"
)

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
