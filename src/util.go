package src

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
)

func HttpPost(jsonData any, url string) (string, error) {
	pBytes, _ := json.Marshal(jsonData)
	buff := bytes.NewBuffer(pBytes)
	resp, err := http.Post(url, "application/json", buff)
	if err != nil {
		log.Fatal("Fail to POST http")
		return "", err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return string(respBody), nil
}

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

func HttpPostToAllNodesByChannel(wg *sync.WaitGroup, nodes []string, cmd string, isOk *bool) {
	tasks := make(chan HostCMD)
	for i := 0; i < len(nodes); i++ {
		wg.Add(1)
		worker := nodes[i]
		go func(num int, ip string, w *sync.WaitGroup, clusteringStatue *bool) {
			defer w.Done()
			respBody, err := HttpPost(<-tasks, fmt.Sprintf("http://%s:%s/%s", ip, AGENT_PORT, HOST_CMD_HANDLER_ROUTE))
			if err != nil {
				*clusteringStatue = false
				log.Fatal(fmt.Sprintf("[ %s ] node join fail \n err: %s", ip, err))
				return
			}
			log.Println(respBody)
			log.Println(fmt.Sprintf("[ %s ] node join complete", ip))
		}(i, worker, wg, isOk)
	}
	for i := 0; i < len(nodes); i++ {
		tasks <- HostCMD{CMD: cmd}
	}
	close(tasks)
	wg.Wait()
}

func TerminateAgent(allNodes []string) {
	for _, node := range allNodes {
		terminateAgentCMD := fmt.Sprintf("http://%s:%s/%s", node, AGENT_PORT, TERMINATE_AGENT_ROUTE)
		_, _ = http.Get(terminateAgentCMD)
	}
}
