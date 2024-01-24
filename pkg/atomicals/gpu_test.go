//go:build cuda
// +build cuda

package atomicals_test

import (
	"go-atomicals/pkg/atomicals"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func TestGPUMine(t *testing.T) {
	var hash chainhash.Hash
	workc := "aabbcc"
	input := atomicals.Input{
		CopiedData: atomicals.CopiedData{
			Args: atomicals.Args{
				Bitworkc:   &workc,
				MintTicker: "quark",
				Nonce:      274483,
				Time:       uint32(time.Now().Unix()),
			},
		},
		WorkerOptions: atomicals.WorkerOptions{
			OpType: "dmt",
		},
		WorkerBitworkInfoCommit: &atomicals.BitworkInfo{
			Prefix: "aabbccd12",
		},
		FundingWIF: "Kz9gWCiZGnHzxQBpbJW6UShBmxrMQXHktEfYAUcsbFkcyNcEAKzA",
		FundingUtxo: atomicals.FundingUtxo{
			Txid: &hash,
		},
	}
	input.Init()
	for i := 0; i < 10; i++ {
		t.Logf("test threads %d", 15+i)
		result := make(chan atomicals.Result, 1)
		start := time.Now()
		go atomicals.Mine(input, uint32(20+i), result)
		r := <-result
		t.Logf("%+v cost: %v", r, time.Since(start))
	}
}
