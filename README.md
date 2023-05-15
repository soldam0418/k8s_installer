# kubenhn: k8s auto installer

## 모듈 소개
* N개의 Instance 들을 kubernetes cluster 환경으로 자동 구축해주는 모듈이다.
* golang(v1.20.3) 으로 개발되었으며 모든 작업(Go-routine 기반)은 병렬적으로 수행된다.
* 실행 가능한 바이너리 파일(Linux/Amd64)과 Config 디렉토리로 구성되어 있으며 hcon 환경에서 정상 구동된다.

## 동작 원리
* hcon 환경에서 config 디텍토리를 읽어들여 다음과 같은 작업을 수행한다. 
  * config.yaml 파일 Read
  * 모든 노드에 k8s_setup.sh 복사(sshpass scp 활용)
  * 1개의 노드에 kubeadm init 수행
  * 나머지 노드에게 kubeadm join 수행
  * Master 노드(1개)에 config/deploy 디렉토리 안의 스크립트 파일 실행 

## 시퀀스 다이어 그램
![image](https://github.nhnent.com/storage/user/3570/files/57329000-ee71-11ed-85f8-64b01ac9ca20)

## 구성 요소 상세 설명
<img width="437" alt="image" src="https://github.nhnent.com/storage/user/3570/files/b393c42c-aef7-42bf-a87b-b0160304565b">
* kubenhn 바이너리
  * Linux/Amd64에서 동작하는 바이너리 파일
  * ./kubenhn -h => 옵션 살펴보기
  * ./kubenhn -f {{ Config File Directory Path }} => 지정한 Config File Directory 를 읽어 동작
  * ./kubenhn => ./config 디렉토리를 읽어 동작 (Default 설정 값)
* config 디렉토리
  * config.yaml
    * 인스턴스 정보, k8s 설정 정보 파일
  * k8s_setup.sh
    * 모든 노드에 실행되는 스크립트 파일
    * k8s 구성에 필요한 패키지 설치 명령어들로 구성
    * 자유롭게 수정 가능
  * deploy
    * 스크립트 파일들이 들어있는 디렉토리
    * 파일들은 1_{{ 파일 이름 }}, 2_ {{ 파일 이름 }}, 3_{{ 파일 이름 }} 으로 파일 이름 앞에 "숫자_"를 prefix로 붙임
    * 새로운 파일을 추가 하고 싶다면 4_{{ 파일 이름 }} 으로 생성
    * kubectl이 설정된 노드에서 extra_script 안에 있는 모든 스크립트 파일 자동 실행
