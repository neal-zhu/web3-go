package atomicals

import (
	"encoding/binary"
	"encoding/hex"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/fxamacker/cbor/v2"
)

type Fees struct {
	CommitAndRevealFee          int64 `json:"commitAndRevealFee"`
	CommitAndRevealFeePlusOutpu int64 `json:"commitAndRevealFeePlusOutputs"`
	RevealFeePlusOutputs        int64 `json:"revealFeePlusOutputs"`
	CommitFeeOnly               int64 `json:"commitFeeOnly"`
	RevealFeeOnly               int64 `json:"revealFeeOnly"`
}

type FundingUtxo struct {
	Txid        *chainhash.Hash `json:"txid"`
	OutputIndex uint32          `json:"outputIndex"`
	Index       uint32          `json:"index"`
	Vout        uint32          `json:"vout"`
	Value       int64           `json:"value"`
}

type WorkerOptions struct {
	OpType   string `json:"opType"`
	SatsByte int64  `json:"satsbyte"`
}

type Args struct {
	Bitworkc   string `json:"bitworkc"`
	MintTicker string `json:"mint_ticker"`
	Nonce      uint32 `json:"nonce"`
	Time       uint32 `json:"time"`
}

type CopiedData struct {
	Args Args `json:"args"`
}

func (c *Input) MustEncodeCbor() []byte {
	enc, err := cbor.CanonicalEncOptions().EncMode()
	if err != nil {
		log.Fatalf("cbor.CanonicalEncOptions().EncMode() failed: %v", err)
	}
	data, err := enc.Marshal(c.CopiedData)
	if err != nil {
		log.Fatalf("enc.Marshal(c) failed: %v", err)
	}
	return data
}

func (c *Input) MustBuildScript(cborData []byte) []byte {
	var builder txscript.ScriptBuilder
	data, err := builder.AddData(c.KeyPairInfo.ChildNodeXOnlyPubkey).
		AddOp(txscript.OP_CHECKSIG).
		AddOp(txscript.OP_0).
		AddOp(txscript.OP_IF).
		AddData([]byte("atom")).
		AddData([]byte(c.WorkerOptions.OpType)).
		AddData(cborData).
		AddOp(txscript.OP_ENDIF).
		Script()

	if err != nil {
		log.Fatalf("builder.Script() failed: %v", err)
	}
	return data
}

func (i *Input) Init() {
	// DO NOT CHANGE THE ORDER OF THESE CALLS
	i.WorkerBitworkInfoCommit.ParsePreifx()
	i.KeyPairInfo = i.MustBuildKeyPairInfo()
	i.cbor_cache = i.MustEncodeCbor()
	i.script_cache = i.MustBuildScript(i.cbor_cache)
}

type Input struct {
	CopiedData              CopiedData    `json:"copiedData"`
	WorkerOptions           WorkerOptions `json:"workerOptions"`
	FundingUtxo             FundingUtxo   `json:"fundingUtxo"`
	WorkerBitworkInfoCommit BitworkInfo   `json:"workerBitworkInfoCommit"`
	FundingWIF              string        `json:"fundingWIF"`
	Fees                    Fees          `json:"fees"`

	KeyPairInfo *KeyPairInfo `json:"-"`

	script_cache []byte
	cbor_cache   []byte
}

type KeyPairInfo struct {
	ChildNode            *btcec.PrivateKey
	TweakedChildNode     *btcec.PrivateKey
	ChildNodeXOnlyPubkey []byte
	ChildNodePubkey      *btcec.PublicKey
	Ouput                []byte
}

func (i *Input) MustBuildKeyPairInfo() *KeyPairInfo {
	wif, err := btcutil.DecodeWIF(i.FundingWIF)
	if err != nil {
		log.Fatalf("btcec.DecodeWIF(i.FundingWIF) failed: %v", err)
	}
	childNode := wif.PrivKey
	childXOnlyPubkey := schnorr.SerializePubKey(childNode.PubKey())
	outputKey := txscript.ComputeTaprootOutputKey(childNode.PubKey(), nil)
	output, err := txscript.PayToTaprootScript(outputKey)
	tweakedChildNode := txscript.TweakTaprootPrivKey(*childNode, nil)
	if err != nil {
		log.Fatalf("txscript.PayToTaprootScript(childNode.PubKey()) failed: %v", err)
	}

	return &KeyPairInfo{
		ChildNode:            childNode,
		TweakedChildNode:     tweakedChildNode,
		ChildNodePubkey:      childNode.PubKey(),
		ChildNodeXOnlyPubkey: childXOnlyPubkey,
		Ouput:                output,
	}
}

