package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"k8s-installer/src"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func main() {
	// env GOOS=linux GOARCH=amd64 go build -o k8s_installer
	serverOrAgent := os.Args[1]
	switch serverOrAgent {
	case "server":
		log.Println("#### [1/4] SCP Agent to nodes & Run Agent ####")
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
		// if the agent already installed and running, then terminate it
		src.TerminateAgent(allNodes)

		// SCP Agent to nodes & Run Agent(in Parallel) by using Go-routine
		stepIsOK := true
		tasks := make(chan []*exec.Cmd)
		var wg sync.WaitGroup
		for i := 0; i < len(allNodes); i++ {
			wg.Add(1)
			go func(node string, w *sync.WaitGroup) {
				defer w.Done()
				for cmd := range tasks {
					log.Println(fmt.Sprintf("[ %s ] ", node), "scp copy start..")
					if err := cmd[0].Run(); err != nil {
						log.Println(fmt.Sprintf("[ %s ] ", node), "scp copy fail", err)
						stepIsOK = false
						return
					} else {
						log.Println(fmt.Sprintf("[ %s ] ", node), "scp copy end :")
					}
					log.Println(fmt.Sprintf("[ %s ] ", node), "install k8s set up start..")
					if err := cmd[1].Run(); err != nil {
						log.Println(fmt.Sprintf("[ %s ] ", node), "install k8s set up fail", err)
						stepIsOK = false
						return
					} else {
						log.Println(fmt.Sprintf("[ %s ] ", node), "install k8s set up end")
					}
					log.Println(fmt.Sprintf("[ %s ] ", node), "agent start..")
					if err := cmd[2].Start(); err != nil {
						log.Println(fmt.Sprintf("[ %s ] ", node), "agent start fail", err)
						stepIsOK = false
						return
					} else {
						log.Println(fmt.Sprintf("[ %s ] ", node), "agent is now running")
					}
				}
			}(allNodes[i], &wg)
		}
		for _, node := range allNodes {
			tasks <- []*exec.Cmd{
				exec.Command("sshpass", "-p", cfg.Password, "scp", "-o", "StrictHostKeyChecking=no", src.K8S_INSTALLER, src.K8S_SETUP_SCRIPT, fmt.Sprintf("%s@%s:/home", cfg.User, node)),
				exec.Command("sshpass", "-p", cfg.Password, "ssh", "-o", "StrictHostKeyChecking=no", fmt.Sprintf("%s@%s", cfg.User, node), "sudo", "sh", fmt.Sprintf("/home/%s", src.K8S_SETUP_SCRIPT)),
				exec.Command("sshpass", "-p", cfg.Password, "ssh", "-o", "StrictHostKeyChecking=no", fmt.Sprintf("%s@%s", cfg.User, node), "sudo", fmt.Sprintf("/home/%s", src.K8S_INSTALLER), "agent"),
			}
		}
		close(tasks)
		wg.Wait()
		if !stepIsOK {
			log.Fatal("Fail to Setup")
			return
		}
		time.Sleep(5 * time.Second)
		log.Println("SCP Agent to nodes & Run success")

		log.Println("#### [2/4] Kubeadm Init Start  ####")
		log.Println("kubeadm init from [", cfg.Masters[0], "] start..")
		reqJson := src.KubeadmConfig{
			K8sVersion:           cfg.K8sVersion,
			PodNetworkCidr:       cfg.PodNetworkCidr,
			ControlPlaneEndpoint: cfg.ControlPlaneEndpoint,
		}
		respBody, err := src.HttpPost(reqJson, fmt.Sprintf("http://%s:%s/%s", cfg.Masters[0], src.AGENT_PORT, src.KUBEADM_HANDLER_ROUTE))
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Println("kubeadm init from [", cfg.Masters[0], "] end")

		log.Println("#### [3/5] Kubeadm join Start  ####")
		log.Println("kubeadm join Start..")
		kdmMasterJoinCMD, kdmWorkerJoinCMD := src.ParsingKubeadmJoinCMD(strings.Split(respBody, "\n"))
		if kdmMasterJoinCMD != "" {
			log.Println("Master Join CMD: ", kdmMasterJoinCMD)
			src.HttpPostToAllNodesByChannel(&wg, cfg.Masters[1:], kdmMasterJoinCMD, &stepIsOK)
		}
		log.Println("Worker Join CMD: ", kdmWorkerJoinCMD)
		src.HttpPostToAllNodesByChannel(&wg, cfg.Workers, kdmWorkerJoinCMD, &stepIsOK)
		if !stepIsOK {
			log.Fatal("Some Node Fail to Join. Please Check the log.")
		}
		log.Println("Kubeadm join end")

		log.Println("#### [4/5] run extra_script/*_bash_script.sh  ####")
		log.Println("bash script from [", cfg.Masters[0], "] execute..")
		files, _ := ioutil.ReadDir(src.EXTRA_SCRIPT)
		filesCnt := len(files)
		log.Println(fmt.Sprintf("Total extra script count: %d.", filesCnt))
		err = exec.Command("sshpass", "-p", cfg.Password, "scp", "-o", "StrictHostKeyChecking=no", "-r", src.EXTRA_SCRIPT,
			fmt.Sprintf("%s@%s:/home", cfg.User, cfg.Masters[0])).Run()
		if err != nil {
			log.Println("fail to scp. err: ", err.Error())
		}
		for i := 0; i < filesCnt; i++ {
			err := exec.Command("sshpass", "-p", cfg.Password, "ssh", "-o", "StrictHostKeyChecking=no",
				fmt.Sprintf("%s@%s", cfg.User, cfg.Masters[0]), "sudo", "sh", fmt.Sprintf("/home/%s/%d*", src.EXTRA_SCRIPT, i+1)).Run()
			if err != nil {
				log.Println("fail to execute bash script.", fmt.Sprintf("%d_{script} excute fail. ", i+1), "err: ", err.Error())
			} else {
				log.Println(fmt.Sprintf("%d_{script} excute sucess", i+1))
			}
		}
		log.Println("bash script from [", cfg.Masters[0], "] complete")

		log.Println("#### [5/5] Terminate Agent Echo Server ####")
		src.TerminateAgent(allNodes)
		log.Println("#### Kubernetes Clustering Complete! ####")
		break
	case "agent":
		e := echo.New()
		// Middleware
		e.Use(middleware.Logger())
		// Routes
		// Execute kubeadm init or join command
		e.POST("/"+src.KUBEADM_HANDLER_ROUTE, src.KubeadmHandler)
		// Execute host command
		e.POST("/"+src.HOST_CMD_HANDLER_ROUTE, src.HostCMDHandler)
		// Terminate agent (self-killed)
		e.GET("/"+src.TERMINATE_AGENT_ROUTE, func(c echo.Context) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			return e.Shutdown(ctx)
		})
		// Start agent server
		e.Logger.Fatal(e.Start(":" + src.AGENT_PORT))
		break
	default:
		log.Println("Input server or agent.")
		return
	}
}
