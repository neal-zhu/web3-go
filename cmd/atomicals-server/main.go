package main

import (
	"encoding/json"
	"go-atomicals/pkg/atomicals"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/mine", func(w http.ResponseWriter, r *http.Request) {
		var input atomicals.Input
		dec := json.NewDecoder(r.Body)
		if dec.Decode(&input) != nil {
			log.Fatalf("decode input error")
		}
		start := time.Now()
		result := make(chan atomicals.Result, 1)
		go atomicals.Mine(input, result)
		finalData := <-result
		log.Printf("found solution cost: %v", time.Since(start))

		enc := json.NewEncoder(w)
		if err := enc.Encode(finalData); err != nil {
			log.Fatalf("encode output error")
		}
	})

	// start a http server
	http.ListenAndServe(os.Getenv("GO_SERVER_ADDR"), nil)

}
