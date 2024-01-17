package main

import (
	"encoding/json"
	"go-atomicals/pkg/atomicals"
	"go-atomicals/pkg/hashrate"
	"log"
	"os"
	"time"
)

func main() {
	// read from stdin, pase it to input
	var input atomicals.Input
	dec := json.NewDecoder(os.Stdin)
	if dec.Decode(&input) != nil {
		log.Fatalf("decode input error")
	}

	start := time.Now()
	reporter := hashrate.NewReporter()
	// core count
	result := make(chan atomicals.Result, 1)
	go atomicals.Mine(input, result, reporter)
	finalData := <-result
	log.Printf("found solution cost: %v", time.Since(start))

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(finalData); err != nil {
		log.Fatalf("encode output error")
	}
}
