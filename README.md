# goss

goss是一个使用go语言编写的简单的自动化运维工具，它包含了ssh和sftp客户端。

为什么要使用goss，它相对于现有的自动化运维工具没有那么复杂的概念和难以处理的依赖。你只需要从仓库下载编译完成的二进制文件，传输到你的服务器上就能愉快的使用。

# 基础使用

使用十分简单 `goss -h` 即可查看帮助。

```go
goss -h

ssh/sftp tools

Usage:
  goss [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  download    Download files from remote servers.
  execute     Remote execution of commands or scripts
  explain     Get detailed explanation of task configuration
  help        Help about any command
  task        Run tasks according to the configuration file
  upload      Upload files to remote servers.

Flags:
      --address string    The address for remote connection to the server, if you want to specify a port, please add a colon and port after the address. For example: 127.0.0.1:22 .
      --auto              Whether to automatically skip remote private key verification.
  -h, --help              help for goss
      --pass string       Remote connection user password.
      --sudopass string   Password for privileged account.
      --timeout int       Maximum waiting time for SSH. (default 180)
  -t, --toggle            Help message for toggle.
      --user string       Remote connection username.
  -v, --version           version for goss

Use "goss [command] --help" for more information about a command.
```

![](https://cdn.nlark.com/yuque/0/2024/png/42497920/1735471394307-7e14c19b-5b74-4509-9dc3-e375c33d4fa1.png)

# 自动化

`goss task -h `

```go
goss task -h

Run tasks according to the configuration file

Usage:
  goss task [flags]

Flags:
  -F, --file string   Develop task configuration file path, please use explain to view configuration details. (default "./task.yml")
  -h, --help          help for task

Global Flags:
      --address string    The address for remote connection to the server, if you want to specify a port, please add a colon and port after the address. For example: 127.0.0.1:22 .
      --auto              Whether to automatically skip remote private key verification.
      --pass string       Remote connection user password.
      --sudopass string   Password for privileged account.
      --timeout int       Maximum waiting time for SSH. (default 180)
      --user string       Remote connection username.
```

![](https://cdn.nlark.com/yuque/0/2024/png/42497920/1735471570847-cf7c52cb-cda0-4cd6-91da-bea87bac86aa.png)

## 配置文件参数参考

`<font style="color:rgb(27, 28, 33);">`结构用于定义远程命令执行、文件上传和下载以及服务器组管理的配置。此配置以 YAML 格式指定。`</font>`

### Global

gossh全局设置。

+ timeOut: `string` - 操作的超时值（秒）。
+ workLoad: `int` - 允许的并发操作数量。
+ remoteUser: `string` - 默认远程操作用户。
+ remotePasswd: `string` - 默认远程用户的密码，如果是纯数字请使用双引号括起来。

### Groups

多个组的集合

#### group

+ name: `string` - 组的名称。
+ remoteUser: `string` - 此组内的远程操作用户，如果不指定默认采用全局。
+ remotePasswd: `string` - 此组内远程用户的密码，如果不指定默认采用全局。如果是纯数字请使用双引号括起来。
+ address: `[]string` - 服务器地址列表。
+ sudoPass: `[]string` - 与地址列表中的服务器对应的特权账号密码列表。

### runs

runs多个任务的集合

#### run

+ name: `string` - 运行的名称。
+ groups: `[]string` - 适用的组名列表，如果不指定默认全体组都执行。
+ upload: `upload` - 可选的上传操作详情。
+ download: `download` - 可选的下载操作详情。
+ execute: `execute` - 可选的命令执行详情。
+ ignore: `ignore` - 是否在运行过程中忽略错误继续执行(当前版本参数无效，默认忽略)。

##### download

定义了下载操作的详细信息。

+ remotePath: `string` - 远程服务器上的目标路径。
+ localPath: `string` - 本地机器上的源路径。
+ cover: `bool` - 是否覆盖远程服务器上已有的文件。

##### upload

定义了上传操作的详细信息。

+ remotePath: `string` - 远程服务器上的目标路径。
+ localPath: `string` - 本地机器上的源路径。
+ cover: `bool` - 是否覆盖远程服务器上已有的文件。

##### execute

结构定义了命令执行操作的详细信息。

+ command: `string` - 远程服务器执行的命令。

```yaml
global:
  timeOut: 30
  workLoad: 5
  remoteUser: admin
  remotePasswd: secret

groups:
  - name: webservers
    remoteuser: webadmin
    remotepasswd: websecret
    address:
      - 192.168.1.101
      - 192.168.1.102
    sudopass:
      - pass1
      - pass2
runs:
  - name: update_webservers
    groups:
      - webservers
    execute:
      command: "sudo apt-get update && sudo apt-get upgrade"
    ignore: false
  - name: "update file"
    upload: 
      remotePath: "/opt/" 
      localPath: "/tmp/111.tar.gz"
  - name: "download file"
    upload: 
      remotePath: "/opt/111.tar.gz" 
      localPath: "/tmp"

```

# 输出样例

## 成功

![](https://cdn.nlark.com/yuque/0/2024/png/42497920/1735473474760-e7c32fe7-db0d-40cd-88b2-c82f887e43bc.png)

## 失败

![](https://cdn.nlark.com/yuque/0/2024/png/42497920/1735473576167-fc4c069e-d5e6-4994-9f2a-d212c2e448f9.png)

## 当前版本问题以及后续优化

- [ ] 完成脚本分发命令编写，当前版本可用使用upload + execute组合完成
- [ ] 优化执行逻辑，执行失败的不再参与后续任务的执行
- [ ] 优化日志输出
- [ ] 完成explain命令编写，类似kubectl explain 让用户更加方便的查看自动化配置项。
- [ ] 传输文件增加进度条
