package sign

import (
	"fmt"
	"github.com/bnb-chain/tss-lib/v2/common"
	"github.com/bnb-chain/tss-lib/v2/tss"
	"go.uber.org/zap"
	"math/big"
	_interface "rosen-bridge/tss-api/app/interface"
	"rosen-bridge/tss-api/models"
)

type Handler interface {
	LoadData(_interface.RosenTss) (*tss.PartyID, error)
	GetData() ([]*big.Int, *big.Int)
	StartParty(
		localTssData *models.TssData,
		threshold int,
		signData []byte,
		outCh chan tss.Message,
		endCh chan *common.SignatureData,
	) error
}

type StructSign struct {
	_interface.SignOperationHandler
	LocalTssData         models.TssData
	SignMessage          models.SignMessage
	Signatures           map[int][]byte
	Logger               *zap.SugaredLogger
	SetupSignMessage     models.SetupSign
	SelfSetupSignMessage models.SetupSign
	Handler
}

//	- Updates party on received message destination.
func (s *StructSign) PartyUpdate(partyMsg models.PartyMessage) error {
	dest := partyMsg.GetTo
	if dest == nil { // broadcast!
		if s.LocalTssData.Party.PartyID().Index == partyMsg.GetFrom.Index {
			return nil
		}
		s.Logger.Infof("updating party state with bradcast message")
	} else { // point-to-point!
		if dest[0].Index == partyMsg.GetFrom.Index {
			err := fmt.Errorf("party %d tried to send a message to itself (%d)", dest[0].Index, partyMsg.GetFrom.Index)
			return err
		}
		s.Logger.Infof("updating party state with p2p message")
	}

	err := s.SignOperationHandler.SharedPartyUpdater(s.LocalTssData.Party, partyMsg)
	if err != nil {
		return err
	}
	return nil
}
