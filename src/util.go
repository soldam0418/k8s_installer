package src

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
)

func ParsingKubeadmJoinCMD(kdmStringsArr []string) (masterJoinCMD string, workerJoinCMD string) {
	parsingKubeadmJoinStr := func(strArr []string) (mergedStr string) {
		for i, str := range strArr {
			if i == 0 {
				mergedStr += strings.ReplaceAll(str, `\`, "")
			} else {
				mergedStr += strings.TrimSpace(strings.ReplaceAll(str, `\`, ""))
			}
		}
		return mergedStr
	}
	for i, v := range kdmStringsArr {
		if strings.Contains(v, "--control-plane") {
			masterJoinCMD = parsingKubeadmJoinStr(kdmStringsArr[i-2 : i+1])
		}

		if strings.Contains(v, "--discovery-token-ca-cert-hash") {
			workerJoinCMD = parsingKubeadmJoinStr(kdmStringsArr[i-1 : i+1])
		}
	}
	return masterJoinCMD, workerJoinCMD
}

func SshCMDToAllNodesByChannel(wg *sync.WaitGroup, nodes []string, cmd string, isOk *bool) {
	tasks := make(chan string)
	for i := 0; i < len(nodes); i++ {
		wg.Add(1)
		node := nodes[i]
		go func(num int, ip string, w *sync.WaitGroup, clusteringStatue *bool) {
			defer w.Done()
			//respBody, err := HttpPost(<-tasks, fmt.Sprintf("http://%s:%s/%s", ip, AGENT_PORT, HOST_CMD_HANDLER_ROUTE))
			command := &exec.Cmd{}
			var out bytes.Buffer
			CMDStr := strings.ReplaceAll(<-tasks, "nodeip", ip)
			CMDStrArr := strings.Split(CMDStr, " ")
			command = exec.Command(CMDStrArr[0], CMDStrArr[1:]...)
			command.Stdout = &out
			if err := command.Run(); err != nil {
				*clusteringStatue = false
				log.Fatal(fmt.Sprintf("[ %s ] fail \n err: %s", ip, err))
				return
			}
			log.Println(fmt.Sprintf("[ %s ] success complete", ip))
		}(i, node, wg, isOk)
	}
	for i := 0; i < len(nodes); i++ {
		tasks <- cmd
	}
	close(tasks)
	wg.Wait()
}
