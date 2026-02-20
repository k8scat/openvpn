#!/bin/bash
set -e

docker run -d \
  --name openvpn \
  --cap-add=NET_ADMIN \
  -p 1194:1194/udp \
  -p 8833:8833 \
  -e OPENVPN_OVPN_GATEWAY=true \
  -e SYSTEM_BASE_ADMIN_USERNAME=admin \
  -e SYSTEM_BASE_ADMIN_PASSWORD=admin \
  -v /srv/openvpn/data:/data \
  -v /srv/openvpn/logs:/var/log \
  -v /etc/localtime:/etc/localtime:ro \
  k8scat/openvpn:v2.5.4
