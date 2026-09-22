package main

import (
	"fmt"
	"github.com/PawelReich/gogdb/client"
)

func flatten(value map[string]any) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func main() {
	gdb, err := client.New()

	if err != nil {
		panic(err)
	}

	ret, err := gdb.Send("target-select", "remote", ":3333")

	if err != nil {
		panic(err)
	}

	str, err := flatten(ret)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(str))

}
