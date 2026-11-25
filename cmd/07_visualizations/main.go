package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/rinatkh/classroom8/internal/demo"
)

func main() {
	bufferOnly := flag.Bool("buffer", false, "JSON дополнительной сцены Linear Buffer")
	flag.Parse()
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	var data any = demo.Traces()
	if *bufferOnly {
		data = demo.BufferTrace()
	}
	if err := encoder.Encode(data); err != nil {
		log.Fatal(err)
	}
}
