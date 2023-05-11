package src

type Config struct {
	User                 string   `yaml:"user"`
	Masters              []string `yaml:"masters"`
	Workers              []string `yaml:"workers"`
	K8sVersion           string   `yaml:"k8s_version"`
	PodNetworkCidr       string   `yaml:"pod_network_cidr"`
	ControlPlaneEndpoint string   `yaml:"control_plane_endpoint"`
}

type KubeadmConfig struct {
	K8sVersion           string `json:"K8sVersion"`
	PodNetworkCidr       string `json:"PodNetworkCidr"`
	ControlPlaneEndpoint string `json:"ControlPlaneEndpoint"`
}

type HostCMD struct {
	CMD string `json:"CMD"`
}
