package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bnb-chain/tss-lib/v2/tss"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"rosen-bridge/tss-api/logger"
	"rosen-bridge/tss-api/models"
)

type Storage interface {
	MakefilePath(peerHome string, protocol string)
	WriteData(data interface{}, peerHome string, fileFormat string, protocol string) error
	LoadEDDSAKeygen(peerHome string) (models.TssConfigEDDSA, *tss.PartyID, error)
	LoadECDSAKeygen(peerHome string) (models.TssConfigECDSA, *tss.PartyID, error)
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

// WriteData writing given data to file in given path
func (f *storage) WriteData(data interface{}, peerHome string, fileFormat string, protocol string) error {

	logging.Info("writing data to the file")

	f.MakefilePath(peerHome, protocol)
	err := os.MkdirAll(f.filePath, os.ModePerm)
	if err != nil {
		return err
	}

	path := filepath.Join(f.filePath, fileFormat)

	logging.Infof("file path: %s", path)
	fd, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	defer func(fd *os.File) {
		err := fd.Close()
		if err != nil {
			logging.Errorf("unable to close file %s, err:{%v}", path, err)
		}
	}(fd)

	if err != nil {
		return fmt.Errorf("unable to open File %s for writing, err:{%v}", path, err)
	}
	bz, err := json.MarshalIndent(&data, "", "    ")
	if err != nil {
		return fmt.Errorf("unable to marshal data, err:{%v}", err)
	}
	_, err = fd.Write(bz)
	if err != nil {
		return fmt.Errorf("unable to write to File %s", path)
	}
	logging.Infof("data was written successfully in a file: %s", path)
	return nil
}

//	Loads the EDDSA keygen data from the file
func (f *storage) LoadEDDSAKeygen(peerHome string) (models.TssConfigEDDSA, *tss.PartyID, error) {
	// locating file
	var keygenFile string

	f.MakefilePath(peerHome, models.EDDSA)
	files, err := ioutil.ReadDir(f.filePath)
	if err != nil || len(files) == 0 {
		logging.Warnf("couldn't find eddsa keygen %v", err)
		return models.TssConfigEDDSA{}, nil, errors.New(models.EDDSANoKeygenDataFoundError)
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

//	Loads the ECDSA keygen data from the file
func (f *storage) LoadECDSAKeygen(peerHome string) (models.TssConfigECDSA, *tss.PartyID, error) {
	// locating file
	var keygenFile string

	f.MakefilePath(peerHome, models.ECDSA)
	files, err := ioutil.ReadDir(f.filePath)
	if err != nil || len(files) == 0 {
		logging.Warnf("couldn't find ecdsa keygen %v", err)
		return models.TssConfigECDSA{}, nil, errors.New(models.ECDSANoKeygenDataFoundError)
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
		return models.TssConfigECDSA{}, nil, errors.Wrapf(
			err,
			"could not open the file for party in the expected location: %s. run keygen first.", filePath,
		)
	}
	var tssConfig models.TssConfigECDSA
	if err = json.Unmarshal(bz, &tssConfig); err != nil {
		return models.TssConfigECDSA{}, nil, errors.Wrapf(
			err,
			"could not unmarshal data for party located at: %s", filePath,
		)
	}

	//creating data from file
	for _, kbxj := range tssConfig.KeygenData.BigXj {
		kbxj.SetCurve(tss.S256())
	}
	tssConfig.KeygenData.ECDSAPub.SetCurve(tss.S256())
	id := xid.New()
	pMoniker := fmt.Sprintf("tssPeer/%s", id.String())
	partyID := tss.NewPartyID(id.String(), pMoniker, tssConfig.KeygenData.ShareID)

	var parties tss.UnSortedPartyIDs
	parties = append(parties, partyID)
	sortedPIDs := tss.SortPartyIDs(parties)
	return tssConfig, sortedPIDs[0], nil
}
