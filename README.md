# kubenhn: k8s auto installer

## 모듈 소개
* N개의 Instance 들을 kubernetes cluster 환경으로 자동 구축해주는 모듈이다.
* golang(v1.20.3) 으로 개발되었으며 모든 작업(Go-routine 기반)은 병렬적으로 수행된다.
* 본 모듈은 Linux/Amd64 환경에서 Test 되었으며, 실행 가능한 바이너리 파일(kubenhn)과 Config 디렉토리만 있으면 정상 동작한다.
* hcon 환경(JumpServer) / AWS 에서 정상 동작 확인

## 사용법
```
mkdir ~/binary
cd ~/binary
cp -r ~/k8s-installer/config ./
cp ~/k8s-installer/kubenhn ./
kubenhn -h
```

## 구성 요소 상세 설명

<img width="252" alt="image" src="https://github.nhnent.com/storage/user/3570/files/a93465b4-93d4-4be3-88b9-ed0bfe32dfd1">

* config 디렉토리
  * config.yaml
    * 인스턴스 정보, k8s 설정 정보 파일
  * k8s_setup.sh
    * install 시 모든 노드에 실행되는 스크립트 파일
    * k8s 구성에 필요한 패키지 설치 명령어들로 구성
    * 자유롭게 수정 가능
  * k8s_remove.sh
    * remove 시 모든 노드에 실행되는 스크립트 파일
    * cluster reset 및 kubeadm, kubectl, kubelet, docker 바이너리 & 설정 파일들 삭제
    * 자유롭게 수정 가능
  * deploy
    * kubectl이 설정된 노드(master1)에서 실행되는 스크립트 파일들이 들어있는 디렉토리
    * 파일들은 1_{{ 파일 이름 }}, 2_ {{ 파일 이름 }}, 3_{{ 파일 이름 }} 으로 파일 이름 앞에 "숫자_"를 prefix로 붙임
    * 새로운 파일을 추가 하고 싶다면 4_{{ 파일 이름 }} 으로 생성
* kubenhn 바이너리
  * Linux/Amd64에서 동작하는 바이너리 파일
    * kubenhn -h => 옵션 살펴보기
    * kubenhn -f {{ Config File Directory Path }} => 지정한 Config File Directory 를 읽어 동작 (Default: "./config")
    * kubenhn -m {{ kubenhn Execute Mode }} => kubenhn 실행 모드로 install 및 remove 지원 (Default: "install")
    * kubenhn -u {{ UserName of Instances }} => 인스턴스 접속에 필요한 계정 정보 (Default: "irteamsu")
    * kubenhn -i {{ PemKey Path }} => 인스턴스 접속에 필요한 Pem Key 파일 경로 (Default: "")
    * kubenhn -p {{ Password }} => 인스턴스 접속에 필요한 Password (Default: "")
  * 동작 예시
    ```
    # 아래 명령어 실행 시킬 경우, "./config" 디렉토리를 읽어 "irteamsu" 계정으로 "install" 모드로 동작
    kubenhn 
    ```
    ```
    # 아래 명령어 실행 시킬 경우, "/etc/config" 디렉토리를 읽어 "sreteam" 계정 및 "/etc/mypem.key" 으로 "remove" 모드로 동작
    kubenhn -u sreteam -f /etc/myconfig -m remove -i /etc/mypem.key
    ```
## 동작 과정
* Install Mode (Step by Step으로 각 Task를 실행하며 오류 발생 시 Stop / 실패 Instance 정보 및 error log 출력)
  * config 설정 파일 Load
  * (1) 모든 노드에 k8s_setup.sh 복사 (sshpass scp 활용)
  * (2) 모든 노드에 k8s_setup.sh 실행 (sshpass ssh 활용)
  * (3) master 노드 중 1개의 노드에 kubeadm init 수행 및 join 명령어 파싱
  * (4) 나머지 노드에게 kubeadm join 수행
  * (5) Master 노드 중 1개의 노드에 config/deploy 디렉토리 안의 스크립트 파일 실행
* Remove Mode (Step by Step으로 각 Task를 실행하며 오류 발생 시 Stop / 실패 Instance 정보 및 error log 출력)
  * config 설정 파일 Load
  * (1) 모든 노드에 k8s_remove.sh 복사 (sshpass scp 활용)
  * (2) 모든 노드에 k8s_remove.sh 실행 (sshpass ssh 활용)
