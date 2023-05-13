package install

import (
	"k8s-installer/src"
	"log"
	"sync"
)

func Installer(cfg *src.Config, configDir string) {
	var err error
	var wg sync.WaitGroup
	hs := HandlerStruct{}
	if err = hs.getConfig(cfg, configDir); err != nil {
		log.Fatal("Config file initialize fail")
	}
	log.Println("#### [1/5] SCP k8s_setup Script ####")
	log.Println("SCP task now processing..")
	if hs.SCPK8sSetup(&wg) {
		log.Println("SCP k8s_setup to all nodes success\n")
	} else {
		// log.Fatal call os.Exit(1)
		log.Fatal("SCP k8s_setup to some node fail")
	}

	log.Println("#### [2/5] Execute Initial Install Script ####")
	log.Println("Executing task now processing..")
	if hs.ExecuteK8sSetup(&wg) {
		log.Println("Execute k8s_setup Script to all nodes success\n")
	} else {
		// log.Fatal call os.Exit(1)
		log.Fatal("Execute k8s_setup Script to some node fail")
	}

	log.Println("#### [3/5] Kubeadm Init Start  ####")
	log.Println("kubeadm init from [", hs.Cfg.Masters[0], "] start..")
	kdmMasterJoinCMD, kdmWorkerJoinCMD := hs.KubeadmInit()
	if kdmMasterJoinCMD == "" && kdmWorkerJoinCMD == "" {
		log.Fatal("Kubeadm init Fail")
	}
	log.Println("kubeadm init from [", hs.Cfg.Masters[0], "] end\n")

	log.Println("#### [4/5] Kubeadm join Start  ####")
	log.Println("kubeadm join Start..")
	if kdmMasterJoinCMD != "" {
		log.Println("Now Master Nodes Join Start")
		log.Println("Master Join CMD: ", kdmMasterJoinCMD)
		if hs.KubeadmJoin(&wg, hs.Cfg.Masters[1:], kdmMasterJoinCMD) {
			log.Println("All Master Join success")
		} else {
			log.Fatal("Some Master Join fail")
		}
	}
	log.Println("Now Worker Nodes Join Start")
	log.Println("Worker Join CMD: ", kdmWorkerJoinCMD)
	if hs.KubeadmJoin(&wg, hs.Cfg.Workers, kdmWorkerJoinCMD) {
		log.Println("All Worker Join success")
	} else {
		log.Fatal("Some Worker Join fail")
	}
	log.Println("Kubeadm join end\n")

	log.Println("#### [5/5] run config/*_bash_script.sh  ####")
	log.Println("Bash script from [", hs.Cfg.Masters[0], "] execute..")
	hs.ExecuteBashScript()
	log.Println("Bash script from [", hs.Cfg.Masters[0], "] complete")
	log.Println("kubernetes clustering finish!")
}
