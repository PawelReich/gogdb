package main

import (
	"fmt"
	"github.com/PawelReich/gogdb/client"
	"os"
)

func main() {
	fmt.Println("gogdb")

	gdb, err := client.New()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ret, err := gdb.Send("target-select", "remote", ":3333")

	if err != nil {
		fmt.Printf("err: %s", err)
		os.Exit(1)
	}

	fmt.Println(ret)

}
