package eth

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"math/rand"
	"sync"
	"time"
)

const (
	clearInterval = 15 * time.Minute
)

type bridgeMsgSet struct {
	bridgeMses map[string]*bridgeMsgData
	lock       sync.RWMutex
}

func newBridgeMsgSet() *bridgeMsgSet {
	b := &bridgeMsgSet{
		bridgeMses: make(map[string]*bridgeMsgData),
	}

	go b.clearLoop()

	return b
}

func (b *bridgeMsgSet) clearLoop() {
	clearTicker := time.NewTicker(clearInterval)
	for {
		select {
		case <-clearTicker.C:
			go b.clear()
		}
	}
}

func (b *bridgeMsgSet) clear() {
	b.lock.Lock()
	defer b.lock.Unlock()
	var clearKeys []string
	for s, data := range b.bridgeMses {
		if data.IsExpiration() {
			clearKeys = append(clearKeys, s)
		}
	}

	for _, key := range clearKeys {
		log.Debug("")
		delete(b.bridgeMses, key)
	}
}

// checkDiscard check bridge message ,discard invalid message.
// if message is repeated ask or answer , discard
func (b *bridgeMsgSet) checkDiscard(msg *bridgeMsgData) bool {
	// expired message
	if msg.IsExpiration() {
		log.Debug("bridge message expired", "id", msg.Id)
		return true
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	old, ok := b.bridgeMses[msg.Id]
	if !ok {
		// new message
		b.bridgeMses[msg.Id] = msg
		return false
	}

	// new msg type is ask
	if msg.data.BridgeType == eth.Ask {
		// save short path , [ need ?]
		// If the path is short, why is the message late ?
		if old.data.BridgeType == eth.Ask && old.data.Relay > msg.data.Relay {
			b.bridgeMses[msg.Id] = msg
		}
		log.Debug("repeat ask bridge msg", "id", msg.Id)
		return true
	}

	// my record is the answer , discard any new message
	if old.data.BridgeType == eth.Answer {
		log.Debug("repeat answer bridge msg", "id", msg.Id)
		return true
	}

	// old msg is ask , new msg is answer
	old.data.BridgeType = msg.data.BridgeType
	old.data.Msg = msg.data.Msg
	old.data.MsgSignBytes = msg.data.MsgSignBytes
	// set old msg sender to answer msg  , we will reply answer to sender
	msg.Sender = old.Sender

	return false
}

type bridgeMsgData struct {
	Id     string
	data   *eth.BridgeMsgPacket
	Sender enode.ID
}

func (b *bridgeMsgData) IsExpiration() bool {
	ti := time.Unix(int64(b.data.Expiration), 0)
	return ti.Before(time.Now())
}

func newBridgeMsgData(msg *eth.BridgeMsgPacket, sender enode.ID) *bridgeMsgData {
	data := &bridgeMsgData{
		Id:   msg.Id,
		data: msg.Copy(),
		//Sender: sender,
	}

	if msg.BridgeType == eth.Ask {
		data.Sender = sender
	}

	return data
}

func (h *handler) IsInWhitelist(address common.Address) bool {
	log.Debug("check in whitelist", "address", address.String())
	currentBlock := h.chain.CurrentBlock()
	vmConfig := h.chain.GetVMConfig()
	context := core.NewEVMBlockContext(currentBlock.Header(), h.chain, nil)
	db, err := h.chain.State()
	if err != nil {
		log.Warn("Handler check whitelist get State DB has err", "err", err.Error())
		return false
	}
	evm := vm.NewEVM(context, vm.TxContext{}, db, h.chain.Config(), *vmConfig)

	return evm.StateDB.IsInWhitelist(evm, address)
}

// validateMsg check route sign and msg sign
func (h *handler) validateMsg(msg *eth.BridgeMsgPacket) error {
	addr, err := msg.RouteValidate()
	if err != nil {
		return err
	}
	log.Debug("Validate bridge message route", "id", msg.Id, "address", addr.String())
	//whiteAccts := h.chain.WhiteListAccounts()
	//log.Debug("Bridge handle white", "whiteAccounts", len(white))
	//_, ok := whiteAccts[addr]
	// msg route signer must be white account
	if !h.IsInWhitelist(addr) {
		log.Debug("bridge message signer not white account", "id", msg.Id)
		return errors.New("route validate failed")
	}

	// if it's my answer ,sign msg
	if msg.BridgeType == eth.Answer && msg.Target == h.localNodeId {
		return msg.MsgSign(h.noddeKey)
	} else {
		return msg.MsgValidate()
	}
}

// SendBridgeMsg is handler bridge message , check message type ,
// if ask myself , return true ,ask other node , send message
// if answer ,send to message sender.
// if no target node connect , broadcast this message
func (h *handler) SendBridgeMsg(msg *eth.BridgeMsgPacket, peer *eth.Peer) bool {
	log.Debug("Handler send bridge msg", "id", msg.Id, "source", msg.Source.String(), "target", msg.Target.String(), "type", msg.BridgeType)
	if err := h.validateMsg(msg); err != nil {
		log.Warn("Validate bridge msg has err", "id", msg.Id, "err", err.Error())
		return false
	}

	data := newBridgeMsgData(msg, peer.Node().ID())
	// check message ,  discard repeat message
	if h.bridgeMses.checkDiscard(data) {
		log.Debug("Discard bridge msg", "id", data.Id)
		return false
	}
	switch {
	case data.data.BridgeType == eth.Ask:
		// if ask me , handle message
		if data.data.Target == h.localNodeId {
			log.Debug("Bridge msg ask to me", "id", data.Id)
			return true
		} else {
			h.sendAskBridgeMsg(data)
			return false
		}
	case data.data.BridgeType == eth.Answer:
		h.sendAnswerBridgeMsg(data)
		return false
	}

	return false

}

// sendAskBridgeMsg  send ask message to target node ,
// if no way ,broadcast message
func (h *handler) sendAskBridgeMsg(msg *bridgeMsgData) {
	log.Debug("Send ask bridge msg", "id", msg.Id, "target", msg.data.Target.String())
	data := msg.data
	data.Relay++

	targetPeer := h.peers.peer(msg.data.Target.String())
	if targetPeer != nil {
		log.Debug("Has target peer ,send to target node", "id", msg.Id)
		err := targetPeer.SendBrBridgeMsg(rand.Uint64(), data)
		if err == nil {
			return
		}
	}

	log.Debug("No target node or send failed, broadcast bridge msg ", "id", msg.Id)
	for _, peer := range h.peers.peers {
		peer := peer
		if peer.Node().ID() == msg.Sender {
			// this peer is sender , continue
			continue
		}
		// broadcast message
		go func() {
			log.Debug("Broadcast ask bridge msg to node", "id", msg.Id, "node", peer.ID())
			peer.SendBrBridgeMsg(rand.Uint64(), data)
		}()
	}

	return
}

// sendAnswerBridgeMsg  send answer message to source node ,
// if no way ,broadcast message
func (h *handler) sendAnswerBridgeMsg(msg *bridgeMsgData) {
	log.Debug("Send answer bridge msg", "id", msg.Id, "source", msg.data.Source.String())
	data := msg.data
	targetPeer := h.peers.peer(msg.data.Source.String())
	if targetPeer != nil {
		log.Debug("Has source node connect", "id", msg.Id)
		err := targetPeer.SendBrBridgeMsg(rand.Uint64(), data)
		if err == nil {
			return
		}
	}

	// no source node connect or send failed , send to message sender
	if msg.Sender != msg.data.Source {
		senderPeer := h.peers.peer(msg.Sender.String())
		if senderPeer != nil {
			log.Debug("Answer to sender peer", "id", msg.Id, "sender", msg.Sender.String())
			err := senderPeer.SendBrBridgeMsg(rand.Uint64(), data)
			if err == nil {
				return
			}
		}
	}
	log.Debug("No source, no sender, broadcast answer msg", "id", msg.Id)
	for _, peer := range h.peers.peers {
		peer := peer
		go func() {
			log.Debug("Broadcast answer bridge msg to peer", "id", msg.Id, "node", peer.ID())
			peer.SendBrBridgeMsg(rand.Uint64(), data)
		}()
	}
	return
}
