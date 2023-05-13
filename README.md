# k8s installer

## 모듈 소개
* N개의 Instance 들을 kubernetes cluster 환경으로 자동 구축해주는 모듈이다.
* 실행 가능한 바이너리 파일(Linux/Amd64) 또는 Dockerfile 기반 Container 로 운용 가능하다.
* golang(v1.20.3) 으로 개발되었으며 모든 작업(Go-routine 기반)은 병렬적으로 수행된다.

## 사전 준비 사항
* Server는 sshpass 및 openssh 또는 Docker(v19.03 이상)이 사전 설치되어 있어야 한다.
* Agent는 1234 tcp port inbound 를 허용해야 한다.

## 동작 원리
![image](https://github.nhnent.com/storage/user/3570/files/9bbd2c00-ee6f-11ed-9b18-752b0bd1ac9c#center)
* 단일 Server & N개의 Agent 구조로 동작하며, 아래의 과정을 수행한다.
  * config.yaml 파일 Read
  * 모든 노드에 k8s_setup.sh, Agent 배포 및 실행(sshpass scp 활용)
  * 1개의 노드에 kubeadm init 수행
  * 나머지 노드에게 kubeadm join 수행
  * 1개의 노드에 설치된 kubectl을 활용하여 cni, metric server 배포 

## 시퀀스 다이어 그램
![image](https://github.nhnent.com/storage/user/3570/files/57329000-ee71-11ed-85f8-64b01ac9ca20)

## 환경 설정
* k8s-installer 가 run-time 시 활용하는 file 리스트는 아래와 같음
  * [config.yaml](https://github.nhnent.com/srep/k8s-installer/blob/master/config.yaml)
    * 인스턴스 정보, k8s 설정 정보 파일
  * [k8s_setup.sh](https://github.nhnent.com/srep/k8s-installer/blob/master/k8s_setup.sh)
    * 모든 노드에 실행되는 스크립트 파일
    * k8s 구성에 필요한 패키지 설치 명령어들로 구성
    * 자유롭게 수정 가능
  * [k8s_installer](https://github.nhnent.com/srep/k8s-installer/blob/master/k8s_installer)
    * 바이너리 파일
    * "server" 또는 "agent" 를 입력받아 동작
    * k8s_installer server => 서버로 동작 / k8s_installer agent => 에이전트로 동작
  * [extra_script](https://github.nhnent.com/srep/k8s-installer/tree/master/extra_script)
    * 스크립트 파일들이 들어있는 디렉토리
    * 파일들은 1_{{ 파일 이름 }}, 2_ {{ 파일 이름 }}, 3_{{ 파일 이름 }} 으로 파일 이름 앞에 "숫자_"를 prefix로 붙임
    * 새로운 파일을 추가 하고 싶다면 4_{{ 파일 이름 }} 으로 생성
    * kubectl이 설정된 노드에서 extra_script 안에 있는 모든 스크립트 파일 자동 실행
  
## 실행 방법
* 바이너리 파일 실행
  ```bash
  $ k8s_installer server
  ```
* Docker 이미지 빌드 및 컨테이너 실행
  ```bash
  $ docker build -t k8s_container .
  $ docker run --name k8s-container -itd k8s_container
  # Host Volume 마운트하여 활용 가능 (config 및 script 파일 custom 가능)
  $ docker run --name k8s-container -itd -v ./config.yaml:/app/config.yaml \
    -v ./k8s_setup.sh:/app/k8s_setup.sh -v ./config:/app/config \
    k8s_container
  ```
* Go run Command
  ```bash
  $ go run main.go server
  ```

## 참고 사항
* k8s_installer 바이너리 파일
  * 실행 입력에 따라 Server 또는 Agent로 동작
    * Server은 Host 환경에 Docker 또는 sshpass & openssh 패키지가 설치되어 있어야 정상 동작함.
    * Agent는 1234 Port(tcp) 를 활용하며 해당 Port 가 열려있어야함. 
  * OS: linux / Arch: amd64 으로 컴파일됨
  * Ubuntu 20.04 버전 기준 테스트 완료
