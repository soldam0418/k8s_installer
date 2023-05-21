package handler

import (
	"k8s-installer/src"
	"log"
	"sync"
)

func Remover(hs *HandlerStruct, configDir string) {
	var err error
	var wg sync.WaitGroup
	if err = hs.SetHandler(configDir); err != nil {
		log.Fatal("Config file initialize fail")
	}
	log.Println("#### [1/2] SCP k8s_remove Script ####")
	log.Println("SCP task now processing..")
	if hs.SCPK8sScript(&wg, src.K8S_REMOVE_SCRIPT) {
		log.Println("SCP k8s_remove.sh to all nodes success\n")
	} else {
		// log.Fatal call os.Exit(1)
		log.Fatal("SCP k8s_remove.sh to some node fail")
	}

	log.Println("#### [2/2] Execute k8s_remove Script ####")
	log.Println("Executing task now processing..")
	if hs.ExecuteK8sScript(&wg, src.K8S_REMOVE_SCRIPT) {
		log.Println("Execute k8s_remove.sh to all nodes success\n")
	} else {
		// log.Fatal call os.Exit(1)
		log.Fatal("Execute k8s_remove.sh to some node fail")
	}
	log.Println("kubeadm reset & cri & cni & kubelet & kubeadm removing finish!")
}
