package main

import (
	"fmt"
	"time"

	"github.com/wulfixyt/anubis-monitors/pkg/socket"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			time.Sleep(2 * time.Second)
			main()
			return
		}
	}()

	socket.Connect()
}
