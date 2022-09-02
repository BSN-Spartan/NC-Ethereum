// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"io"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/forkid"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// Constants to match up protocol versions and messages
const (
	ETH66 = 66
)

// ProtocolName is the official short name of the `eth` protocol used during
// devp2p capability negotiation.
const ProtocolName = "eth"

// ProtocolVersions are the supported versions of the `eth` protocol (first
// is primary).
var ProtocolVersions = []uint{ETH66}

// protocolLengths are the number of implemented message corresponding to
// different protocol versions.
var protocolLengths = map[uint]uint64{ETH66: 22}

// maxMessageSize is the maximum cap on the size of a protocol message.
const maxMessageSize = 10 * 1024 * 1024

const (
	StatusMsg                     = 0x00
	NewBlockHashesMsg             = 0x01
	TransactionsMsg               = 0x02
	GetBlockHeadersMsg            = 0x03
	BlockHeadersMsg               = 0x04
	GetBlockBodiesMsg             = 0x05
	BlockBodiesMsg                = 0x06
	NewBlockMsg                   = 0x07
	GetNodeDataMsg                = 0x0d
	NodeDataMsg                   = 0x0e
	GetReceiptsMsg                = 0x0f
	ReceiptsMsg                   = 0x10
	NewPooledTransactionHashesMsg = 0x08
	GetPooledTransactionsMsg      = 0x09
	PooledTransactionsMsg         = 0x0a
	BridgeMsg                     = 0x13
	GetHealthCheckMsg             = 0x14
	HealthCheckMsg                = 0x15
)

const (
	BridgeGetHealthCheckMsg = 0x01
)

var (
	errNoStatusMsg             = errors.New("no status message")
	errMsgTooLarge             = errors.New("message too long")
	errDecode                  = errors.New("invalid message")
	errInvalidMsgCode          = errors.New("invalid message code")
	errProtocolVersionMismatch = errors.New("protocol version mismatch")
	errNetworkIDMismatch       = errors.New("network ID mismatch")
	errGenesisMismatch         = errors.New("genesis mismatch")
	errForkIDRejected          = errors.New("fork ID rejected")
)

// Packet represents a p2p message in the `eth` protocol.
type Packet interface {
	Name() string // Name returns a string corresponding to the message type.
	Kind() byte   // Kind returns the message type.
}

// StatusPacket is the network packet for the status message for eth/64 and later.
type StatusPacket struct {
	ProtocolVersion uint32
	NetworkID       uint64
	TD              *big.Int
	Head            common.Hash
	Genesis         common.Hash
	ForkID          forkid.ID
}

// NewBlockHashesPacket is the network packet for the block announcements.
type NewBlockHashesPacket []struct {
	Hash   common.Hash // Hash of one particular block being announced
	Number uint64      // Number of one particular block being announced
}

// Unpack retrieves the block hashes and numbers from the announcement packet
// and returns them in a split flat format that's more consistent with the
// internal data structures.
func (p *NewBlockHashesPacket) Unpack() ([]common.Hash, []uint64) {
	var (
		hashes  = make([]common.Hash, len(*p))
		numbers = make([]uint64, len(*p))
	)
	for i, body := range *p {
		hashes[i], numbers[i] = body.Hash, body.Number
	}
	return hashes, numbers
}

// TransactionsPacket is the network packet for broadcasting new transactions.
type TransactionsPacket []*types.Transaction

// GetBlockHeadersPacket represents a block header query.
type GetBlockHeadersPacket struct {
	Origin  HashOrNumber // Block from which to retrieve headers
	Amount  uint64       // Maximum number of headers to retrieve
	Skip    uint64       // Blocks to skip between consecutive headers
	Reverse bool         // Query direction (false = rising towards latest, true = falling towards genesis)
}

// GetBlockHeadersPacket66 represents a block header query over eth/66
type GetBlockHeadersPacket66 struct {
	RequestId uint64
	*GetBlockHeadersPacket
}

// HashOrNumber is a combined field for specifying an origin block.
type HashOrNumber struct {
	Hash   common.Hash // Block hash from which to retrieve headers (excludes Number)
	Number uint64      // Block hash from which to retrieve headers (excludes Hash)
}

