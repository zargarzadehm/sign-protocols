package models

import (
	eddsaKeygen "github.com/bnb-chain/tss-lib/eddsa/keygen"
	"github.com/bnb-chain/tss-lib/tss"
)

const (
	DuplicatedMessageIdError = "duplicated messageId"
	OperationIsRunningError  = "operation is running"
	NoKeygenDataFoundError   = "no keygen data found"
	WrongCryptoProtocolError = "wrong crypto protocol"
)

type SignMessage struct {
	Crypto      string `json:"crypto"`
	Message     string `json:"message"`
	CallBackUrl string `json:"callBackUrl"`
	Peers       []Peer `json:"peers"`
}

type Peer struct {
	ShareId string `json:"shareId"`
	P2PId   string `json:"p2pId"`
}

type SignData struct {
	Signature string `json:"signature"`
	R         string `json:"r"`
	S         string `json:"s"`
	M         string `json:"m"`
}

type Message struct {
	Message string `json:"message"`
	Sender  string `json:"sender"`
	Topic   string `json:"channel"`
}

type GossipMessage struct {
	MessageId  string `json:"messageId"`
	Message    string `json:"message"`
	SenderId   string `json:"senderId"`
	ReceiverId string `json:"receiverId"`
	Index      int    `json:"index"`
}

type MetaData struct {
	PeersCount int `json:"peersCount"`
	Threshold  int `json:"threshold"`
}

type TssConfigEDDSA struct {
	MetaData   MetaData                       `json:"metaData"`
	KeygenData eddsaKeygen.LocalPartySaveData `json:"keygenData"`
}

type TssData struct {
	PartyID  *tss.PartyID
	Params   *tss.Parameters
	PartyIds tss.SortedPartyIDs
	Party    tss.Party
}

type PartyMessage struct {
	Message                 []byte
	GetFrom                 *tss.PartyID
	GetTo                   []*tss.PartyID
	IsBroadcast             bool
	IsToOldCommittee        bool
	IsToOldAndNewCommittees bool
}

type Config struct {
	HomeAddress                string  `mapstructure:"HOME_ADDRESS"`
	LogLevel                   string  `mapstructure:"LOG_LEVEL"`
	LogMaxSize                 int     `mapstructure:"LOG_MAX_SIZE"`
	LogMaxBackups              int     `mapstructure:"LOG_MAX_BACKUPS"`
	LogMaxAge                  int     `mapstructure:"LOG_MAX_AGE"`
	OperationTimeout           int     `mapstructure:"OPERATION_TIMEOUT"`
	MessageTimeout             int     `mapstructure:"MESSAGE_TIMEOUT"`
	LeastProcessRemainingTime  int64   `mapstructure:"LEAST_PROCESS_REMAINING_TIME"`
	SetupBroadcastInterval     int64   `mapstructure:"SETUP_BROADCAST_INTERVAL"`
	SignStartTimeTracker       float64 `mapstructure:"SIGN_START_TIME_TRACKER"`
	TurnDuration               int64   `mapstructure:"TRUN_DURATION"`
	WaitInPartyMessageHandling int64   `mapstructure:"WAIT_IN_PARTY_MESSAGE_HANDLING"`
}

type Payload struct {
	MessageId string `json:"messageId"`
	Message   string `json:"message"`
	SenderId  string `json:"senderId"`
}

type SetupSign struct {
	Hash      string        `json:"hash"`
	Peers     []tss.PartyID `json:"peers"`
	Timestamp int64         `json:"timestamp"`
	StarterId string        `json:"starterId"`
}
