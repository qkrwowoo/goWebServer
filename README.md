# 개요

Go & gRPC 를 활용한 RestAPI WebServer 기본 구조 구축 소스파일
 - 기본 구조를 참조하여 다양한 서비스 확장이 가능하도록 구성
 - RestAPI 요청을 받아 gRPC 서버에 등록된 함수를 호출
 - MySQL, MsSQL, Oracle, Redis 기능 구현
 - 환경파일(goWeb.ini, grpcWeb.ini)을 활용한 데이터베이스 Destination 설정

```markdown
## 디렉토리 구조
project-root/
├── goWeb/ # Rest API Web 서버
│ ├── goWeb.go
│ ├── goWeb.ini
│ └── ...
├── grpcWeb/ # gRPC 서버
│ ├── grpcWeb.go
│ ├── grpcWeb.ini
│ └── ...
├── common/ # 공통 라이브러리
│ ├── config.go
│ ├── log.go
│ └── ...
└── 
```
## 실행 예시
```bash
./goWeb/goWeb -f ./goWeb/goWeb.ini
./grpcWeb/grpcWeb -f ./grpcWeb/grpcWeb.ini
```


# 사전작업
### 1. Golang 설치

```bash
wget https://go.dev/dl/go1.23.10.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.10.linux-amd64.tar.gz

echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

go version
```

### 2. Docker 설치
```
# 필수 패키지 설치
sudo apt-get update
sudo apt-get install -y \
  ca-certificates curl gnupg lsb-release

# Docker GPG 키 추가 및 저장소 등록
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/docker-archive-keyring.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker-archive-keyring.gpg] \
   https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Docker 설치
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 현재 사용자 docker 그룹에 추가
sudo usermod -aG docker $USER
newgrp docker
```

### 3. MySQL 설치
``` bash
docker run --name my-mysql \
  -e MYSQL_ROOT_PASSWORD=password123 \
  -e MYSQL_DATABASE=test \
  -e MYSQL_USER=testuser \
  -e MYSQL_PASSWORD=password123 \
  -p 3306:3306 \
  -d mysql

# 테스트 테이블 생성
CREATE DATABASE IF NOT EXISTS test;
USE test;

CREATE TABLE IF NOT EXISTS users (
    UserID     VARCHAR(50)  NOT NULL PRIMARY KEY,
    UserPW     VARCHAR(100) NOT NULL,
    status     VARCHAR(20)  DEFAULT 'active',
    LastLogin  DATETIME     DEFAULT NULL
);

```

### 4. Redis 설치
``` bash
docker run -d --name my_redis -p 6379:6379 redis
```