type P2TR struct {
	InternalPubKey *btcec.PublicKey
	OutputKey      *btcec.PublicKey
	Output         []byte
}

func (i *Input) ScriptP2TR(script []byte) *P2TR {
	pubkey := i.KeyPairInfo.ChildNodePubkey
	tapLeaf := txscript.NewBaseTapLeaf(script)
	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)
	tapScriptRootHash := tapScriptTree.RootNode.TapHash()
	outputKey := txscript.ComputeTaprootOutputKey(
		pubkey, tapScriptRootHash[:],
	)
	p2trScript, err := txscript.PayToTaprootScript(outputKey)
	if err != nil {
		log.Fatalf("txscript.PayToTaprootScript(outputKey) failed: %v", err)
	}
	return &P2TR{
		InternalPubKey: i.KeyPairInfo.ChildNodePubkey,
		OutputKey:      outputKey,
		Output:         p2trScript,
	}
}

type BitworkInfo struct {
	Prefix string `json:"prefix"`
	Ext    byte   `json:"ext"`

	PrefixBytes   []byte `json:"-"`
	PrefixPartial *byte  `json:"-"`
}

func (bw *BitworkInfo) ParsePreifx() {
	var (
		prefixBytes []byte
		err         error
	)
	if len(bw.Prefix)&1 == 0 {
		prefixBytes, err = hex.DecodeString(bw.Prefix)
		if err != nil {
			log.Fatalf("hex.DecodeString(bw.Prefix[:len(bw.Prefix)-1]) failed: %v", err)
		}
	} else {
		prefixBytes, err = hex.DecodeString(bw.Prefix[:len(bw.Prefix)-1])
		if err != nil {
			log.Fatalf("hex.DecodeString(bw.Prefix[:len(bw.Prefix)-1]) failed: %v", err)
		}
		bw.PrefixPartial = new(byte)
		if bw.Prefix[len(bw.Prefix)-1] > '9' {
			*bw.PrefixPartial = byte(bw.Prefix[len(bw.Prefix)-1] - 'a' + 10)
		} else {
			*bw.PrefixPartial = byte(bw.Prefix[len(bw.Prefix)-1] - '0')
		}
	}
	bw.PrefixBytes = prefixBytes
}

// 30x faster than simple compare hex prefix
func (bw *BitworkInfo) HasValidBitwork(hash *chainhash.Hash) bool {
	i := 0
	for ; i < len(bw.PrefixBytes); i++ {
		// hash is little endian
		if bw.PrefixBytes[i] != hash[31-i] {
			return false
		}
	}

	if bw.PrefixPartial != nil {
		if hash[31-i]>>4 != *bw.PrefixPartial {
			return false
		}
	}

	if bw.PrefixPartial != nil {
		if hash[31-i]&0xf < bw.Ext {
			return false
		}
	} else {
		if hash[31-i]>>4 < bw.Ext {
			return false
		}
	}
	return true
}

func (i *Input) GetCommitChange() int64 {
	totalInputsValue := i.FundingUtxo.Value
	totalOutputsValue := i.Fees.RevealFeePlusOutputs
	calculatedFee := totalInputsValue - totalOutputsValue
	if calculatedFee <= 0 {
		log.Printf("calculatedFee <= 0, totalInputsValue: %d, totalOutputsValue: %d", totalInputsValue, totalOutputsValue)
		return 0
	}
	expectedFee := i.Fees.CommitFeeOnly + i.WorkerOptions.SatsByte*43
	differenceBetweenCalculatedAndExpected := calculatedFee - expectedFee
	if differenceBetweenCalculatedAndExpected >= 546 {
		return int64(differenceBetweenCalculatedAndExpected)
	}
	return 0
}

func (c *Input) UpdateCborCache() []byte {
	binary.BigEndian.PutUint32(c.cbor_cache[24:], c.CopiedData.Args.Nonce)
	return c.cbor_cache
}

func (c *Input) UpdateScript() []byte {
	binary.BigEndian.PutUint32(c.script_cache[70:], c.CopiedData.Args.Nonce)
	return c.script_cache
}
