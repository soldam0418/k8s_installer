package install

import (
	"fmt"
	"io/ioutil"
	"k8s-installer/src"
	"log"
	"os/exec"
	"strings"
	"sync"
)

type HandlerStruct struct {
	Cfg         *src.Config
	CfgDir      string
	AllNodes    []string
	HostBaseDir string
}

func (hs *HandlerStruct) getConfig(cfg *src.Config, configDir string) (err error) {
	hs.CfgDir = configDir
	if string(hs.CfgDir[len(hs.CfgDir)-1]) == "/" {
		hs.CfgDir = hs.CfgDir[:len(hs.CfgDir)-1]
	}
	hs.Cfg = cfg
	hs.AllNodes = append(hs.Cfg.Masters, hs.Cfg.Workers...)
	// Get Base Path. ex) baseDir := "/home1/irteamsu"
	log.Println("Get Instance Base Path")
	hs.HostBaseDir = strings.ReplaceAll(src.SshCMDToGetOutput(fmt.Sprintf("ssh %s@%s pwd", hs.Cfg.User, hs.Cfg.Masters[0])), "\n", "")
	return nil
}

func (hs *HandlerStruct) SCPK8sSetup(wg *sync.WaitGroup) (isOk bool) {
	stepIsOK := true
	src.SshCMDToAllNodesByChannel(wg, hs.AllNodes,
		fmt.Sprintf("scp %s %s", fmt.Sprintf("%s/k8s_setup.sh", hs.CfgDir),
			fmt.Sprintf("%s@nodeip:%s", hs.Cfg.User, hs.HostBaseDir)), &stepIsOK)
	if !stepIsOK {
		return false
	}
	return true
}

func (hs *HandlerStruct) ExecuteK8sSetup(wg *sync.WaitGroup) (isOk bool) {
	stepIsOK := true
	src.SshCMDToAllNodesByChannel(wg, hs.AllNodes,
		fmt.Sprintf("ssh %s sh %s", fmt.Sprintf("%s@nodeip", hs.Cfg.User),
			fmt.Sprintf("%s/%s", hs.HostBaseDir, "k8s_setup.sh")), &stepIsOK)
	if !stepIsOK {
		return false
	}
	return true
}

func (hs *HandlerStruct) KubeadmInit() (kdmMasterJoinCMD string, kdmWorkerJoinCMD string) {
	kdmJoinCMD := fmt.Sprintf("ssh %s sudo kubeadm init --kubernetes-version %s --pod-network-cidr %s --upload-certs",
		fmt.Sprintf("%s@%s", hs.Cfg.User, hs.Cfg.Masters[0]), hs.Cfg.K8sVersion, hs.Cfg.PodNetworkCidr)
	if hs.Cfg.ControlPlaneEndpoint != "" {
		kdmJoinCMD = kdmJoinCMD + " " + fmt.Sprintf("--control-plane-endpoint %s", hs.Cfg.ControlPlaneEndpoint)
	}
	kdmJoinCMDStr := src.SshCMDToGetOutput(kdmJoinCMD)
	return src.ParsingKubeadmJoinCMD(strings.Split(kdmJoinCMDStr, "\n"))
}

func (hs *HandlerStruct) KubeadmJoin(wg *sync.WaitGroup, nodes []string, kdmJoinCMD string) bool {
	stepIsOK := true
	src.SshCMDToAllNodesByChannel(wg, nodes, fmt.Sprintf("ssh %s sudo %s",
		fmt.Sprintf("%s@nodeip", hs.Cfg.User), kdmJoinCMD), &stepIsOK)
	return stepIsOK
}

func (hs *HandlerStruct) ExecuteBashScript() {
	deployDir := hs.CfgDir + "/deploy"
	files, _ := ioutil.ReadDir(deployDir)
	filesCnt := len(files)
	err := exec.Command("scp", "-r", deployDir,
		fmt.Sprintf("%s@%s:%s", hs.Cfg.User, hs.Cfg.Masters[0], hs.HostBaseDir)).Run()
	if err != nil {
		log.Println("fail to scp config directory. err: ", err.Error())
	}
	for i := 0; i < filesCnt; i++ {
		err := exec.Command("ssh", fmt.Sprintf("%s@%s", hs.Cfg.User, hs.Cfg.Masters[0]),
			"sh", fmt.Sprintf("%s/%s/%d*", hs.HostBaseDir, "deploy", i+1)).Run()
		if err != nil {
			log.Println(fmt.Sprintf("[ %s ] excute fail. ", files[i].Name()), "err: ", err.Error())
		} else {
			log.Println(fmt.Sprintf("[ %s ] excute sucess", files[i].Name()))
		}
	}
}
