package main

import (
	"encoding/json"
	"go-atomicals/pkg/atomicals"
	"go-atomicals/pkg/hashrate"
	"log"
	"syscall/js"
	"time"
)

func main() {
	js.Global().Set("mine", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
			}
		}()
		// read from stdin, pase it to input
		var input atomicals.Input
		if err := json.Unmarshal([]byte(args[0].String()), &input); err != nil {
			return "{}"
		}
		start := time.Now()
		reporter := hashrate.NewReporter()
		// core count
		resultCh := make(chan atomicals.Result, 1)
		go atomicals.Mine(input, resultCh, reporter)
		result := <-resultCh

		log.Printf("found solution cost: %v", time.Since(start))

		d, _ := json.Marshal(result)
		return string(d)
	}))
	c := make(chan struct{}, 0)
	<-c
}
