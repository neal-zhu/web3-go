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
	input := atomicals.Input{
		CopiedData: atomicals.CopiedData{
			Args: atomicals.Args{
				Bitworkc:   "aabbcc",
				MintTicker: "quark",
				Nonce:      274483,
				Time:       1703516711,
			},
		},
		WorkerOptions: atomicals.WorkerOptions{
			OpType: "dmt",
		},
		WorkerBitworkInfoCommit: atomicals.BitworkInfo{
			Prefix: "aabbccde",
			Ext:    5,
		},
		FundingWIF: "Kz9gWCiZGnHzxQBpbJW6UShBmxrMQXHktEfYAUcsbFkcyNcEAKzA",
		FundingUtxo: atomicals.FundingUtxo{
			Txid: &hash,
		},
	}
	input.Init()
	result := make(chan atomicals.Result, 1)
	start := time.Now()
	go atomicals.Mine(input, result)
	r := <-result
	t.Logf("%+v cost: %v", r, time.Since(start))
}
