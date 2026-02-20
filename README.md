# openvpn

![image-20240529110841439](https://raw.githubusercontent.com/GavinTan/files/master/picgo/image-20240529110841439.png)

![20220930173030](https://raw.githubusercontent.com/GavinTan/files/master/picgo/20220930173030.png)

![20220930173103](https://raw.githubusercontent.com/GavinTan/files/master/picgo/20220930173103.png)

![image-20241015170847764](https://raw.githubusercontent.com/GavinTan/files/master/picgo/image-20241015170847764.png)

**openvpn 安全与加密相关配置参考于[openvpn-install](https://github.com/angristan/openvpn-install?tab=readme-ov-file#security-and-encryption)的Security and Encryption部分。**

> 提示：
>
> 1. 管理员登录账号密码默认`admin:admin`，登录后可在系统设置里修改。
> 2. 系统同时支持使用vpn账号登录，vpn账号登录后可设置mfa、修改密码、下载配置文件。
> 3. 第一步：web -> 管理 -> 客户端里生成下载客户端配置文件。
> 4. 第二部：web -> 管理 -> VPN 账号里管理添加账号（默认启用账号验证可在 VPN 账号里开启或关闭）。
> 5. 第三步：使用vpn账号登录 -> web -> 下载客户端 （设置mfa、修改密码）-> 下载配置文件 -> 导入客户端进行连接。
>
> 
>
> 注意：
>
> 1. 默认场景应用于组网禁用vpn网关，如果需要客户端所有流量都走 openvpn 需要在系统设置openvpn里启用vpn网关。
> 2. 系统设置里修改openvpn配置生效必须启用自动更新配置，如果要保留自己修改的server.conf配置则需要禁用自动更新配置。
> 3. 系统设置openvpn配置只是修改了server.conf，应用需要手动重启openvpn服务。
> 4. 设置mfa需要使用要设置的vpn账号登录用户页面后进行mfa设置，在用户页面下载配置文件会动态更新配置文件mfa参数，客户端页面的开启mfa选项是直接在配置文件添加mfa参数（直接在后台下载配置文件需要设置该选项，通过用户页面下载配置文件可忽略）。
> 5. 在创建客户端后关闭账号验证客户端的配置文件存在auth-user-pass参数客户端会依旧弹出登录，登录信息可以随便输入不会做验证，若有弹窗困扰的建议手动编辑客户端配置文件注释掉参数或重新生成客户端配置文件。

## 支持功能

- 账号管理
- 证书管理
- ipv6支持
- ldap支持
- mfa支持
- 邮件通知
- 连接历史记录
- vpn账号固定ip
- vpn账号分组路由推送
- 在线编辑server.conf
- 在线重启openvpn服务
- 一键生成客户端 & CCD配置文件

## Quick Start

> 裸机部署：目前提供一个支持部分发行版的安装脚本`scripts/openvpn-install.sh`。

运行 openvpn

```shell
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
  k8scat/openvpn:latest
```

### compose

- 安装 docker-compose

  ```bash
  yum install -y docker-compose-plugin
  ```
  
- 创建 docker-compose.yml

  ```yaml
  services:
  openvpn:
    image: k8scat/openvpn:latest
    container_name: openvpn
    cap_add:
      - NET_ADMIN
    ports:
      - "1194:1194/udp"
      - "8833:8833"
    environment:
      - SYSTEM_BASE_ADMIN_USERNAME=admin
      - SYSTEM_BASE_ADMIN_PASSWORD=admin
      - OVPN_GATEWAY=false # 如果是放在网关上用，需要设置为true
    volumes:
      - /srv/openvpn/data:/data
      - /etc/localtime:/etc/localtime:ro
      - /srv/openvpn/logs:/var/log
  ```
  
- 运行 openvpn

  ```bash
  docker compose up -d
  ```

### Docker Secret（Swarm 模式）

Docker Secret 仅在 **Docker Swarm** 下可用，密码以只读文件形式挂载到容器的 `/run/secrets/`，由 entrypoint 自动注入为环境变量。

**1. 初始化 Swarm（若未初始化）**

```bash
docker swarm init
```

**2. 创建 secret**

```bash
# 从 stdin 创建（推荐，不落盘）
echo -n 'your_admin_password' | docker secret create admin_password -

# 或从文件创建
echo -n 'your_admin_password' > /tmp/admin_password
docker secret create admin_password /tmp/admin_password
rm /tmp/admin_password
```

**3. 使用 compose 部署 stack**

```bash
docker stack deploy -c docker-compose.swarm.yml openvpn
```

**4. 常用命令**

```bash
# 查看 secret 列表
docker secret ls

# 更新密码：先创建新 secret，再更新服务使用新 secret（需在 compose 中改 secret 名或滚动更新）
# 删除旧 secret（需先从服务中移除）
docker secret rm admin_password
```

镜像内已支持：若存在 `/run/secrets/admin_password` 且未设置环境变量 `SYSTEM_BASE_ADMIN_PASSWORD`，会自动从该文件读取并导出。因此 Swarm 部署时只需在 compose 中挂载 secret，无需再传密码环境变量。

## IPV6

>注意：
>
>1. 需要在页面系统设置openvpn里启用ipv6
>2. 启用ipv6后客户端跟服务器的proto需要都指定udp6/tcp6
>3. docker的网络需要启用ipv6支持
>4. 使用openvpn-connect客户端的需要使用3.4.1以后的版本

```bash
services:
  openvpn:
    image: k8scat/openvpn
    cap_add:
      - NET_ADMIN
    ports:
      - "1194:1194/udp"
      - "8833:8833"
    volumes:
      - ./data:/data
      - /etc/localtime:/etc/localtime:ro
    sysctls:
      - net.ipv6.conf.default.disable_ipv6=0
      - net.ipv6.conf.all.forwarding=1

networks:
  default:
    enable_ipv6: true
```

## LDAP

> 在系统设置里启用LDAP认证，启用LDAP认证后本地的VPN账号将不在工作。

部分参数说明：

- LDAP_URL：ldap连接TLS 例：ldaps://example.org:636
- LDAP_USER_ATTRIBUTE：根据当前使用的LDAP服务器设置，例：OpenLDAP：uid ； Windows AD: sAMAccountName
- LDAP_USER_ATTR_IPADDR_NAME：可在ldap服务器添加ipaddr自定义字段，也可以设置为ldap已经存在其他的未使用字段 例：mobile、homePhone
- LDAP_USER_ATTR_CONFIG_NAME：可在ldap服务器添加config自定义字段，也可以设置为ldap已经存在其他的未使用字段 例：mobile、homePhone
