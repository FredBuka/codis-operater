package util

import (
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

const (
	// DefaultProxyTargetPort define proxy default port
	DefaultProxyPort int32 = 19000
	// DefaultProxyHostPort define proxy host default port
	DefaultProxyHostPort int32 = 11080
	// DefaultDashboardTargetPort define dashboard default port
	DefaultDashboardPort int32 = 18080
	// DefaultFeTargetPort define fe default port
	DefaultFePort int32 = 8080
	// DefaultServerPort define redis default port
	DefaultServerPort int32 = 6379
	// DefaultServerPort define redis-exporter default port
	DefaultServerExporterPort int32 = 9121
	// DefaultHaInterval define ha health check intervals
	DefaultHaInterval int = 5

	// Codis component name
	// Modifying the component name will affect the service name, pvc name, volume mount, cronjob and other resources
	CodisDashboard = "codis-dashboard"
	CodisFe        = "codis-fe"
	CodisHa        = "codis-ha"
	CodisProxy     = "codis-proxy"
	CodisRedis     = "codis-server"
	CodisRedisPVC  = "codis-server-data"
	CodisBackup    = "codis-backup"
	CodisConf      = "codis-conf"

	// Codis og path
	CommonLogPath = "/codis/log"
	// Codis Server rdb path
	CodisServerDataPath = "/codis/data"
	// Codis Server configuration path
	CodisServerConfPath = "/data/service/codis/conf"

	// health check telnet image
	TelnetImg = "mikesplain/telnet"
	// health check curl image
	CurlImg = "curlimages/curl"
)

// DefaultCommonEnvs provide common env.
// The environment variables will be automatically added to all pod containers created by deployment and statefulset
func DefaultCommonEnvs() []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "POD_NAMESPACE", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}}},
		{Name: "POD_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}}},
		{Name: "POD_IP", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}}},
		{Name: "NODE_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
	}
}

// DefaultDashboardCommand provide default startup command of codis-dashboard
func DefaultDashboardCommand() []string {
	return []string{"codis-dashboard"}
}

// DefaultDashboardArgs provide default startup command startup parameters of codis-dashboard
func DefaultDashboardArgs() []string {
	return []string{
		"-l", "/codis/log/$(POD_NAMESPACE)_$(POD_NAME).log",
		"-c", "/data/service/codis/config-default/dashboard.toml",
		"--host-admin", fmt.Sprintf("$(POD_IP):%v", DefaultDashboardPort),
		"--zookeeper", "$(ZK_ADDR)",
		"--product_name", "$(PRODUCT_NAME)",
		"--pidfile", "/tmp/dashboard.pid",
		"--remove-lock",
	}
}

// DefaultDashboardLifeCycle provide default lifCycle of codis-dashboard
// Support for customizing this parameter in k8s deployment files
func DefaultDashboardLifeCycle() *corev1.Lifecycle {
	return &corev1.Lifecycle{
		PostStart: &corev1.Handler{
			Exec: &corev1.ExecAction{Command: []string{
				"/bin/bash", "-c", "codis-admin -v --dashboard-list  --zookeeper=${ZK_ADDR}",
			}},
		},
		PreStop: &corev1.Handler{
			Exec: &corev1.ExecAction{Command: []string{
				"/bin/sh", "-c", "PID=$(cat /tmp/dashboard.pid) && kill $PID && while ps -p 1 > /dev/null; do sleep 1; done",
			}},
		},
	}
}

// DefaultFeCommand provide default startup command of codis-fe
func DefaultFeCommand() []string {
	return []string{"codis-fe"}
}

// DefaultFeArgs provide default startup command startup parameters of codis-fe
func DefaultFeArgs() []string {
	return []string{
		"-l", "/codis/log/$(POD_NAMESPACE)_$(POD_NAME).log",
		"--zookeeper", "$(ZK_ADDR)",
		"--listen", fmt.Sprintf("$(POD_IP):%v", DefaultFePort),
		"--assets=/data/service/codis/bin/assets",
	}
}

// DefaultFeLifeCycle provide default lifCycle of codis-fe
//// Support for customizing this parameter in k8s deployment files
func DefaultFeLifeCycle() *corev1.Lifecycle {
	return &corev1.Lifecycle{}
}

// DefaultHaCommand provide default startup command of codis-ha
func DefaultHaCommand() []string {
	return []string{"codis-ha"}
}

// DefaultHaArgs provide default startup command startup parameters of codis-ha
func DefaultHaArgs() []string {
	return []string{
		"--log", "/codis/log/$(POD_NAMESPACE)_$(POD_NAME).log",
		"--interval", strconv.Itoa(DefaultHaInterval),
		"--dashboard", "$(DASHBOARD_ADDR)",
	}
}

// DefaultHaLifeCycle provide default lifCycle of codis-ha
// Support for customizing this parameter in k8s deployment files
func DefaultHaLifeCycle() *corev1.Lifecycle {
	return &corev1.Lifecycle{}
}

// DefaultProxyCommand provide default startup command of codis-proxy
func DefaultProxyCommand() []string {
	return []string{"codis-proxy"}
}

