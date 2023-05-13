package main

import (
	"flag"
	"k8s-installer/src"
	"k8s-installer/src/install"
	"log"
)

func main() {
	// env GOOS=linux GOARCH=amd64 go build -o kubenhn
	var configFilePath string
	flag.StringVar(&configFilePath, "file", src.DEFAULT_CONFIG_DIR_PATH, "Config Directory Path")
	flag.StringVar(&configFilePath, "f", src.DEFAULT_CONFIG_DIR_PATH, "-f equal --file")

	var executeMode string
	flag.StringVar(&executeMode, "mode", src.INSTALL_MODE, "kubecli execute mode. support [\"install\", \"reset\", \"test\"]")
	flag.StringVar(&executeMode, "m", src.INSTALL_MODE, "-m equal --mode")
	flag.Parse()

	cfg := src.Config{}
	if err := cfg.GetConfig(configFilePath); err != nil {
		log.Fatal(err)
	}
	switch executeMode {
	case src.INSTALL_MODE:
		install.Installer(&cfg, configFilePath)
	case src.RESET_MODE:
		log.Println("reset")
	case src.TEST_MODE:
		log.Println("test")
	default:
		log.Fatal("kubecli support only [\"install\", \"reset\", \"test\"]")
	}
}
