# WindCtl 风电变桨与偏航控制系统

WindCtl 是风电机组变桨与偏航控制系统的进程内控制服务，负责机组注册、风速/转速
采样、变桨指令序列与并发串行化、偏航对风与解缆、安全链触发与复位、控制策略切换
与回滚、采样配额与全链路审计。全部数据通过本地文件持久化，无外部数据库依赖。

## 构建与运行

```bash
go build -mod=vendor ./...
go run ./cmd/windctl -addr :8080 -data ./data
```

控制台默认监听 8080 端口：

- `GET /healthz` 健康检查
- `GET /api/units` 机组状态列表
- `POST /api/units` 注册机组
- `POST /api/units/{id}/pitch` 下发变桨指令
- `POST /api/units/{id}/feather` 顺桨
- `POST /api/units/{id}/limit` 限功率
- `POST /api/units/{id}/yaw` 偏航对风
- `POST /api/units/{id}/unwind` 解缆
- `POST /api/units/{id}/trip` 安全链触发
- `POST /api/units/{id}/release` 安全链释放
- `POST /api/units/{id}/reset` 复位
- `POST /api/units/{id}/stop` 紧急停机
- `POST /api/units/{id}/speed-source/switch` 转速源切换
- `GET/POST /api/units/{id}/strategy` 策略查询与切换
- `GET /api/audit` 审计日志
- `GET /api/state` 持久化状态快照

## Docker

```bash
bash build_benzhi_docker.sh
```

镜像使用离线 vendor 构建，容器内以 `/app/data` 持久化控制数据。
