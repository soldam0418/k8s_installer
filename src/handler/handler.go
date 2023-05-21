package handler

import (
	"fmt"
	"io/ioutil"
	"k8s-installer/src"
	"log"
	"strings"
	"sync"
)

type HandlerStruct struct {
	Cfg         *src.Config
	User        string
	Password    string
	PemPath     string
	CfgDir      string
	AllNodes    []string
	HostBaseDir string
	SshCMD      string
	ScpCMD      string
}

func (hs *HandlerStruct) SetHandler(configDir string) (err error) {
	if string(configDir[len(configDir)-1]) == "/" {
		configDir = configDir[:len(configDir)-1]
	}
	hs.CfgDir = configDir
	hs.Cfg = &src.Config{}
	hs.Cfg.GetConfig(hs.CfgDir)
	hs.AllNodes = append(hs.Cfg.Masters, hs.Cfg.Workers...)
	hs.SshCMD = "ssh -o StrictHostKeyChecking=no"
	hs.ScpCMD = "scp -o StrictHostKeyChecking=no"
	if hs.Password != "" {
		hs.SshCMD = fmt.Sprintf("sshpass -p %s %s", hs.Password, hs.SshCMD)
		hs.ScpCMD = fmt.Sprintf("sshpass -p %s %s", hs.Password, hs.ScpCMD)
	}
	if hs.PemPath != "" {
		hs.SshCMD = fmt.Sprintf("sshpass %s -i %s", hs.SshCMD, hs.PemPath)
		hs.ScpCMD = fmt.Sprintf("sshpass %s -i %s", hs.ScpCMD, hs.PemPath)
	}
	// Get Base Path. ex) baseDir := "/home1/irteamsu"
	log.Println("Get Instance Base Path")
	hs.HostBaseDir = strings.ReplaceAll(src.SshCMDToGetOutput(
		fmt.Sprintf("%s %s@%s pwd", hs.SshCMD, hs.User, hs.Cfg.Masters[0])), "\n", "")
	return nil
}

/** Installer Method */
// Step 1. SCP k8s_Setup.sh Files to All Nodes
func (hs *HandlerStruct) SCPK8sScript(wg *sync.WaitGroup, script string) (isOk bool) {
	stepIsOK := true
	// scp ~/config/k8s_setup.sh {user}@{ip}:{home directory}
	src.SshCMDToAllNodesByChannel(wg, hs.AllNodes,
		fmt.Sprintf("%s %s %s", hs.ScpCMD, fmt.Sprintf("%s/%s", hs.CfgDir, script),
			fmt.Sprintf("%s@nodeip:%s", hs.User, hs.HostBaseDir)), &stepIsOK)
	if !stepIsOK {
		return false
	}
	return true
}

// Step 2. Execute k8s_Setup.sh Files to All Nodes
func (hs *HandlerStruct) ExecuteK8sScript(wg *sync.WaitGroup, script string) (isOk bool) {
	stepIsOK := true
	// ssh {user}@{ip} sh ~/k8s_setup.sh
	src.SshCMDToAllNodesByChannel(wg, hs.AllNodes,
		fmt.Sprintf("%s %s sh %s", hs.SshCMD, fmt.Sprintf("%s@nodeip", hs.User),
			fmt.Sprintf("%s/%s", hs.HostBaseDir, script)), &stepIsOK)
	if !stepIsOK {
		return false
	}
	return true
}

// Step 3. Kubeadm init from first master node
func (hs *HandlerStruct) KubeadmInit() (kdmMasterJoinCMD string, kdmWorkerJoinCMD string) {
	// ssh {user}@{ip} sudo kubeadm init
	kdmJoinCMD := fmt.Sprintf("%s %s sudo kubeadm init --kubernetes-version %s --pod-network-cidr %s --upload-certs",
		hs.SshCMD, fmt.Sprintf("%s@%s", hs.User, hs.Cfg.Masters[0]), hs.Cfg.K8sVersion, hs.Cfg.PodNetworkCidr)
	if hs.Cfg.ControlPlaneEndpoint != "" {
		kdmJoinCMD = kdmJoinCMD + " " + fmt.Sprintf("--control-plane-endpoint %s", hs.Cfg.ControlPlaneEndpoint)
	}
	kdmJoinCMDStr := src.SshCMDToGetOutput(kdmJoinCMD)
	return src.ParsingKubeadmJoinCMD(strings.Split(kdmJoinCMDStr, "\n"))
}

