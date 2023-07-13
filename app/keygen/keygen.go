package keygen

import (
	"fmt"
	eddsaKeygen "github.com/bnb-chain/tss-lib/eddsa/keygen"
	"github.com/bnb-chain/tss-lib/tss"
	"go.uber.org/zap"
	_interface "rosen-bridge/tss-api/app/interface"
	"rosen-bridge/tss-api/models"
)

const (
	KeygenFileName = "keygen_data.json"
)

type Handler interface {
	StartParty(
		localTssData *models.TssData,
		threshold int,
		outCh chan tss.Message,
		endCh chan eddsaKeygen.LocalPartySaveData,
	) error
}

type StructKeygen struct {
	_interface.KeygenOperationHandler
	LocalTssData  models.TssData
	KeygenMessage models.KeygenMessage
	Logger        *zap.SugaredLogger
	Handler
}

//	- Updates party on received message destination.
func (s *StructKeygen) PartyUpdate(partyMsg models.PartyMessage) error {
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

	err := s.KeygenOperationHandler.SharedPartyUpdater(s.LocalTssData.Party, partyMsg)
	if err != nil {
		return err
	}
	return nil
}
