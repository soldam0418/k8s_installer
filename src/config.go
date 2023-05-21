package src

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

type Config struct {
	Masters              []string `yaml:"masters"`
	Workers              []string `yaml:"workers"`
	K8sVersion           string   `yaml:"k8s_version"`
	PodNetworkCidr       string   `yaml:"pod_network_cidr"`
	ControlPlaneEndpoint string   `yaml:"control_plane_endpoint"`
}

func (cfg *Config) GetConfig(configDir string) {
	// Read Config file. ${pwd}/config.yaml
	if buf, err := ioutil.ReadFile(fmt.Sprintf("%s/config.yaml", configDir)); err != nil {
		log.Fatal("Fail to Read Config file.", err)
	} else {
		if err = yaml.Unmarshal(buf, cfg); err != nil {
			log.Fatal("Fail to Unmarshal Yaml.", err)
		}
	}
}