// EncodeRLP is a specialized encoder for HashOrNumber to encode only one of the
// two contained union fields.
func (hn *HashOrNumber) EncodeRLP(w io.Writer) error {
	if hn.Hash == (common.Hash{}) {
		return rlp.Encode(w, hn.Number)
	}
	if hn.Number != 0 {
		return fmt.Errorf("both origin hash (%x) and number (%d) provided", hn.Hash, hn.Number)
	}
	return rlp.Encode(w, hn.Hash)
}

// DecodeRLP is a specialized decoder for HashOrNumber to decode the contents
// into either a block hash or a block number.
func (hn *HashOrNumber) DecodeRLP(s *rlp.Stream) error {
	_, size, err := s.Kind()
	switch {
	case err != nil:
		return err
	case size == 32:
		hn.Number = 0
		return s.Decode(&hn.Hash)
	case size <= 8:
		hn.Hash = common.Hash{}
		return s.Decode(&hn.Number)
	default:
		return fmt.Errorf("invalid input size %d for origin", size)
	}
}

// BlockHeadersPacket represents a block header response.
type BlockHeadersPacket []*types.Header

// BlockHeadersPacket represents a block header response over eth/66.
type BlockHeadersPacket66 struct {
	RequestId uint64
	BlockHeadersPacket
}

// BlockHeadersRLPPacket represents a block header response, to use when we already
// have the headers rlp encoded.
type BlockHeadersRLPPacket []rlp.RawValue

// BlockHeadersPacket represents a block header response over eth/66.
type BlockHeadersRLPPacket66 struct {
	RequestId uint64
	BlockHeadersRLPPacket
}

// NewBlockPacket is the network packet for the block propagation message.
type NewBlockPacket struct {
	Block *types.Block
	TD    *big.Int
}

// sanityCheck verifies that the values are reasonable, as a DoS protection
func (request *NewBlockPacket) sanityCheck() error {
	if err := request.Block.SanityCheck(); err != nil {
		return err
	}
	//TD at mainnet block #7753254 is 76 bits. If it becomes 100 million times
	// larger, it will still fit within 100 bits
	if tdlen := request.TD.BitLen(); tdlen > 100 {
		return fmt.Errorf("too large block TD: bitlen %d", tdlen)
	}
	return nil
}

// GetBlockBodiesPacket represents a block body query.
type GetBlockBodiesPacket []common.Hash

// GetBlockBodiesPacket represents a block body query over eth/66.
type GetBlockBodiesPacket66 struct {
	RequestId uint64
	GetBlockBodiesPacket
}

// BlockBodiesPacket is the network packet for block content distribution.
type BlockBodiesPacket []*BlockBody

// BlockBodiesPacket is the network packet for block content distribution over eth/66.
type BlockBodiesPacket66 struct {
	RequestId uint64
	BlockBodiesPacket
}

// BlockBodiesRLPPacket is used for replying to block body requests, in cases
// where we already have them RLP-encoded, and thus can avoid the decode-encode
// roundtrip.
type BlockBodiesRLPPacket []rlp.RawValue

// BlockBodiesRLPPacket66 is the BlockBodiesRLPPacket over eth/66
type BlockBodiesRLPPacket66 struct {
	RequestId uint64
	BlockBodiesRLPPacket
}

// BlockBody represents the data content of a single block.
type BlockBody struct {
	Transactions []*types.Transaction // Transactions contained within a block
	Uncles       []*types.Header      // Uncles contained within a block
}

// Unpack retrieves the transactions and uncles from the range packet and returns
// them in a split flat format that's more consistent with the internal data structures.
func (p *BlockBodiesPacket) Unpack() ([][]*types.Transaction, [][]*types.Header) {
	var (
		txset    = make([][]*types.Transaction, len(*p))
		uncleset = make([][]*types.Header, len(*p))
	)
	for i, body := range *p {
		txset[i], uncleset[i] = body.Transactions, body.Uncles
	}
	return txset, uncleset
}

// GetNodeDataPacket represents a trie node data query.
type GetNodeDataPacket []common.Hash

// GetNodeDataPacket represents a trie node data query over eth/66.
type GetNodeDataPacket66 struct {
	RequestId uint64
	GetNodeDataPacket
}

// NodeDataPacket is the network packet for trie node data distribution.
type NodeDataPacket [][]byte

// NodeDataPacket is the network packet for trie node data distribution over eth/66.
type NodeDataPacket66 struct {
	RequestId uint64
	NodeDataPacket
}

// GetReceiptsPacket represents a block receipts query.
type GetReceiptsPacket []common.Hash

