package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"rosen-bridge/tss-api/logger"
	"rosen-bridge/tss-api/models"
)

type Connection interface {
	Publish(message models.GossipMessage) error
	Subscribe(port string) error
	CallBack(string, interface{}, string) error
	GetPeerId() (string, error)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type connect struct {
	publishUrl      string
	subscriptionUrl string
	getPeerIDUrl    string
	Client          HTTPClient
}

var logging *zap.SugaredLogger

func InitConnection(publishPath string, subscriptionPath string, p2pPort string, getPeerIDPath string) Connection {
	publishUrl := fmt.Sprintf("http://localhost:%s%s", p2pPort, publishPath)
	subscriptionUrl := fmt.Sprintf("http://localhost:%s%s", p2pPort, subscriptionPath)
	getPeerIDUrl := fmt.Sprintf("http://localhost:%s%s", p2pPort, getPeerIDPath)
	logging = logger.NewSugar("connection")
	return &connect{
		publishUrl:      publishUrl,
		subscriptionUrl: subscriptionUrl,
		getPeerIDUrl:    getPeerIDUrl,
		Client:          &http.Client{},
	}

}

//	publishes a message to p2p
func (c *connect) Publish(msg models.GossipMessage) error {
	logging.Infof("publishing new message on p2p")
	marshalledMessage, _ := json.Marshal(&msg)

	type message struct {
		Message  string `json:"message"`
		Channel  string `json:"channel"`
		Receiver string `json:"receiver"`
	}

	values := message{
		Message:  string(marshalledMessage),
		Channel:  "tss",
		Receiver: msg.ReceiverId,
	}
	jsonData, err := json.Marshal(values)
	if err != nil {
		logging.Error(err)
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.publishUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		logging.Errorf("error occurred in creating request: %+v", err)
		return err
	}
	req.Header.Add("content-type", "application/json")
	resp, err := c.Client.Do(req)
	if err != nil {
		logging.Errorf("error occurred in doing request: %+v", err)
		return err
	}
	type response struct {
		Message string `json:"message"`
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("not ok response code: {%d}", resp.StatusCode)
		logging.Error(err)
		return err
	}

	var res = response{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		logging.Error(err)
		return err
	}
	if res.Message != "ok" {
		err = fmt.Errorf("not ok response message: {%s}", res.Message)
		logging.Error(err)
		return err
	}

	logging.Infof("new {%s} message published", msg.Name)
	logging.Debugf("message: %+v", msg.Message)

	return nil
}

//	Subscribe to p2p at first
func (c *connect) Subscribe(port string) error {
	logging.Infof("Subscribing to: %s", c.subscriptionUrl)
	values := map[string]string{
		"channel": "tss",
		"url":     fmt.Sprintf("http://localhost:%s/message", port),
	}
	jsonData, err := json.Marshal(values)
	if err != nil {
		logging.Error(err)
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.subscriptionUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		logging.Error(err)
		return err
	}
	req.Header.Add("content-type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		logging.Error(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("not ok response code: {%v}", resp.StatusCode)
	}

	type response struct {
		Message string `json:"message"`
	}
	var res = response{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		logging.Error(err)
		return err
	}
	if res.Message != "ok" {
		err = fmt.Errorf("not ok response message: {%s}", res.Message)
		logging.Error(err)
		return err
	}

	return nil
}

//	sends sign data to this url
func (c *connect) CallBack(url string, data interface{}, status string) error {
	logging.Info("sending callback data")

	response := struct {
		Message interface{} `json:"message"`
		Status  string      `json:"status"`
	}{
		Message: data,
		Status:  status,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		logging.Error(err)
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		logging.Error(err)
		return err
	}
	req.Header.Add("content-type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		logging.Error(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("not ok response code: {%v}", resp.StatusCode)
		logging.Error(err)
		return err
	}
	return nil
}

//	to get p2pId
func (c *connect) GetPeerId() (string, error) {
	logging.Infof("Getting PeerId")

	req, err := http.NewRequest(http.MethodGet, c.getPeerIDUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("content-type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		logging.Error(err)
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("not ok response: {%v}", resp.StatusCode)
		logging.Error(err)
		return "", err
	}

	type response struct {
		Status string `json:"status"`
		PeerId string `json:"message"`
	}
	var res = response{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		logging.Error(err)
		return "", err
	}
	if res.Status != "ok" {
		err = fmt.Errorf("not ok response message: {%s}", res.Status)
		logging.Error(err)
		return "", err
	}
	if res.PeerId == "" {
		return "", fmt.Errorf("nil peerId")
	}
	logging.Infof("peerId: %+v", res.PeerId)
	return res.PeerId, nil
}
