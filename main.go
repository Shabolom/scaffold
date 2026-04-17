package main

import (
	"fmt"
	"os"
	"scaffold/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