// GetReceiptsPacket represents a block receipts query over eth/66.
type GetReceiptsPacket66 struct {
	RequestId uint64
	GetReceiptsPacket
}

// ReceiptsPacket is the network packet for block receipts distribution.
type ReceiptsPacket [][]*types.Receipt

// ReceiptsPacket is the network packet for block receipts distribution over eth/66.
type ReceiptsPacket66 struct {
	RequestId uint64
	ReceiptsPacket
}

// ReceiptsRLPPacket is used for receipts, when we already have it encoded
type ReceiptsRLPPacket []rlp.RawValue

// ReceiptsPacket66 is the eth-66 version of ReceiptsRLPPacket
type ReceiptsRLPPacket66 struct {
	RequestId uint64
	ReceiptsRLPPacket
}

// NewPooledTransactionHashesPacket represents a transaction announcement packet.
type NewPooledTransactionHashesPacket []common.Hash

// GetPooledTransactionsPacket represents a transaction query.
type GetPooledTransactionsPacket []common.Hash

type GetPooledTransactionsPacket66 struct {
	RequestId uint64
	GetPooledTransactionsPacket
}

// PooledTransactionsPacket is the network packet for transaction distribution.
type PooledTransactionsPacket []*types.Transaction

// PooledTransactionsPacket is the network packet for transaction distribution over eth/66.
type PooledTransactionsPacket66 struct {
	RequestId uint64
	PooledTransactionsPacket
}

// PooledTransactionsPacket is the network packet for transaction distribution, used
// in the cases we already have them in rlp-encoded form
type PooledTransactionsRLPPacket []rlp.RawValue

// PooledTransactionsRLPPacket66 is the eth/66 form of PooledTransactionsRLPPacket
type PooledTransactionsRLPPacket66 struct {
	RequestId uint64
	PooledTransactionsRLPPacket
}

type GetHealthCheckPacket66 struct {
	RequestId uint64
}

type HealthCheckPacket66 struct {
	RequestId uint64
	*HealthCheckPacket
}

type HealthCheckPacket struct {

	// 邻节点信息
	Peers []string
	// 节点模式 full  light
	SyncMode string
	// 当前块高
	BlockNumber *big.Int
	BlockHash   string
	// 节点验证者地址
	Validator common.Address

	ChainId string
}

type BridgeMsgType uint8

const (
	Ask BridgeMsgType = iota
	Answer
)

type BridgeMsgPacket66 struct {
	RequestId uint64
	*BridgeMsgPacket
}

type BridgeMsgPacket struct {
	Id             string
	Source         enode.ID
	Target         enode.ID
	Expiration     uint64
	Relay          uint
	BridgeType     BridgeMsgType
	Msg            *BridgeMsgData
	RouteSignBytes []byte
	MsgSignBytes   []byte
}

func (b *BridgeMsgPacket) Digest() []byte {

	var data []byte
	data = append(data, []byte(b.Id)...)
	data = append(data, b.Source.Bytes()...)
	data = append(data, b.Target.Bytes()...)
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, b.Expiration)
	data = append(data, buf...)

	return crypto.Keccak256(data)
}
func (b *BridgeMsgPacket) MsgDigest() []byte {

	var data []byte
	data = append(data, uint8(b.BridgeType))
	if b.Msg != nil {
		data = append(data, b.Msg.Bytes()...)
	}
	data = append(data, b.RouteSignBytes...)
	return crypto.Keccak256(data)
}

func (b *BridgeMsgPacket) RouteValidate() (common.Address, error) {
	pubkey, err := crypto.SigToPub(b.Digest(), b.RouteSignBytes)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pubkey), nil
}
func (b *BridgeMsgPacket) RouteSign(key *ecdsa.PrivateKey) error {
	sig, err := crypto.Sign(b.Digest(), key)
	if err != nil {
		return err
	}
	b.RouteSignBytes = sig
	return nil
}

func (b *BridgeMsgPacket) MsgValidate() error {
	pubkey, err := crypto.SigToPub(b.MsgDigest(), b.MsgSignBytes)
	if err != nil {
		return err
	}

	sigId := enode.PubkeyToIDV4(pubkey)

	if b.BridgeType == Answer && sigId != b.Target {
		return errors.New("answer validate failed")
	}

	if b.BridgeType == Ask && sigId != b.Source {
		return errors.New("ask validate failed")
	}
	return nil
}
func (b *BridgeMsgPacket) MsgSign(key *ecdsa.PrivateKey) error {
	sig, err := crypto.Sign(b.MsgDigest(), key)
	if err != nil {
		return err
	}
	b.MsgSignBytes = sig
	return nil
}