// DefaultProxyArgs provide default startup command startup parameters of codis-proxy
func DefaultProxyArgs() []string {
	return []string{
		"-l", "/codis/log/$(POD_NAMESPACE)_$(POD_NAME).log",
		"-c", "/data/service/codis/config-default/proxy.toml",
		"--session_auth", "$(SESSION_AUTH)",
		"--host-admin", fmt.Sprintf("$(POD_IP):%v", DefaultProxyHostPort),
		"--host-proxy", fmt.Sprintf("$(POD_IP):%v", DefaultProxyPort),
		"--zookeeper", "$(ZK_ADDR)",
		"--product_name", "$(PRODUCT_NAME)",
		"--pidfile", "/tmp/proxy.pid",
	}
}

// DefaultProxyLifeCycle provide default lifCycle of codis-proxy
// Support for customizing this parameter in k8s deployment files
func DefaultProxyLifeCycle() *corev1.Lifecycle {
	return &corev1.Lifecycle{
		PreStop: &corev1.Handler{
			Exec: &corev1.ExecAction{Command: []string{
				"/bin/sh", "-c", fmt.Sprintf("codis-admin -v --dashboard=${DASHBOARD_ADDR} --remove-proxy --addr=${POD_IP}:%v >> /codis/log/${POD_NAMESPACE}_${POD_NAME}.log 2>&1",
					DefaultProxyHostPort,
				),
			}},
		},
	}
}

// DefaultServerCommand provide default startup command of codis-server
func DefaultServerCommand() []string {
	return []string{"codis-server"}
}

// DefaultServerArgs provide default startup command startup parameters of codis-server
func DefaultServerArgs() []string {
	return []string{
		"/data/service/codis/conf/redis.conf",
		"--logfile", "/codis/log/$(POD_NAMESPACE)_$(POD_NAME).log",
		"--protected-mode", "no",
		"--bind", "$(POD_IP)",
		"--daemonize", "no",
	}
}

var postStartServerCmd = `
until [ %s = 1 ]; do echo 'Waiting 1s for Redis to load'; sleep 1; done; \
codis-admin --dashboard=${DASHBOARD_ADDR} --reload; echo 'reload dashboard'; if [ $? != 0 ]; then exit 1; fi; \
sid=%s; gid=$(expr $sid / ${SERVER_REPLICA} + 1);\
codis-admin --dashboard=${DASHBOARD_ADDR} --list-group | grep "\"id\": ${gid}," 1>/dev/null 2>&1 || codis-admin --dashboard=${DASHBOARD_ADDR} --create-group --gid=${gid}; \
codis-admin --dashboard=${DASHBOARD_ADDR} --group-add --gid=${gid} --addr=${HOST_NAME}:6379; \
if [ $? != 0 -a $SERVER_REPLICA -gt 1 ]; then exit 2; fi; \
codis-admin --dashboard=${DASHBOARD_ADDR} --sync-action --create --addr=${HOST_NAME}:6379 1>/dev/null 2>&1
`

var preStopServerCmd = `
codis-admin --dashboard=${DASHBOARD_ADDR} --reload; if [ $? != 0 ]; then exit 1; fi; \
sid=%s; gid=$(expr $sid / ${SERVER_REPLICA} + 1); sleep 5;\
codis-admin --dashboard=${DASHBOARD_ADDR} --group-del --gid=${gid} --addr=${HOST_NAME}:6379 1>/dev/null 2>&1
`

// DefaultServerLifeCycle provide default lifCycle of codis-server
// Support for customizing this parameter in k8s deployment files
func DefaultServerLifeCycle() *corev1.Lifecycle {
	checkRedisCmd := "`redis-cli -h $POD_IP ping | grep -c PONG`"
	getSidCmd := "`hostname |awk -F'-' '{print $NF}'`"
	return &corev1.Lifecycle{
		PostStart: &corev1.Handler{
			Exec: &corev1.ExecAction{Command: []string{
				"/bin/bash", "-c", fmt.Sprintf(postStartServerCmd, checkRedisCmd, getSidCmd),
			}},
		},
		PreStop: &corev1.Handler{
			Exec: &corev1.ExecAction{Command: []string{
				"/bin/sh", "-c", fmt.Sprintf(preStopServerCmd, getSidCmd),
			}},
		},
	}
}

var redisConfContent = `
bind 127.0.0.1
protected-mode yes
port 6379
tcp-backlog 511
timeout 0
tcp-keepalive 300
daemonize yes
supervised no
pidfile /tmp/redis_6379.pid
loglevel notice
logfile "/tmp/redis_6379.log"
databases 16
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dir ./
slave-serve-stale-data yes
slave-read-only yes
repl-diskless-sync no
repl-diskless-sync-delay 5
repl-disable-tcp-nodelay no
slave-priority 100
maxmemory 32G
appendonly no
appendfilename "appendonly.aof"
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
aof-load-truncated yes
lua-time-limit 5000
slowlog-log-slower-than 10000
slowlog-max-len 128
latency-monitor-threshold 0
notify-keyspace-events ""
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
list-max-ziplist-size -2
list-compress-depth 0
set-max-intset-entries 512
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
hll-sparse-max-bytes 3000
activerehashing yes
client-output-buffer-limit normal 0 0 0
client-output-buffer-limit slave 256mb 64mb 60
client-output-buffer-limit pubsub 32mb 8mb 60
hz 10
aof-rewrite-incremental-fsync yes
`

// DefaultRedisConfiguration provide default redis config
func DefaultRedisConfiguration() map[string]string {
	return map[string]string{"redis.conf": redisConfContent}
}
