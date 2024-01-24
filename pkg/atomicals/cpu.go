//go:build cpu
// +build cpu

package atomicals

import (
	"bytes"
	"encoding/binary"
	"runtime"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func mine(i int, input Input, resultCh chan<- Result) {
	// set different time for each goroutine
	input.CopiedData.Args.Time += uint32(i)
	// use uint32 so we can avoid cbor encoding at runtime
	input.CopiedData.Args.Nonce = uint32(^uint16(0)) + 1
	input.Init()

	result := MineCommitTx(i, &input)
	revealTxHash, embed := MineRevealTx(&input, result)
	result.RevealTxhash = &revealTxHash
	result.Embed = string(embed)
	resultCh <- *result
}

func MineCommitTx(i int, input *Input) *Result {
	msgTx := wire.NewMsgTx(wire.TxVersion)
	output := wire.NewOutPoint(input.FundingUtxo.Txid, input.FundingUtxo.Index)
	txIn := wire.NewTxIn(output, nil, nil)
	txIn.Sequence = 0
	msgTx.AddTxIn(txIn)

	scriptP2TR := input.MustBuildScriptP2TR()
	txOut := wire.NewTxOut(int64(input.Fees.RevealFeePlusOutputs), scriptP2TR.Output)
	msgTx.AddTxOut(txOut)
	// add change utxo
	if change := input.GetCommitChange(); change != 0 {
		msgTx.AddTxOut(wire.NewTxOut(change, input.KeyPairInfo.Ouput))
	}

	buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSizeStripped()))
	msgTx.SerializeNoWitness(buf)
	serializedTx := buf.Bytes()
	var hash chainhash.Hash
	for {
		hash = chainhash.DoubleHashH(serializedTx)
		if input.WorkerBitworkInfoCommit.HasValidBitwork(&hash) {
			break
		}
		txIn.Sequence++
		binary.LittleEndian.PutUint32(serializedTx[42:], txIn.Sequence)
		if txIn.Sequence == MAX_SEQUENCE {
			input.CopiedData.Args.Nonce++
			scriptP2TR := input.MustBuildScriptP2TR()
			txOut.PkScript = scriptP2TR.Output
			txIn.Sequence = 0
			buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSizeStripped()))
			msgTx.SerializeNoWitness(buf)
			serializedTx = buf.Bytes()
		}
	}
	return &Result{
		FinalCopyData: input.CopiedData,
		FinalSequence: txIn.Sequence,
		CommitTxHash:  &hash,
	}
}

func Mine(input Input, result chan<- Result) {
	for i := 0; i < runtime.NumCPU(); i++ {
		go mine(i, input, result)
	}
}
