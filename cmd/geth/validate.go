package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"net"
)

func NewValidateInfo(key *ecdsa.PrivateKey, ipAddress string, port int) (*ValidateInfo, error) {
	node := enode.NewV4(&key.PublicKey, net.ParseIP(ipAddress), port, 0)
	v := &ValidateInfo{
		NodeId:  node.ID(),
		Address: node.String(),
	}

	err := v.Sign(key)
	if err != nil {
		return nil, err
	}

	return v, nil

}

type ValidateInfo struct {
	NodeId enode.ID `json:"nodeId"`
	//Nonce   string `json:"key"`
	Address   string `json:"address"`
	Signature string `json:"signature"`
}

func (v *ValidateInfo) Sign(priv *ecdsa.PrivateKey) error {

	digest := crypto.Keccak256(v.NodeId.Bytes())
	sig, err := crypto.Sign(digest, priv)
	if err != nil {
		return errors.New("sign failed")
	}
	v.Signature = hexutil.Encode(sig)
	return nil
}

func (v *ValidateInfo) Validate() bool {
	node, err := enode.Parse(enode.ValidSchemes, v.Address)
	if err != nil {
		return false
	}
	if v.NodeId != node.ID() {
		return false
	}
	sig, err := hexutil.Decode(v.Signature)
	if err != nil {
		return false
	}
	puk, err := crypto.SigToPub(crypto.Keccak256(v.NodeId.Bytes()), sig)
	if err != nil {
		return false
	}

	return enode.PubkeyToIDV4(puk) == v.NodeId
}

func (v *ValidateInfo) ToJson() string {
	jb, _ := json.MarshalIndent(v, "", "\t")
	return string(jb)
}
