package main

import (
	"fmt"
	"log"
	"scaffold/cmd"
)

func main() {
	fmt.Println("cli Start")
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
