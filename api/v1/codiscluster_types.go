/*
Copyright 2019 fangjianfeng.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FeSpec defines the desired state of CodisFe
type CodisFeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Service Type string describes ingress methods for a service
	// Default: ClusterIP NodePort LoadBalancer ExternalName
	ServiceType v1.ServiceType `json:"serviceType,required"`

	// If ServiceType is LoadBalancer, need to provide one aliyun slb-instanceid
	SlbInstanceId string `json:"slbInstanceId,omitempty"`

	// The port of access to codis-fe website,
	// Default: 8080
	ServicePort int32 `json:"servicePort,omitempty"`

	// Specify the number of codis-fe, Default: 1
	Replicas *int32 `json:"replicas,required"`

	// Complex select a node with a specific label to run pod
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Simple select a node with a specific label to run pod
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Custom lifestyle cmd script, default: None
	LifeCycle *v1.Lifecycle `json:"lifecycle,omitempty"`

	// ResourceRequirements describes the compute resource requirements.
	Requires v1.ResourceRequirements `json:"codisFeResources,omitempty"`
}

// DashboardSpec defines the desired state of CodisDashboard
type CodisDashboardSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Service Type string describes ingress methods for a service.
	// Default: ClusterIP NodePort LoadBalancer ExternalName
	ServiceType v1.ServiceType `json:"serviceType,required"`

	// If ServiceType is LoadBalancer, need to provide one aliyun slb-instanceid
	SlbInstanceId string `json:"slbInstanceId,omitempty"`

	// The port of access to codis-dashboard
	// Default: 18080
	ServicePort int32 `json:"servicePort,required"`

	// Specify the number of codis-dashboard
	Replicas *int32 `json:"replicas,required"`

	// Complex select a node with a specific label to run pod
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Simple select a node with a specific label to run pod
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Custom lifestyle cmd script.
	// Default: /bin/bash -c "codis-admin -v --dashboard-list --zookeeper=${ZK_ADDR}"
	LifeCycle *v1.Lifecycle `json:"lifecycle,omitempty"`

	// ResourceRequirements describes the compute resource requirements.
	Requires v1.ResourceRequirements `json:"codisDashboardResources,omitempty"`
}

// HaSpec defines the desired state of CodisHa
type CodisHaSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Specify the number of codis-ha
	Replicas *int32 `json:"replicas,required"`

	// Complex select a node with a specific label to run pod
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Simple select a node with a specific label to run pod
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Custom lifestyle cmd script, default: None
	LifeCycle *v1.Lifecycle `json:"lifecycle,omitempty"`

	// ResourceRequirements describes the compute resource requirements.
	Requires v1.ResourceRequirements `json:"codisHaResources,omitempty"`
}

// ProxySpec defines the desired state of CodisProxy
type CodisProxySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Service Type string describes ingress methods for a service.
	// Default: ClusterIP NodePort LoadBalancer ExternalName
	ServiceType v1.ServiceType `json:"serviceType,required"`

	// If ServiceType is LoadBalancer, need to provide one aliyun slb-instanceid
	SlbInstanceId string `json:"slbInstanceId,omitempty"`

	// The port of access to codis-proxy
	// Default: 19000
	ServicePort int32 `json:"servicePort,omitempty"`

	// Specify the number of codis-proxy
	Replicas *int32 `json:"replicas,required"`

	// Complex select a node with a specific label to run pod
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Simple select a node with a specific label to run pod
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Custom lifestyle cmd script, default: None
	LifeCycle *v1.Lifecycle `json:"lifecycle,omitempty"`

	// ResourceRequirements describes the compute resource requirements.
	Requires v1.ResourceRequirements `json:"codisProxyResources,omitempty"`

	// The password access Redis-server, default is empty.
	Auth string `json:"auth,omitempty"`
}

type CronJobSpec struct {

	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	// Optional deadline in seconds for starting the job if it misses scheduled
	// time for any reason.  Missed jobs executions will be counted as failed ones.
	// +optional
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty"`

	// Specifies how to treat concurrent executions of a Job.
	// Valid values are:
	// - "Allow" (default): allows CronJobs to run concurrently;
	// - "Forbid": forbids concurrent runs, skipping next run if previous run hasn't finished yet;
	// - "Replace": cancels currently running job and replaces it with a new one
	// +optional
	ConcurrencyPolicy batchv1beta1.ConcurrencyPolicy `json:"concurrencyPolicy,omitempty"`

	// The number of successful finished jobs to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// Defaults to 3.
	// +optional
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty"`
}

// RedisSpec defines the desired state of CodisServer
type CodisRedisSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Specify the number of codis-server
	Replicas *int32 `json:"replicas,required"`

	CodisServerNumsPerGroup int `json:"codisServerNumsPerGroup,required"`

	// Scheduler name.
	// If you implement your own scheduler, you can specify the name here
	// Default is empty
	Scheduler string `json:"scheduler"`

	// Complex select a node with a specific label to run pod
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Simple select a node with a specific label to run pod
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Custom lifestyle cmd script, default is empty
	// Default PostStart:
	//   "/bin/sh", "-c", "until [ "`redis-cli -h $POD_IP ping | grep -c PONG`" = 1 ]; do echo 'Waiting 1s for Redis to load'; sleep 1; done; \
	//   codis-admin --dashboard=${DASHBOARD_ADDR} --reload; echo 'reload dashboard'; if [ $? != 0 ]; then exit 1; fi; \
	//   sid=`hostname |awk -F'-' '{print $NF}'`; gid=$(expr $sid / ${SERVER_REPLICA} + 1);\
	//   codis-admin --dashboard=${DASHBOARD_ADDR} --list-group | grep "\"id\": ${gid}," > /dev/null || codis-admin --dashboard=${DASHBOARD_ADDR} --create-group --gid=${gid}; \
	//   codis-admin --dashboard=${DASHBOARD_ADDR} --group-add --gid=${gid} --addr=${HOST_NAME}:6379; \
	//   if [ $? != 0 -a $SERVER_REPLICA -gt 1 ]; then exit 2; fi; \
	//   codis-admin --dashboard=${DASHBOARD_ADDR} --sync-action --create --addr=${HOST_NAME}:6379 1>/dev/null 2>&1"
	//
	// Default PreStop:
	//   "/bin/sh", "-c", "codis-admin --dashboard=${DASHBOARD_ADDR} --reload; if [ $? != 0 ]; then exit 1; fi; \
	//   sid=`hostname |awk -F'-' '{print $NF}'`; gid=$(expr $sid / ${SERVER_REPLICA} + 1);\
	//   codis-admin --dashboard=${DASHBOARD_ADDR} --group-del --gid=${gid} --addr=${HOST_NAME}:6379"
	LifeCycle *v1.Lifecycle `json:"lifecycle,omitempty"`

	// CodisServerRequires describes the CodisServer compute resource requirements.
	CodisServerRequires v1.ResourceRequirements `json:"codisServerResources,omitempty"`

	// CodisServerExporterRequires describes the CodisServerExporter compute resource requirements.
	CodisServerExporterRequires v1.ResourceRequirements `json:"codisServerExporterResources,omitempty"`

	// Redis-exporter image
	ExporterImage string `json:"exporterImage,required"`

	// Redis-exporter image
	BackupImage string `json:"backupImage,required"`

	// Redis RDB backup policy
	BackupCronJob CronJobSpec `json:"backupCronJob,omitempty"`

	// StorageClassName
	StorageClassName string `json:"storageClassName,omitempty"`

	// PVC source
	StorageRequires v1.ResourceRequirements `json:"storageResources,omitempty"`

	// Redis configuration
	// It will generate a configmap
	RedisConf map[string]string `json:"redisConf,omitempty"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CodisClusterSpec defines the desired state of CodisCluster
type CodisClusterSpec struct {
	LogPath       string             `json:"logPath,required"`
	ZkAddr        string             `json:"zkAddr,required"`
	ProductName   string             `json:"productName"`
	CodisImage    string             `json:"codisImage,required"`
	RedisSpec     CodisRedisSpec     `json:"redisSpec"`
	DashboardSpec CodisDashboardSpec `json:"dashboardSpec"`
	FeSpec        CodisFeSpec        `json:"feSpec"`
	HaSpec        CodisHaSpec        `json:"haSpec"`
	ProxySpec     CodisProxySpec     `json:"proxySpec"`
}

// CodisClusterStatus defines the observed state of CodisCluster
type CodisClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// CodisCluster is the Schema for the codisclusters API
type CodisCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CodisClusterSpec   `json:"spec,omitempty"`
	Status CodisClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CodisClusterList contains a list of CodisCluster
type CodisClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CodisCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CodisCluster{}, &CodisClusterList{})
}
