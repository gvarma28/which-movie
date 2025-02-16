.PHONY: build-server

include .env
export $(shell sed 's/=.*//' .env)

SERVER_BUILD_PATH = ./build/server

build-server:
	cd server && GOOS=linux GOARCH=amd64 go build -o ${SERVER_BUILD_PATH} 

deploy-server: build-server
	cd server && scp -i $(EC2_KEY_PATH) ${SERVER_BUILD_PATH} $(EC2_USER)@$(EC2_IP):$(REMOTE_PATH)

# ssh -i ${EC2_KEY_PATH} $(EC2_USER)@$(EC2_IP) "sudo systemctl restart which-movie"
# restart-server: 
# 	# ssh -i ${EC2_KEY_PATH} $(EC2_USER)@$(EC2_IP) "mkdir /home/ubuntu/testing"