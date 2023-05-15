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
	flag.StringVar(&configFilePath, "f", src.DEFAULT_CONFIG_DIR_PATH, "Config Directory Path.")

	var executeMode string
	flag.StringVar(&executeMode, "m", src.INSTALL_MODE, "kubenhn execute mode. support [\"install\", \"reset\", \"test\"]")
	flag.Parse()

	var userName string
	flag.StringVar(&userName, "u", src.USER, "Instance access user name.")
	flag.Parse()

	var pemPath string
	flag.StringVar(&pemPath, "i", "", "PemKey Path.")
	flag.Parse()

	var password string
	flag.StringVar(&password, "p", "", "Instance access password")
	flag.Parse()
	cfg := src.Config{User: userName, PemPath: pemPath, Password: password}
	if err := cfg.GetConfig(configFilePath); err != nil {
		log.Fatal(err)
	}

	if pemPath != "" && password != "" {
		log.Fatal("Input pemPath or password. Do not input both.")
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
