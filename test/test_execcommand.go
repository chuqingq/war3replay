package main

import (
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("cmd", "/c", "start http://www.baidu.com")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
