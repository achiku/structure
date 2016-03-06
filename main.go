package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	filePath := flag.String("f", "", "source file")
	flag.Parse()

	var st []*Structure
	var err error
	if *filePath != "" {
		st, err = JSONParse(*filePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("no file specified")
	}

	for _, s := range st {
		fmt.Printf("%s", s.String(true))
	}
}
