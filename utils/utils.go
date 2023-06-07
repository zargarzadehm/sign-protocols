package utils

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"rosen-bridge/tss-api/models"
)

func GetAbsoluteAddress(address string) (string, error) {
	var absAddress string
	switch address[0:1] {
	case ".":
		addr, err := filepath.Abs(address)
		if err != nil {
			return "", err
		}
		absAddress = addr
	case "~":
		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		addr := filepath.Join(userHome, address[1:])
		absAddress = addr
	case "/":
		absAddress = address
	default:
		return "", fmt.Errorf("wrong address format: %s", address)
	}
	return absAddress, nil
}

//	reads in config file and ENV variables if set.
func InitConfig(configFile string) (models.Config, error) {
	// Search config in home directory with name "default" (without extension).
	viper.SetConfigFile(configFile)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		return models.Config{}, fmt.Errorf("error using config file: %s", err.Error())
	}
	conf := models.Config{}
	err = viper.Unmarshal(&conf)
	if err != nil {
		return models.Config{}, fmt.Errorf("error Unmarshalling config file: %s", err.Error())
	}
	return conf, nil
}

//	finds index of element in a slice of bigInt
func IndexOf(collection []*big.Int, el *big.Int) int {
	for i, x := range collection {
		if x.Cmp(el) == 0 {
			return i
		}
	}
	return -1
}

func Decoder(message string) ([]byte, error) {
	return hex.DecodeString(message)
}

func Encoder(message []byte) string {
	return hex.EncodeToString(message)
}