// Step 3. Kubeadm join from other nodes
func (hs *HandlerStruct) KubeadmJoin(wg *sync.WaitGroup, nodes []string, kdmJoinCMD string) bool {
	stepIsOK := true
	// ssh {user}@{ip} sudo kubeadm join ~
	src.SshCMDToAllNodesByChannel(wg, nodes, fmt.Sprintf("%s %s sudo %s",
		hs.SshCMD, fmt.Sprintf("%s@nodeip", hs.User), kdmJoinCMD), &stepIsOK)
	return stepIsOK
}

// Step 4. SCP & Execute bash script from first master node
func (hs *HandlerStruct) ExecuteBashScript() {
	deployDir := hs.CfgDir + "/deploy"
	files, _ := ioutil.ReadDir(deployDir)
	filesCnt := len(files)
	// scp -r ~/config/deploy {user}@{ip}:{home directory}
	scpCommand := src.ParsingCommand(fmt.Sprintf("%s -r %s %s", hs.ScpCMD, deployDir,
		fmt.Sprintf("%s@%s:%s", hs.User, hs.Cfg.Masters[0], hs.HostBaseDir)))
	if err := scpCommand.Run(); err != nil {
		log.Println("fail to scp config directory. err: ", err.Error())
	}
	for i := 0; i < filesCnt; i++ {
		// ssh {user}@{ip} sh ~/deploy/1*
		// ssh {user}@{ip} sh ~/deploy/2*
		// ssh {user}@{ip} sh ~/deploy/3*
		// ...
		sshCommand := src.ParsingCommand(fmt.Sprintf("%s %s sh %s", hs.SshCMD,
			fmt.Sprintf("%s@%s", hs.User, hs.Cfg.Masters[0]), fmt.Sprintf("%s/%s/%d*", hs.HostBaseDir, "deploy", i+1)))
		if err := sshCommand.Run(); err != nil {
			log.Println(fmt.Sprintf("[ %s ] excute fail. ", files[i].Name()), "err: ", err.Error())
		} else {
			log.Println(fmt.Sprintf("[ %s ] excute sucess", files[i].Name()))
		}
	}
}

/** Remover Method */
// Step 1. SCP k8s_remove.sh Files to All Nodes
func (hs *HandlerStruct) SCPK8sRemove(wg *sync.WaitGroup) (isOk bool) {
	stepIsOK := true
	// scp ~/config/k8s_setup.sh {user}@{ip}:{home directory}
	src.SshCMDToAllNodesByChannel(wg, hs.AllNodes,
		fmt.Sprintf("%s %s %s", hs.ScpCMD, fmt.Sprintf("%s/k8s_remove.sh", hs.CfgDir),
			fmt.Sprintf("%s@nodeip:%s", hs.User, hs.HostBaseDir)), &stepIsOK)
	if !stepIsOK {
		return false
	}
	return true
}

// Step 2. Execute k8s_remove.sh Files to All Nodes
func (hs *HandlerStruct) ExecuteK8sRemove(wg *sync.WaitGroup) (isOk bool) {
	stepIsOK := true
	// ssh {user}@{ip} sh ~/k8s_setup.sh
	src.SshCMDToAllNodesByChannel(wg, hs.AllNodes,
		fmt.Sprintf("%s %s sh %s", hs.SshCMD, fmt.Sprintf("%s@nodeip", hs.User),
			fmt.Sprintf("%s/%s", hs.HostBaseDir, "k8s_remove.sh")), &stepIsOK)
	if !stepIsOK {
		return false
	}
	return true
}
