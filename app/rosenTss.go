package app

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/blake2b"
	"rosen-bridge/tss-api/app/interface"
	eddsaSign "rosen-bridge/tss-api/app/sign/eddsa"
	"rosen-bridge/tss-api/logger"
	"rosen-bridge/tss-api/models"
	"rosen-bridge/tss-api/network"
	"rosen-bridge/tss-api/storage"
	"rosen-bridge/tss-api/utils"
)

type rosenTss struct {
	ChannelMap   map[string]chan models.GossipMessage
	OperationMap map[string]_interface.Operation
	metaData     models.MetaData
	storage      storage.Storage
	connection   network.Connection
	Config       models.Config
	peerHome     string
	P2pId        string
}

var logging *zap.SugaredLogger

// NewRosenTss Constructor of an app
func NewRosenTss(connection network.Connection, storage storage.Storage, config models.Config) _interface.RosenTss {
	logging = logger.NewSugar("app")
	return &rosenTss{
		ChannelMap:   make(map[string]chan models.GossipMessage),
		OperationMap: make(map[string]_interface.Operation),
		metaData:     models.MetaData{},
		storage:      storage,
		connection:   connection,
		Config:       config,
	}
}

func (r *rosenTss) errorCallBackCall(signMessage models.SignMessage, err error) {
	data := struct {
		Err string `json:"error"`
		M   string `json:"m"`
	}{
		Err: err.Error(),
		M:   signMessage.Message,
	}
	callbackErr := r.GetConnection().CallBack(signMessage.CallBackUrl, data, "error")
	if callbackErr != nil {
		logging.Error(callbackErr)
	}
}

// StartNewSign starts sign scenario for app based on given protocol.
func (r *rosenTss) StartNewSign(signMessage models.SignMessage) error {
	logging.Info("Starting New Sign process")
	msgBytes, _ := utils.Decoder(signMessage.Message)
	signData := new(big.Int).SetBytes(msgBytes)
	signDataBytes := blake2b.Sum256(signData.Bytes())
	signDataHash := utils.Encoder(signDataBytes[:])
	logging.Infof("encoded sign data: %v", signDataHash)

	messageId := fmt.Sprintf("%s%s", signMessage.Crypto, signDataHash)
	_, ok := r.ChannelMap[messageId]
	if !ok {
		messageCh := make(chan models.GossipMessage, 100)
		r.ChannelMap[messageId] = messageCh
		logging.Infof("new communication channel for signning process: %v", messageId)
	} else {
		return fmt.Errorf(models.DuplicatedMessageIdError)
	}

	var operation _interface.Operation
	println(signMessage.Crypto)
	switch signMessage.Crypto {
	case "ecdsa":
		operation = eddsaSign.NewSignEDDSAOperation(signMessage)
	case "eddsa":
		operation = eddsaSign.NewSignEDDSAOperation(signMessage)
	default:
		return fmt.Errorf(models.WrongCryptoProtocolError)
	}
	channelId := fmt.Sprintf("%s%s", operation.GetClassName(), messageId)
	r.OperationMap[channelId] = operation
	errorCh := make(chan error)

	go func() {
		timeout := time.After(time.Second * time.Duration(r.Config.OperationTimeout))
		for {
			select {
			case <-timeout:
				if _, ok := r.ChannelMap[messageId]; ok {
					close(r.ChannelMap[messageId])
					err := fmt.Errorf("sign operation timeout")
					errorCh <- err
				}
				return
			}
		}
	}()

	err := operation.Init(r, signMessage.Peers)
	if err != nil {
		return err
	}
	go func() {
		logging.Infof("calling start action for %s sign", signMessage.Crypto)
		err = operation.StartAction(r, r.ChannelMap[messageId], errorCh)
		if err != nil {
			logging.Errorf("en error occurred in %s sign action, err: %+v", signMessage.Crypto, err)
			r.errorCallBackCall(signMessage, err)
		}
		r.deleteInstance(messageId, channelId, errorCh)
		logging.Infof("end of %s sign action", signMessage.Crypto)
	}()

	return nil
}

// MessageHandler handles the receiving message from message route
func (r *rosenTss) MessageHandler(message models.Message) error {

	msgBytes := []byte(message.Message)
	gossipMsg := models.GossipMessage{}
	err := json.Unmarshal(msgBytes, &gossipMsg)
	if err != nil {
		return err
	}

	logging.Infof("callback route called. new %+v message from: %+v", gossipMsg.Name, gossipMsg.SenderId)

	// handling recover in case the channel is closed but not removed from the list yet, and there is a message to send on that
	send := func(c chan models.GossipMessage, t models.GossipMessage) {
		defer func() {
			if x := recover(); x != nil {
				logging.Warnf("unable to send: %v", x)
			}
		}()
		c <- t
	}

	var state bool
	for i, start := 0, time.Now(); ; i++ {
		if time.Since(start) > time.Second*time.Duration(r.Config.MessageTimeout) {
			state = false
			break
		}
		if _, ok := r.ChannelMap[gossipMsg.MessageId]; ok {
			send(r.ChannelMap[gossipMsg.MessageId], gossipMsg)
			state = true
			break
		}
		// TODO: add config for this time
		time.Sleep(time.Millisecond * 500)
	}
	if !state {
		logging.Warnf("message timeout, channel not found: %+v", gossipMsg.MessageId)
		return nil
	} else {
		return nil
	}
}

// GetStorage returns the storage
func (r *rosenTss) GetStorage() storage.Storage {
	return r.storage
}

// GetConnection returns the connection
func (r *rosenTss) GetConnection() network.Connection {
	return r.connection
}

//SetPeerHome setups peer home address and creates that
func (r *rosenTss) SetPeerHome(homeAddress string) error {
	logging.Info("setting up home directory")

	absAddress, err := utils.GetAbsoluteAddress(homeAddress)
	if err != nil {
		return err
	}
	r.peerHome = absAddress

	if err := os.MkdirAll(r.peerHome, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// GetPeerHome returns the peer's home
func (r *rosenTss) GetPeerHome() string {
	return r.peerHome
}

// SetMetaData setting ups metadata from given file in the home directory
func (r *rosenTss) SetMetaData(meta models.MetaData) error {
	r.metaData = meta
	return nil
}

// GetMetaData returns peer's meta data
func (r *rosenTss) GetMetaData() models.MetaData {
	return r.metaData
}

// GetOperations returns list of operations
func (r *rosenTss) GetOperations() map[string]_interface.Operation {
	return r.OperationMap
}

// deleteInstance removes operation and related channel from list
func (r *rosenTss) deleteInstance(messageId string, channelId string, errorCh chan error) {
	delete(r.OperationMap, channelId)
	delete(r.ChannelMap, messageId)
	close(errorCh)
}

// SetP2pId set p2p to the variable
func (r *rosenTss) SetP2pId() error {
	p2pId, err := r.GetConnection().GetPeerId()
	if err != nil {
		return err
	}
	r.P2pId = p2pId
	return nil
}

// GetP2pId get p2pId
func (r *rosenTss) GetP2pId() string {
	return r.P2pId
}

// GetConfig get Config
func (r *rosenTss) GetConfig() models.Config {
	return r.Config
}
