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

// TssController Interface of an app controller
type TssController interface {
	Sign() echo.HandlerFunc
	Message() echo.HandlerFunc
}

type tssController struct {
	rosenTss _interface.RosenTss
}

type response struct {
	Message string `json:"message"`
}

var logging *zap.SugaredLogger

// NewTssController Constructor of an app controller
func NewTssController(rosenTss _interface.RosenTss) TssController {
	logging = logger.NewSugar("controller")
	return &tssController{
		rosenTss: rosenTss,
	}
}

// checkOperation check if there is any common between forbidden list of requested operation and running operations
func (tssController *tssController) checkOperation(forbiddenOperations []string) error {
	operations := tssController.rosenTss.GetOperations()
	for _, operation := range operations {
		for _, forbidden := range forbiddenOperations {
			if operation.GetClassName() == forbidden {
				return fmt.Errorf("%s "+models.OperationIsRunningError, forbidden)
			}
		}
	}
	return nil
}

// Sign returns echo handler, starting new sign process.
func (tssController *tssController) Sign() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := models.SignMessage{}

		if err := c.Bind(&data); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		logging.Infof("sign data: %+v ", data)

		forbiddenOperations := []string{data.Crypto + "Keygen", data.Crypto + "Regroup"}
		err := tssController.checkOperation(forbiddenOperations)
		if err != nil {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		err = tssController.rosenTss.StartNewSign(data)
		// TODO: should delete instance of operation and call back
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

//Message returns echo handler, receiving message from p2p and passing to related channel
func (tssController *tssController) Message() echo.HandlerFunc {
	return func(c echo.Context) error {
		var data models.Message
		logging.Infof("message route called")
		if err := c.Bind(&data); err != nil {
			logging.Errorf("can not bind data, err: %+v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

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
