# 사전작업필요.

# golang 설치
wget https://go.dev/dl/go1.23.10.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.10.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
source ~/.bashrc
go version

# docker 설치
sudo apt-get remove docker docker-engine docker.io containerd runc    # 기존 삭제
sudo apt-get update
sudo apt-get install \
  ca-certificates curl gnupg lsb-release
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/docker-archive-keyring.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker-archive-keyring.gpg] \
   https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo usermod -aG docker ubuntu # 사용자이름 작성

# mysql 설치
docker run --name my-mysql \
  -e MYSQL_ROOT_PASSWORD=password123 \
  -e MYSQL_DATABASE=mydb \
  -e MYSQL_USER=testuser \
  -e MYSQL_PASSWORD=password123 \
  -p 3306:3306 \
  -d mysql
  
테스트 테이블 생성
CREATE DATABASE IF NOT EXISTS test;
USE test;
CREATE TABLE IF NOT EXISTS users (
    UserID     VARCHAR(50)  NOT NULL PRIMARY KEY,
    UserPW     VARCHAR(100) NOT NULL,
    status     VARCHAR(20)  DEFAULT 'active',
    LastLogin  DATETIME     DEFAULT NULL
);


# Redis 설치
docker run -d --name my_redis -p 6379:6379 redis



