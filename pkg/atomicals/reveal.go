package atomicals

import (
	"bytes"
	"fmt"
	"go-atomicals/pkg/hashrate"
	"log"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func MineRevealTx(input *Input, result *Result, reporter *hashrate.HashRateReporter) (revealTxHash chainhash.Hash, data []byte) {
	var (
		err   error
		embed []byte
	)

	if input.WorkerBitworkInfoReveal == nil {
		return
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	// commit tx as input
	outpoint := wire.NewOutPoint(result.CommitTxHash, 0)
	tx.AddTxIn(wire.NewTxIn(outpoint, nil, nil))
	if input.RBF {
		tx.TxIn[0].Sequence = 0xfffffffd
	}

	// 加入所有的输出
	for _, output := range input.AdditionOutputs {
		tx.AddTxOut(wire.NewTxOut(output.Value, output.Output()))
	}

	nonce := 0
	now := time.Now().Unix()
	data = []byte(fmt.Sprintf("%d:%08d", now, nonce))
	embed, err = txscript.NullDataScript(data)
	if err != nil {
		log.Fatalf("txscript.NullDataScript(data) failed: %v", err)
	}
	output := wire.NewTxOut(0, embed)
	tx.AddTxOut(output)

	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSizeStripped()))
	tx.SerializeNoWitness(buf)
	serializedTx := buf.Bytes()
	embedIndex := bytes.Index(serializedTx, embed)

	for {
		revealTxHash = chainhash.DoubleHashH(serializedTx)
		if input.WorkerBitworkInfoReveal.HasValidBitwork(&revealTxHash) {
			output.PkScript = embed
			PrintMsgTx(tx)
			break
		}
		nonce++
		data = []byte(fmt.Sprintf("%d:%08d", now, nonce))
		embed, _ := txscript.NullDataScript(data)
		copy(serializedTx[embedIndex:], embed)
		if nonce == 10000000 {
			nonce = 0
			now = time.Now().Unix()
		}
	}

	return
}