func (b *BridgeMsgPacket) Copy() *BridgeMsgPacket {

	n := &BridgeMsgPacket{
		Id:             b.Id,
		Source:         b.Source,
		Target:         b.Target,
		BridgeType:     b.BridgeType,
		Expiration:     b.Expiration,
		Relay:          b.Relay,
		RouteSignBytes: b.RouteSignBytes,
		MsgSignBytes:   b.MsgSignBytes,
	}

	if b.Msg != nil {
		n.Msg = &BridgeMsgData{
			Code:    b.Msg.Code,
			Payload: b.Msg.Payload,
		}
	}

	return n
}

func (b *BridgeMsgPacket) SetMsg(msg *p2p.Msg) {
	pl := make([]byte, msg.Size)
	msg.Payload.Read(pl)
	b.Msg = &BridgeMsgData{
		Code:    msg.Code,
		Payload: pl,
	}
}

type BridgeMsgData struct {
	Code    uint64
	Payload []byte
}

func (d *BridgeMsgData) Bytes() []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, d.Code)

	var data []byte
	data = append(data, buf...)
	data = append(data, d.Payload...)
	return data
}

func (d *BridgeMsgData) MSG() *p2p.Msg {
	m := &p2p.Msg{
		ReceivedAt: time.Now(),
		Code:       d.Code,
		Size:       uint32(len(d.Payload)),
		Payload:    bytes.NewReader(d.Payload),
	}
	return m
}

func (*BridgeMsgPacket) Name() string { return "BridgeMsg" }
func (*BridgeMsgPacket) Kind() byte   { return BridgeMsg }

func (*StatusPacket) Name() string { return "Status" }
func (*StatusPacket) Kind() byte   { return StatusMsg }

func (*NewBlockHashesPacket) Name() string { return "NewBlockHashes" }
func (*NewBlockHashesPacket) Kind() byte   { return NewBlockHashesMsg }

func (*TransactionsPacket) Name() string { return "Transactions" }
func (*TransactionsPacket) Kind() byte   { return TransactionsMsg }

func (*GetBlockHeadersPacket) Name() string { return "GetBlockHeaders" }
func (*GetBlockHeadersPacket) Kind() byte   { return GetBlockHeadersMsg }

func (*BlockHeadersPacket) Name() string { return "BlockHeaders" }
func (*BlockHeadersPacket) Kind() byte   { return BlockHeadersMsg }

func (*GetBlockBodiesPacket) Name() string { return "GetBlockBodies" }
func (*GetBlockBodiesPacket) Kind() byte   { return GetBlockBodiesMsg }

func (*BlockBodiesPacket) Name() string { return "BlockBodies" }
func (*BlockBodiesPacket) Kind() byte   { return BlockBodiesMsg }

func (*NewBlockPacket) Name() string { return "NewBlock" }
func (*NewBlockPacket) Kind() byte   { return NewBlockMsg }

func (*GetNodeDataPacket) Name() string { return "GetNodeData" }
func (*GetNodeDataPacket) Kind() byte   { return GetNodeDataMsg }

func (*NodeDataPacket) Name() string { return "NodeData" }
func (*NodeDataPacket) Kind() byte   { return NodeDataMsg }

func (*GetReceiptsPacket) Name() string { return "GetReceipts" }
func (*GetReceiptsPacket) Kind() byte   { return GetReceiptsMsg }

func (*ReceiptsPacket) Name() string { return "Receipts" }
func (*ReceiptsPacket) Kind() byte   { return ReceiptsMsg }

func (*NewPooledTransactionHashesPacket) Name() string { return "NewPooledTransactionHashes" }
func (*NewPooledTransactionHashesPacket) Kind() byte   { return NewPooledTransactionHashesMsg }

func (*GetPooledTransactionsPacket) Name() string { return "GetPooledTransactions" }
func (*GetPooledTransactionsPacket) Kind() byte   { return GetPooledTransactionsMsg }

func (*PooledTransactionsPacket) Name() string { return "PooledTransactions" }
func (*PooledTransactionsPacket) Kind() byte   { return PooledTransactionsMsg }

func (*HealthCheckPacket) Name() string { return "HealthCheck" }
func (*HealthCheckPacket) Kind() byte   { return HealthCheckMsg }
