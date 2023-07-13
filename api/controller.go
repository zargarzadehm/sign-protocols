package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"rosen-bridge/tss-api/app/interface"
	"rosen-bridge/tss-api/logger"
	"rosen-bridge/tss-api/models"
)

//	Interface of an app controller
type TssController interface {
	Sign() echo.HandlerFunc
	Keygen() echo.HandlerFunc
	Message() echo.HandlerFunc
}

type tssController struct {
	rosenTss _interface.RosenTss
}

type response struct {
	Message string `json:"message"`
}

var logging *zap.SugaredLogger

//	Constructor of an app controller
func NewTssController(rosenTss _interface.RosenTss) TssController {
	logging = logger.NewSugar("controller")
	return &tssController{
		rosenTss: rosenTss,
	}
}

//	check if there is any common operation between forbidden and running ones.
func (tssController *tssController) checkKeygenOperation(crypto string) error {
	forbiddenOperations := []string{crypto + "Sign", crypto + "Regroup"}
	operations := tssController.rosenTss.GetKeygenOperations()
	for _, operation := range operations {
		for _, forbidden := range forbiddenOperations {
			if operation.GetClassName() == forbidden {
				return fmt.Errorf("%s "+models.OperationIsRunningError, forbidden)
			}
		}
	}
	return nil
}

//	check if there is any common operation between forbidden and running ones.
func (tssController *tssController) checkSignOperation(crypto string) error {
	forbiddenOperations := []string{crypto + "Keygen", crypto + "Regroup"}
	operations := tssController.rosenTss.GetSignOperations()
	for _, operation := range operations {
		for _, forbidden := range forbiddenOperations {
			if operation.GetClassName() == forbidden {
				return fmt.Errorf("%s "+models.OperationIsRunningError, forbidden)
			}
		}
	}
	return nil
}

//	check if there is any common operation between forbidden and running ones.
func (tssController *tssController) checkOperation(operationName string, crypto string) error {
	switch operationName {
	case "keygen":
		return tssController.checkKeygenOperation(crypto)
	case "sign":
		return tssController.checkSignOperation(crypto)
	default:
		return fmt.Errorf(models.WrongOperationError)
	}
}

// Keygen returns echo handler, starting new keygen process
func (tssController *tssController) Keygen() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := models.KeygenMessage{}

		if err := c.Bind(&data); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		logging.Debugf("keygen controller called with data: {%v}", data)
		err := tssController.checkOperation("keygen", data.Crypto)
		if err != nil {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		err = tssController.rosenTss.StartNewKeygen(data)
		if err != nil {
			switch err.Error() {
			case models.DuplicatedMessageIdError:
				return echo.NewHTTPError(http.StatusConflict, err.Error())
			case models.KeygenFileExistError, models.WrongCryptoProtocolError:
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			default:
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
		}
		return c.JSON(
			http.StatusOK, response{
				Message: "ok",
			},
		)
	}
}

//	returns echo handler, starting new sign process.
func (tssController *tssController) Sign() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := models.SignMessage{}

		if err := c.Bind(&data); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		logging.Debugf("sign controller called with data: {%v}", data)
		err := tssController.checkOperation("sign", data.Crypto)
		if err != nil {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		err = tssController.rosenTss.StartNewSign(data)
		if err != nil {
			switch err.Error() {
			case models.DuplicatedMessageIdError:
				return echo.NewHTTPError(http.StatusConflict, err.Error())
			case models.NoKeygenDataFoundError, models.WrongCryptoProtocolError:
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			default:
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
		}

		return c.JSON(
			http.StatusOK, response{
				Message: "ok",
			},
		)
	}
}

//	returns echo handler, receiving message from p2p and passing to related channel
func (tssController *tssController) Message() echo.HandlerFunc {
	return func(c echo.Context) error {
		var data models.Message
		if err := c.Bind(&data); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		logging.Debugf("message controller called with data: {%v}", data)
		err := tssController.rosenTss.MessageHandler(data)
		if err != nil {
			logging.Error(err)
		}
		return c.JSON(
			http.StatusOK, response{
				Message: "ok",
			},
		)
	}
}
