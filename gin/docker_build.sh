#!/bin/bash

sudo docker build -t smart_shelter:latest ./
# 사설 레지스트리
sudo docker tag smart_shelter:latest 10.1.0.118:443/smart_shelter:latest
sudo docker push 10.1.0.118:443/smart_shelter:latest
