package src

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

type Config struct {
	User                 string   `yaml:"user"`
	Masters              []string `yaml:"masters"`
	Workers              []string `yaml:"workers"`
	K8sVersion           string   `yaml:"k8s_version"`
	PodNetworkCidr       string   `yaml:"pod_network_cidr"`
	ControlPlaneEndpoint string   `yaml:"control_plane_endpoint"`
}

func (cfg *Config) GetConfig(configDir string) (err error) {
	if configDir == "" {
		configDir = DEFAULT_CONFIG_DIR_PATH
	}
	if string(configDir[len(configDir)-1]) == "/" {
		configDir = configDir[:len(configDir)-1]
	}
	// Read Config file. ${pwd}/config.yaml
	if buf, err := ioutil.ReadFile(fmt.Sprintf("%s/config.yaml", configDir)); err != nil {
		log.Println(err)
		return err
	} else {
		if err = yaml.Unmarshal(buf, cfg); err != nil {
			log.Println("Unmarshal: %v", err)
			return err
		}
	}
	return nil
}
