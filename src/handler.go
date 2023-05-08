package src

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"os/exec"
	"strings"
)

func KubeadmHandler(c echo.Context) error {
	u := new(KubeadmConfig)
	if err := c.Bind(u); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	//sudo kubeadm init --kubernetes-version "v1.21.0" --pod-network-cidr "10.244.0.0/16" --control-plane-endpoint "xxx.xxx.xxx.xxx:6443" --upload-certs
	command := &exec.Cmd{}
	var out bytes.Buffer
	if u.ControlPlaneEndpoint != "" {
		command = exec.Command("sudo", "kubeadm", "init", "--kubernetes-version", u.K8sVersion,
			"--control-plane-endpoint", u.ControlPlaneEndpoint, "--pod-network-cidr", u.PodNetworkCidr, "--upload-certs")
	}
	command = exec.Command("sudo", "kubeadm", "init", "--kubernetes-version", u.K8sVersion, "--pod-network-cidr", u.PodNetworkCidr, "--upload-certs")
	command.Stdout = &out
	if err := command.Run(); err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("script output: %s \nerr: %s", out.String(), err))
	}
	return c.String(http.StatusOK, out.String())
}

func HostCMDHandler(c echo.Context) error {
	u := new(HostCMD)
	if err := c.Bind(u); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	command := &exec.Cmd{}
	var out bytes.Buffer
	CMDStrArr := strings.Split(u.CMD, " ")
	command = exec.Command(CMDStrArr[0], CMDStrArr[1:]...)
	command.Stdout = &out
	if err := command.Run(); err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("script output: %s \nerr: %s", out.String(), err))
	}
	return c.String(http.StatusOK, out.String())
}
