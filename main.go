package main

import (
	"flag"
	"fmt"

	"github.com/labstack/echo/v4"
	"rosen-bridge/tss-api/api"
	"rosen-bridge/tss-api/app"
	"rosen-bridge/tss-api/logger"
	"rosen-bridge/tss-api/network"
	"rosen-bridge/tss-api/storage"
	"rosen-bridge/tss-api/utils"
)

func main() {

	// parsing cli flags
	projectPort := flag.String("port", "4000", "project port (e.g. 4000)")
	p2pPort := flag.String("p2pPort", "8080", "p2p port (e.g. 8080)")
	publishPath := flag.String(
		"publishPath", "/p2p/send", "publish path of p2p (e.g. /p2p/send)",
	)
	subscriptionPath := flag.String(
		"subscriptionPath", "/p2p/channel/subscribe", "subscriptionPath for p2p (e.g. /p2p/channel/subscribe)",
	)
	getPeerIDPath := flag.String(
		"getP2pIDPath", "/p2p/getPeerID", "getP2pIDPath for p2p (e.g. /p2p/getPeerID)",
	)
	configFile := flag.String(
		"configFile", "./conf/conf.env", "config file",
	)
	flag.Parse()

	config, err := utils.InitConfig(*configFile)
	if err != nil {
		panic(err)
	}

	absAddress, err := utils.GetAbsoluteAddress(config.HomeAddress)
	if err != nil {
		panic(err)
	}
	logFile := fmt.Sprintf("%s/%s", absAddress, "tss.log")

	err = logger.Init(logFile, config, false)
	if err != nil {
		panic(err)
	}
	logging := logger.NewSugar("main")

	defer func() {
		err = logger.Sync()
		if err != nil {
			logging.Error(err)
		}
	}()

	logging.Infof("config: %+v", config)
	// creating new instance of echo framework
	e := echo.New()
	// initiating and reading configs

	// creating connection and storage and app instance
	conn := network.InitConnection(*publishPath, *subscriptionPath, *p2pPort, *getPeerIDPath)
	localStorage := storage.NewStorage()

	tss := app.NewRosenTss(conn, localStorage, config)
	// setting up peer home based on configs
	err = tss.SetPeerHome(config.HomeAddress)
	if err != nil {
		logging.Fatal(err)
	}

	// subscribe to p2p
	err = tss.GetConnection().Subscribe(*projectPort)
	if err != nil {
		logging.Fatal(err)
	}

	// running echo framework
	tssController := api.NewTssController(tss)

	// get p2pId
	err = tss.SetP2pId()
	if err != nil {
		logging.Fatal(err)
	}

	api.InitRouting(e, tssController)
	logging.Fatal(e.Start(fmt.Sprintf(":%s", *projectPort)))
}
