//go:build cpu
// +build cpu

package atomicals_test

import (
	"go-atomicals/pkg/atomicals"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func TestCPUMine(t *testing.T) {
	var hash chainhash.Hash
	workc := "aabbccd"
	input := atomicals.Input{
		CopiedData: atomicals.CopiedData{
			Args: atomicals.Args{
				Bitworkc:   &workc,
				MintTicker: "quark",
				Nonce:      274483,
				Time:       1703516711,
			},
		},
		WorkerOptions: atomicals.WorkerOptions{
			OpType: "dmt",
		},
		WorkerBitworkInfoCommit: &atomicals.BitworkInfo{
			Prefix: "8888888",
			Ext:    6,
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
