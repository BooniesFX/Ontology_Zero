/*
 * Copyright (C) 2018 Onchain <onchain@onchain.com>
 *
 * This file is part of The ontology_Zero.
 *
 * The ontology_Zero is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology_Zero is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology_Zero.  If not, see <http://www.gnu.org/licenses/>.
 */

package dbft

import (
	"fmt"
	cl "github.com/Ontology/account"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/log"
	ser "github.com/Ontology/common/serialization"
	"github.com/Ontology/core/ledger"
	tx "github.com/Ontology/core/transaction"
	"github.com/Ontology/crypto"
	"github.com/Ontology/net"
	msg "github.com/Ontology/net/message"
	"sync"
)

const ContextVersion uint32 = 0

type ConsensusContext struct {
	State               ConsensusState
	PrevHash            Uint256
	Height              uint32
	ViewNumber          byte
	BookKeepers         []*crypto.PubKey
	NextBookKeepers     []*crypto.PubKey
	Owner               *crypto.PubKey
	BookKeeperIndex     int
	PrimaryIndex        uint32
	Timestamp           uint32
	Nonce               uint64
	NextBookKeeper      Uint160
	Transactions        []*tx.Transaction
	Signatures          [][]byte
	ExpectedView        []byte

	header              *ledger.Block

	contextMu           sync.Mutex
	isBookKeeperChanged bool
	nmChangedblkHeight  uint32
}

func (cxt *ConsensusContext) M() int {
	log.Debug()
	return len(cxt.BookKeepers) - (len(cxt.BookKeepers) - 1) / 3
}

func NewConsensusContext() *ConsensusContext {
	log.Debug()
	return &ConsensusContext{}
}

func (cxt *ConsensusContext) ChangeView(viewNum byte) {
	log.Debug()
	p := (cxt.Height - uint32(viewNum)) % uint32(len(cxt.BookKeepers))
	cxt.State &= SignatureSent
	cxt.ViewNumber = viewNum
	if p >= 0 {
		cxt.PrimaryIndex = uint32(p)
	} else {
		cxt.PrimaryIndex = uint32(p) + uint32(len(cxt.BookKeepers))
	}

	if cxt.State == Initial {
		cxt.Transactions = nil
		cxt.Signatures = make([][]byte, len(cxt.BookKeepers))
		cxt.header = nil
	}
}

func (cxt *ConsensusContext) MakeChangeView() *msg.ConsensusPayload {
	log.Debug()
	cv := &ChangeView{
		NewViewNumber: cxt.ExpectedView[cxt.BookKeeperIndex],
	}
	cv.msgData.Type = ChangeViewMsg
	return cxt.MakePayload(cv)
}

func (cxt *ConsensusContext) MakeHeader() *ledger.Block {
	log.Debug()
	if cxt.Transactions == nil {
		return nil
	}
	if cxt.header == nil {
		txHash := []Uint256{}
		for _, t := range cxt.Transactions {
			txHash = append(txHash, t.Hash())
		}
		txRoot, err := crypto.ComputeRoot(txHash)
		if err != nil {
			return nil
		}
		blockRoot := ledger.DefaultLedger.Store.GetBlockRootWithNewTxRoot(txRoot)
		header := &ledger.Header{
			Version:          ContextVersion,
			PrevBlockHash:    cxt.PrevHash,
			TransactionsRoot: txRoot,
			BlockRoot:        blockRoot,
			Timestamp:        cxt.Timestamp,
			Height:           cxt.Height,
			ConsensusData:    cxt.Nonce,
			NextBookKeeper:   cxt.NextBookKeeper,
		}
		cxt.header = &ledger.Block{
			Header:    header,
			Transactions: []*tx.Transaction{},
		}
	}
	return cxt.header
}

func (cxt *ConsensusContext) MakePayload(message ConsensusMessage) *msg.ConsensusPayload {
	log.Debug()
	message.ConsensusMessageData().ViewNumber = cxt.ViewNumber
	return &msg.ConsensusPayload{
		Version:         ContextVersion,
		PrevHash:        cxt.PrevHash,
		Height:          cxt.Height,
		BookKeeperIndex: uint16(cxt.BookKeeperIndex),
		Timestamp:       cxt.Timestamp,
		Data:            ser.ToArray(message),
		Owner:           cxt.Owner,
	}
}

func (cxt *ConsensusContext) MakePrepareRequest() *msg.ConsensusPayload {
	log.Debug()
	preReq := &PrepareRequest{
		Nonce:          cxt.Nonce,
		NextBookKeeper: cxt.NextBookKeeper,
		Transactions:   cxt.Transactions,
		Signature:      cxt.Signatures[cxt.BookKeeperIndex],
	}
	preReq.msgData.Type = PrepareRequestMsg
	return cxt.MakePayload(preReq)
}

func (cxt *ConsensusContext) MakePrepareResponse(signature []byte) *msg.ConsensusPayload {
	log.Debug()
	preRes := &PrepareResponse{
		Signature: signature,
	}
	preRes.msgData.Type = PrepareResponseMsg
	return cxt.MakePayload(preRes)
}

func (cxt *ConsensusContext) GetSignaturesCount() (count int) {
	log.Debug()
	count = 0
	for _, sig := range cxt.Signatures {
		if sig != nil {
			count += 1
		}
	}
	return count
}

func (cxt *ConsensusContext) GetStateDetail() string {

	return fmt.Sprintf("Initial: %t, Primary: %t, Backup: %t, RequestSent: %t, RequestReceived: %t, SignatureSent: %t, BlockGenerated: %t, ",
		cxt.State.HasFlag(Initial),
		cxt.State.HasFlag(Primary),
		cxt.State.HasFlag(Backup),
		cxt.State.HasFlag(RequestSent),
		cxt.State.HasFlag(RequestReceived),
		cxt.State.HasFlag(SignatureSent),
		cxt.State.HasFlag(BlockGenerated))

}

func (cxt *ConsensusContext) Reset(client cl.Client, localNode net.Neter) {
	log.Debug()
	cxt.State = Initial
	cxt.PrevHash = ledger.DefaultLedger.Blockchain.CurrentBlockHash()
	cxt.Height = ledger.DefaultLedger.Blockchain.BlockHeight + 1
	cxt.ViewNumber = 0
	cxt.BookKeeperIndex = -1

	cxt.BookKeepers, cxt.NextBookKeepers, _ = ledger.DefaultLedger.Store.GetBookKeeperList()
	log.Info("curr bookkeeper, len:", len(cxt.BookKeepers))
	log.Info("next bookkeeper, len:", len(cxt.NextBookKeepers))

	var err error
	cxt.NextBookKeeper, err = ledger.GetBookKeeperAddress(cxt.NextBookKeepers)
	if err != nil {
		log.Error("[ConsensusContext] GetBookKeeperAddres failed")
	}

	cxt.Owner = cxt.BookKeepers[0]

	bookKeeperLen := len(cxt.BookKeepers)
	cxt.PrimaryIndex = cxt.Height % uint32(bookKeeperLen)
	cxt.Transactions = nil
	cxt.header = nil
	cxt.Signatures = make([][]byte, bookKeeperLen)
	cxt.ExpectedView = make([]byte, bookKeeperLen)

	for i := 0; i < bookKeeperLen; i++ {
		ac, _ := client.GetDefaultAccount()
		if ac.PublicKey.X.Cmp(cxt.BookKeepers[i].X) == 0 {
			cxt.BookKeeperIndex = i
			cxt.Owner = cxt.BookKeepers[i]
			break
		}
	}

}
