package main

import (
	"flag"
	"k8s-installer/src"
	"k8s-installer/src/handler"
	"log"
)

func main() {
	// Host(Linux) Build Command:
	// env GOOS=linux GOARCH=amd64 go build -o kubenhn
	var configFilePath string
	flag.StringVar(&configFilePath, "f", src.DEFAULT_CONFIG_DIR_PATH, "Config Directory Path.")

	var executeMode string
	flag.StringVar(&executeMode, "m", src.INSTALL_MODE, "kubenhn execute mode. support [\"install\", \"remove\"]")

	var userName string
	flag.StringVar(&userName, "u", src.USER, "Instance access user name.")

	var pemPath string
	flag.StringVar(&pemPath, "i", "", "PemKey Path.")

	var password string
	flag.StringVar(&password, "p", "", "Instance access password")
	flag.Parse()

	if pemPath != "" && password != "" {
		log.Fatal("Input pemPath or password. Do not input both.")
	}

	hs := &handler.HandlerStruct{User: userName, PemPath: pemPath, Password: password}

	switch executeMode {
	case src.INSTALL_MODE:
		handler.Installer(hs, configFilePath)
	case src.REMOVE_MODE:
		handler.Remover(hs, configFilePath)
	default:
		log.Fatal("kubecli support only [\"install\", \"remove\"]")
	}
}
