
# Go OpenBmclAPI

Golang 版的 [OpenBmclAPI](https://github.com/bangbang93/openbmclapi)

> 注: 本仓库部分代码按照 <https://github.com/bangbang93/openbmclapi> 编写, **非官方, 不保证稳定性/时效性**

*如果本项目有用, 就给个 star ⭐️ 吧 :)*

欢迎进行 PR.

-> [完整贡献者列表](./CONTRIBUTORS.md)

## 特性

- 使用协程/多线程, 超高的文件同步速度
- 不依赖大量的三方包, 体积小巧
- 静态文件, 无需配置任何依赖
- 得益于 Golang 强大的跨平台编译器, 支持大部分主流平台/处理器
- 内置一个实时更新的状态面板
- 支持一些小众的~~奇怪~~需求
- ~~更好的压榨节点~~

## FAQ

Q: 支持`.env`文件吗?  
A: 不支持, Go-OpenBmclAPI 使用 yaml 配置文件, 该文件应位于程序的运行目录. 初次运行本程序时会自动创建该配置文件

Q: 没有适合我的平台的程序怎么办?  
A: 检查 [Go Platform Values](https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63#platform-values), 如果您的系统/架构存在, 那么可以发个 issue

Q: 版本号格式  
A: `v<openbmclapi 版本>-<go-openbmclapi 构建计数>`, 例: `v1.6.7-60`

Q: 面板在哪里访问?  
A: 服务端启动时会输出 `https://<public-host>:<public-port>`, 访问这个地址就可以了
   > 注: 面板与 openbmclapi 下载服务共用一个端口, 并非单独配置

## 安装

### 无依赖直接运行

1. 从 github release 找到**适合您服务器平台**的程序并下载到节点服务器上
2. 配置配置文件, 可以直接使用与bangbang93的openbmclapi相同的环境变量配置, 也可以从`config.yaml`进行配置 _(下文有讲)_
3. 运行程序

### 从docker运行

- 可直接运行仓库目录下的 [`docker-run.sh`](./docker-run.sh) 文件, 运行之前请确保存在`craftmine/go-openbmclapi:latest`镜像, 或网络连通
- 也可使用`docker build -t craftmine/go-openbmclapi:latest .`手动编译镜像

### 从`systemd`运行 _(仅Linux)_

1. **确保systemd是您的*启动进程***
2. 执行
   ```sh
   sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/LiterMC/go-openbmclapi/HEAD/service/installer.sh)"
   ```
   > 注意, 新版新增使用 `openbmclapi` 用户执行程序, 可能需要执行 `sudo chown -R openbmclapi /opt/openbmclapi` 指令修复权限

   国内对 Github 的支持较差, 可以使用 ghproxy 等镜像站运行脚本, 本例中使用了 [crashmc.com](https://crashmc.com) 提供的 CDN:
   ```sh
   MIRROR_PREFIX=https://cdn.crashmc.com/ sudo bash -c "$(curl -fsSL ${MIRROR_PREFIX}https://raw.githubusercontent.com/LiterMC/go-openbmclapi/HEAD/service/installer.sh)"
   ```
3. 配置`/opt/openbmclapi/config.yaml`配置文件
4. 使用`systemctl start go-openbmclapi.service`启动服务
   - 使用`systemctl reload go-openbmclapi.service`可重新加载配置文件
   - 使用`systemctl stop go-openbmclapi.service`停止服务
   - 使用`systemctl enable go-openbmclapi.service`让服务自启动
   - 使用`systemctl disable go-openbmclapi.service`禁止服务自启动
   - 使用`journalctl -f --output cat -u go-openbmclapi.service`实时监听日志

### 从源代码运行

1. 下载本仓库 _(可以使用`git clone https://github.com/LiterMC/go-openbmclapi.git`)_
2. 确保您的服务器上装有 `go 1.21+` 以及 `node & npm`
3. cd进入本仓库
4. 配置配置文件或环境变量
5. 使用 `go generate . && go run .` 运行本程序

### 第三方自更新+无依赖直接运行脚本

源代码位于 <https://github.com/8Mi-Tech/Shell/blob/main/go-openbmclapi-in-tmux.sh>

```
cd /opt/openbmclapi
wget -O start.sh https://raw.githubusercontent.com/8Mi-Tech/Shell/main/go-openbmclapi-in-tmux.sh
bash ./start.sh
```

## 配置

### 使用配置文件

#### 注意: 配置文件指定的值会被环境变量的值 _(如果存在)_ 覆盖掉

配置文件应为**运行目录**下的`config.yaml`文件, 使用`yaml`格式解析  
例:
```yaml
# 是否打印访问信息, 默认为否 (这个选项对于压缩日志文件十分有用)
record-serve-info: false
# 日志最长保存时间 (天). 设置为 0 禁用清理过期日志
log-slots: 7
# 是否禁用 bmclapi 分发的证书, 同 CLUSTER_BYOC
byoc: false
# 是否信任 X-Forwarded-For 标头 (有反代时启用)
trusted-x-forwarded-for: false
# 实际开放的公网主机名, 同 CLUSTER_IP
public-host: example.com
# 实际开放的公网端口, 同 CLUSTER_PUBLIC_PORT
public-port: 8080
# 要监听的本地端口, 同 CLUSTER_PORT
port: 4000
# CLUSTER_ID
cluster-id: ${CLUSTER_ID}
# CLUSTER_SECRET
cluster-secret: ${CLUSTER_SECRET}
# 文件同步间隔 (分钟)
sync-interval: 10
# 同步文件时最多打开的连接数量. 注意: 该选项目前没用
download-max-conn: 64

# 缓存
cache:
  # 缓存类型:
  #   nocache: 不缓存
  #   inmem: 程序内内存缓存
  #   redis: Redis 缓存
  type: inmem
  # 如果使用 Redis 缓存则还需要配置用户名密码等:
  data:
    network: tcp
    addr: "redis.server.hostname:6379"
    client-name: "go-openbmclapi"
    username: redis-username
    password: redis-password

# 服务器上行限制
serve-limit:
  # 是否启用上行限制
  enable: false
  # 最大连接数量
  max-conn: 16384
  # 上行速率限制 (KiB/s), 0 表示无限制
  upload-rate: 0

# 内置的仪表板
dashboard:
  # 是否启用
  enable: true
  # PWA 的名称, 在桌面设备上显示
  pwa-name: GoOpenBmclApi Dashboard
  # PWA 短名称, 在移动设备上显示
  pwa-short_name: GOBA Dash
  # PWA 描述
  pwa-description: Go-Openbmclapi Internal Dashboard

# 子存储节点列表
# 注意: measure 测量请求总是以第一个存储为准
storages:
  # local 为本地存储
  - type: local
    # 节点 ID
    id: local-storage-1
    # 使用该子节点的概率 (非负整数)
    weight: 100
    # 节点附加数据
    data:
      # cache 文件夹到路径
      cache-path: cache
      # 压缩方式 (目前未使用)
      compressor: ""
  # mount 为网络存储 (与旧版 oss 选项含义大致相同)
  - type: mount
    # 节点 ID
    id: mount-storage-1
    # 使用该子节点的概率 (非负整数)
    # 设为 0 将使该子节点成为备选节点 (若该节点前一个节点失效才会使用该节点), 如果所有子节点均为 0 则平均分配
    weight: 0
    # 节点附加数据
    data:
      # 文件夹路径
      path: oss_mirror
      # 对应的网络URL路径
      redirect-base: https://oss.example.com/base/paths
      # 启动之前在 measure 子文件夹内生成 1-200MB 的测速文件 (默认为动态生成)
      pre-gen-measures: false
  # webdav 使用 webdav 存储
  - type: webdav
    # 节点 ID
    id: webdav-storage-1
    # 使用该子节点的概率 (非负整数)
    weight: 100
    # 节点附加数据
    data:
      # 最多同时发起的连接数
      max-conn: 24
      # 最大上传速率 (KiB/s), 0 表示无限制
      max-upload-rate: 0
      # 最大下载速率 (KiB/s), 0 表示无限制
      max-download-rate: 0
      # 启动之前生成 1-200MB 的测速文件 (默认为动态生成)
      pre-gen-measures: false
      # 设置为 true 后将跟踪 302 请求 (即不会将最终用户重定向到网盘)
      follow-redirect: false
      # 重定向链接的缓存时间, 仅当 follow-redirect 为 false 时有用. 0 表示不缓存重定向链接
      redirect-link-cache: 0s
      # 链接到下方 webdav-users 的键值对
      alias: example-user
      # 相对于 alias 中的 Webdav 入口 URL **注意⚠️: 不要使用非 ascii (包括中文) 路径**
      endpoint: ../optional/another/endpoint/
      # [可选] 覆盖 alias 中的用户名
      username: optional-another-username
      # [可选] 覆盖 alias 中的密码
      password: optional-another-password

webdav-users:
    example-user:
        # Webdav 入口 URL **注意⚠️: 不要使用非 ascii (包括中文) 路径**
        endpoint: https://webdav.example.com/path/to/endpoint/
        # 用户名
        username: example-username
        # 密码
        password: example-password

# 以下为高级选项, 通常用于调试. ** 如果不理解其工作原理请不要碰 **
advanced:
  # 是否打印调试日志
  debug-log: false
  # 跳过第一次同步, 直接启动节点
  skip-first-sync: false
  # 是否在连接断开后直接退出
  exit-when-disconnected: false
  # 是否仅从主服务器下载文件
  noopen: false
  # 跳过文件哈希值校验
  no-heavy-check: false
  # 发送心跳包的超时限制 (秒), 网不好就调高点
  keepalive-timeout: 10

```

## 子命令

Go-OpenBmclAPI 提供了一组子命令:
```
Sub commands:
  help
        显示帮助消息

  main | serve | <empty>
        执行主程序

  license
        打印完整许可证

  version
        打印程序版本

  zip-cache [options ...]
        使用 gzip 压缩 cache 文件夹内的文件 (迁移用)

    Options:
      verbose | v : 显示正在压缩的文件
      all | a : 压缩所有文件 (默认不会压缩10KB以下的文件)
      overwrite | o : 覆盖存在的已压缩的目标文件
      keep | k : 不删除压缩过的文件

  unzip-cache [options ...]
        解压缩 cache 文件夹内的文件 (迁移用)

    Options:
      verbose | v : 显示正在解压缩的文件
      overwrite | o : 覆盖存在的未压缩的目标文件
      keep | k : 不删除解压缩过的文件
  upload-webdav
        将本地 cache 文件夹上传到 webdav 存储
        上传之前请确保 config.yaml 下存在至少一个 local 存储和至少一个 webdav 存储
```

## 致谢

[OpenBMCLAPI-TypeScript](https://github.com/bangbang93/openbmclapi)（[@bangbang93](https://github.com/bangbang93)）
