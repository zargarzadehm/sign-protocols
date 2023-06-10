package _interface

import (
	"rosen-bridge/tss-api/models"
	"rosen-bridge/tss-api/network"
	"rosen-bridge/tss-api/storage"
)

//	(sign eddsa protocol)
type Operation interface {
	Init(RosenTss, []models.Peer) error
	StartAction(RosenTss, chan models.GossipMessage, chan error) error
	GetClassName() string
}

//	Interface of an app
type RosenTss interface {
	StartNewSign(models.SignMessage) error
	MessageHandler(models.Message) error

	GetStorage() storage.Storage
	GetConnection() network.Connection

	SetMetaData(data models.MetaData) error
	GetMetaData() models.MetaData

	SetPeerHome(string) error
	GetPeerHome() string

	GetOperations() map[string]Operation

	SetP2pId() error
	GetP2pId() string
	GetConfig() models.Config
}
