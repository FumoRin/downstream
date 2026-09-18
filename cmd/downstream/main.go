package main

import (
	"log"

	"github.com/fumorin/downstream/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
