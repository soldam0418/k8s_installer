package main

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"k8s-installer/src"
	"log"
	"os/exec"
	"strings"
	"sync"
)

func main() {
	// env GOOS=linux GOARCH=amd64 go build -o k8s_installer
	log.Println("#### [1/5] SCP k8s_setup Script ####")
	log.Println("SCP task now processing..")
	var err error
	cfg := &src.Config{}
	// Read Config file. ${pwd}/config.yaml
	buf, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
		return
	}
	err = yaml.Unmarshal(buf, cfg)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
		return
	}
	allNodes := append(cfg.Masters, cfg.Workers...)

	// SCP Agent to nodes & Run Agent(in Parallel) by using Go-routine
	stepIsOK := true
	var wg sync.WaitGroup
	src.SshCMDToAllNodesByChannel(&wg, allNodes,
		fmt.Sprintf("scp %s %s", src.K8S_SETUP_SCRIPT, fmt.Sprintf("%s@nodeip:/home1/irteamsu", cfg.User)), &stepIsOK)
	if stepIsOK {
		log.Println("SCP k8s_setup to all nodes success\n")
	} else {
		log.Println("SCP k8s_setup to some node fail")
	}

	log.Println("#### [2/5] Execute Initial Install Script ####")
	log.Println("Executing task now processing..")
	src.SshCMDToAllNodesByChannel(&wg, allNodes,
		fmt.Sprintf("ssh %s sh %s", fmt.Sprintf("%s@nodeip", cfg.User), fmt.Sprintf("/home1/irteamsu/%s", src.K8S_SETUP_SCRIPT)), &stepIsOK)
	if stepIsOK {
		log.Println("Execute k8s_setup Script to all nodes success\n")
	} else {
		log.Fatal("Execute k8s_setup Script to some node fail")
	}

	log.Println("#### [3/5] Kubeadm Init Start  ####")
	log.Println("kubeadm init from [", cfg.Masters[0], "] start..")
	command := &exec.Cmd{}
	cmd_kubeadm_out := bytes.Buffer{}
	if cfg.ControlPlaneEndpoint != "" {
		command = exec.Command("ssh", fmt.Sprintf("%s@%s", cfg.User, cfg.Masters[0]), "sudo", "kubeadm", "init", "--kubernetes-version", cfg.K8sVersion,
			"--control-plane-endpoint", cfg.ControlPlaneEndpoint, "--pod-network-cidr", cfg.PodNetworkCidr, "--upload-certs")
	} else {
		command = exec.Command("ssh", fmt.Sprintf("%s@%s", cfg.User, cfg.Masters[0]), "sudo", "kubeadm", "init", "--kubernetes-version", cfg.K8sVersion, "--pod-network-cidr", cfg.PodNetworkCidr, "--upload-certs")
	}
	command.Stdout = &cmd_kubeadm_out
	if err := command.Run(); err != nil {
		log.Fatal(err.Error())
		return
	}
	log.Println("kubeadm init from [", cfg.Masters[0], "] end\n")

	log.Println("#### [4/5] Kubeadm join Start  ####")
	log.Println("kubeadm join Start..")
	kdmMasterJoinCMD, kdmWorkerJoinCMD := src.ParsingKubeadmJoinCMD(strings.Split(cmd_kubeadm_out.String(), "\n"))
	if kdmMasterJoinCMD != "" {
		log.Println("Now Master Nodes Join Start")
		log.Println("Master Join CMD: ", kdmMasterJoinCMD)
		src.SshCMDToAllNodesByChannel(&wg, cfg.Masters[1:], fmt.Sprintf("ssh %s sudo %s",
			fmt.Sprintf("%s@nodeip", cfg.User), kdmMasterJoinCMD), &stepIsOK)
		if stepIsOK {
			log.Println("All Master Join success\n")
		} else {
			log.Println("Some Master Join fail")
		}
	}
	log.Println("Now Worker Nodes Join Start")
	log.Println("Worker Join CMD: ", kdmWorkerJoinCMD)
	src.SshCMDToAllNodesByChannel(&wg, cfg.Workers, fmt.Sprintf("ssh %s sudo %s",
		fmt.Sprintf("%s@nodeip", cfg.User), kdmWorkerJoinCMD), &stepIsOK)
	if stepIsOK {
		log.Println("All Worker Join success")
	} else {
		log.Println("Some Worker Join fail")
	}
	log.Println("Kubeadm join end\n")

	log.Println("#### [5/5] run extra_script/*_bash_script.sh  ####")
	log.Println("bash script from [", cfg.Masters[0], "] execute..")
	files, _ := ioutil.ReadDir(src.EXTRA_SCRIPT)
	filesCnt := len(files)
	log.Println(fmt.Sprintf("Total extra script count: %d.", filesCnt))
	err = exec.Command("scp", "-r", src.EXTRA_SCRIPT,
		fmt.Sprintf("%s@%s:/home1/irteamsu", cfg.User, cfg.Masters[0])).Run()
	if err != nil {
		log.Println("fail to scp extra_script directory. err: ", err.Error())
	}
	for i := 0; i < filesCnt; i++ {
		err := exec.Command("ssh", fmt.Sprintf("%s@%s", cfg.User, cfg.Masters[0]),
			"sh", fmt.Sprintf("/home1/irteamsu/%s/%d*", src.EXTRA_SCRIPT, i+1)).Run()
		if err != nil {
			log.Println(fmt.Sprintf("%d_{script} excute fail. ", i+1), "err: ", err.Error())
		} else {
			log.Println(fmt.Sprintf("%d_{script} excute sucess", i+1))
		}
	}
	log.Println("bash script from [", cfg.Masters[0], "] complete")
	log.Println("kubernetes clustering finish!")
}
