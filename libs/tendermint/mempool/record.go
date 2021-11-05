package mempool

import (
	"fmt"
	"sync"

	"github.com/okex/exchain/libs/tendermint/libs/log"
	"github.com/okex/exchain/libs/tendermint/p2p"
)

var globalRecord *recorder

type sendStatus struct {
	PeerKey          string
	PeerHeight       int64
	SuccessSendCount int64
	FailSendCount    int64
}

type recorder struct {
	logger        log.Logger
	body          sync.Map
	currentHeight int64
}

func GetGlobalRecord(l log.Logger) *recorder {
	if globalRecord == nil {
		globalRecord = &recorder{
			logger: l,
		}
	}
	globalRecord.logger = l
	return globalRecord
}

func (s *recorder) DoLog() {
	if s.currentHeight == 0 {
		return
	}

	s.logger.Info(fmt.Sprintf("mp broadcast log height :%d, detail : %s", s.currentHeight, s.Detail()))

	s.body = sync.Map{}
	s.currentHeight = 0
}

func (s *recorder) AddPeer(peer p2p.Peer, success bool, txHeight, peerHeight int64) {
	if txHeight > s.currentHeight {
		s.currentHeight = txHeight
	}

	addr, err := peer.NodeInfo().NetAddress()
	if err != nil {
		return
	}
	peerKey := addr.String()

	sendTmp := &sendStatus{
		PeerKey:    peerKey,
		PeerHeight: peerHeight,
	}
	if success {
		sendTmp.SuccessSendCount++
	} else {
		sendTmp.FailSendCount++
	}

	if v, ok := s.body.Load(peerKey); !ok {
		s.body.Store(peerKey, sendTmp)
	} else {
		sendInfo, ok := v.(*sendStatus)
		if !ok {
			return
		}

		sendInfo.PeerHeight = peerHeight
		if success {
			sendInfo.SuccessSendCount++
		} else {
			sendInfo.FailSendCount++
		}
		s.body.Store(peerKey, sendInfo)
	}
}

func (s *recorder) DelPeer(peer p2p.Peer) {
	addr, err := peer.NodeInfo().NetAddress()
	if err != nil {
		return
	}
	peerKey := addr.String()
	s.body.Delete(peerKey)
}

func (s *recorder) Detail() string {
	var res string
	var peersNum, successNum, failedNum int64
	var successRate float64
	s.body.Range(func(k, v interface{}) bool {
		info, ok := v.(*sendStatus)
		if !ok {
			res += "peer sendInfo type wrong"
			return false
		}
		peersNum++
		successNum += info.SuccessSendCount
		failedNum += info.FailSendCount
		return true
	})

	if len(res) != 0 {
		return res
	}
	if successNum+failedNum > 0 {
		successRate = float64(successNum) / float64(successNum+failedNum)
	}
	res = fmt.Sprintf("peersNum : %d, broadTxNum : %d, successRate : %f", peersNum, successNum+failedNum, successRate)
	return res
}
