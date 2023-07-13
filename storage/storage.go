package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/bnb-chain/tss-lib/tss"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"rosen-bridge/tss-api/logger"
	"rosen-bridge/tss-api/models"
)

type Storage interface {
	MakefilePath(peerHome string, protocol string)
	LoadEDDSAKeygen(peerHome string) (models.TssConfigEDDSA, *tss.PartyID, error)
}

type storage struct {
	filePath string
}

var logging *zap.SugaredLogger

//	Constructor of a storage struct
func NewStorage() Storage {
	logging = logger.NewSugar("storage")
	return &storage{
		filePath: "",
	}
}

//	Constructor of a storage struct
func (f *storage) MakefilePath(peerHome string, protocol string) {
	f.filePath = fmt.Sprintf("%s/%s", peerHome, protocol)
}

//	Loads the EDDSA keygen data from the file
func (f *storage) LoadEDDSAKeygen(peerHome string) (models.TssConfigEDDSA, *tss.PartyID, error) {
	// locating file
	var keygenFile string

	f.MakefilePath(peerHome, "eddsa")
	files, err := ioutil.ReadDir(f.filePath)
	if err != nil {
		return models.TssConfigEDDSA{}, nil, err
	}
	if len(files) == 0 {
		return models.TssConfigEDDSA{}, nil, errors.New(models.NoKeygenDataFoundError)
	}
	for _, File := range files {
		if strings.Contains(File.Name(), "keygen") {
			keygenFile = File.Name()
		}
	}
	filePath := filepath.Join(f.filePath, keygenFile)
	logging.Infof("key file path: %v", filePath)

	// reading file
	bz, err := ioutil.ReadFile(filePath)
	if err != nil {
		return models.TssConfigEDDSA{}, nil, errors.Wrapf(
			err,
			"could not open the file for party in the expected location: %s. run keygen first.", filePath,
		)
	}
	var tssConfig models.TssConfigEDDSA
	if err = json.Unmarshal(bz, &tssConfig); err != nil {
		return models.TssConfigEDDSA{}, nil, errors.Wrapf(
			err,
			"could not unmarshal data for party located at: %s", filePath,
		)
	}

	//creating data from file
	for _, kbxj := range tssConfig.KeygenData.BigXj {
		kbxj.SetCurve(tss.Edwards())
	}
	tssConfig.KeygenData.EDDSAPub.SetCurve(tss.Edwards())
	id := xid.New()
	pMoniker := fmt.Sprintf("tssPeer/%s", id.String())
	partyID := tss.NewPartyID(id.String(), pMoniker, tssConfig.KeygenData.ShareID)

	var parties tss.UnSortedPartyIDs
	parties = append(parties, partyID)
	sortedPIDs := tss.SortPartyIDs(parties)
	return tssConfig, sortedPIDs[0], nil
}
