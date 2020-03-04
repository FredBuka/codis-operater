package util

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	infrav1 "github/oarfah/codis-operater/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// WithSvcOptions define service options
type WithSvcOptions struct {
	Name            string
	Namespace       string
	TargetPort      int32
	ServicePort     int32
	ServiceType     corev1.ServiceType
	Instanceid      string
	OwnerReferences []metav1.OwnerReference
}

const (
	// common suffix
	DeploymentSuffix      = "-deployment"
	StatefulsetSuffix     = "-statefulset"
	ServiceSuffix         = "-service"
	ServerConfSuffix      = "-conf"
	ServerRdbBackupSuffix = "-backupjob"
	PVCSuffix             = "-pvc"
)

// GenerateService returns service of kubernetes, version: v1
func GenerateService(opt *WithSvcOptions) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            opt.Name + ServiceSuffix,
			Namespace:       opt.Namespace,
			OwnerReferences: opt.OwnerReferences,
			Labels:          map[string]string{"kubernetes.io/cluster-service": "true"},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": opt.Namespace + "-" + opt.Name},
			PublishNotReadyAddresses:true,
		},
	}

	switch opt.ServiceType {
	case corev1.ServiceTypeClusterIP:
		service.Spec.Type = corev1.ServiceTypeClusterIP
		service.Spec.ClusterIP = "None"
		service.Spec.Ports = []corev1.ServicePort{
			{Port: opt.ServicePort, Name: "codis-server"},
		}

	case corev1.ServiceTypeNodePort:
		service.Spec.Type = corev1.ServiceTypeNodePort
		service.Spec.Ports = []corev1.ServicePort{
			{
				Port:       opt.ServicePort,
				TargetPort: intstr.FromInt(int(opt.TargetPort)),
				Name:       opt.Name,
				NodePort:   opt.ServicePort,
				Protocol:   corev1.ProtocolTCP,
			},
		}

	case corev1.ServiceTypeLoadBalancer:
		service.Annotations = map[string]string{
			"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id":                       opt.Instanceid,
			"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners": "true",
			"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-protocol-port":            fmt.Sprintf("http:%v", opt.ServicePort),
			"service.beta.kubernetes.io/backend-type":                                        "eni",
		}
		service.Spec.Type = corev1.ServiceTypeLoadBalancer
		service.Spec.Ports = []corev1.ServicePort{
			{
				Port:       opt.ServicePort,
				TargetPort: intstr.FromInt(int(opt.TargetPort)),
				Name:       opt.Name,
				Protocol:   corev1.ProtocolTCP,
			},
		}
	}

	return service
}

// ContainerSpec struct define all information to start pod
type ContainerSpec struct {
	ContainerName   string
	Image           string
	ImagePullPolicy corev1.PullPolicy
	Command         []string
	Args            []string
	Envs            []corev1.EnvVar
	LifeCycle       *corev1.Lifecycle
	Resources       corev1.ResourceRequirements
	Ports           []corev1.ContainerPort
	VolumeMounts    []corev1.VolumeMount
}

// BackupCronJobSpec struct define all information to start CronJob
type BackupCronJobSpec struct {
	ConcurrencyPolicy          batchv1beta1.ConcurrencyPolicy
	Schedule                   string
	StartingDeadlineSeconds    *int64
	SuccessfulJobsHistoryLimit *int32
	Completions                *int32
}

// Meta struct define common meta
type Meta struct {
	Name            string
	Namespace       string
	Annotations     map[string]string
	OwnerReferences []metav1.OwnerReference
}

// Meta struct define all spec, contains ContainerSpec, BackupCronJobSpec and so on.
type Spec struct {
	InitContainers []ContainerSpec
	Containers     []ContainerSpec
	Affinity       *corev1.Affinity
	NodeSelector   map[string]string
	Replicas       *int32
	LabelSelector  string
	Volumes        []corev1.Volume
	Scheduler      string
	BackupCronJob  BackupCronJobSpec
}

// WithCtrlOption define Options of Deployment AND StatefulSet
type WithCtrlOption struct {
	ObjectKind string
	Meta
	Spec
}

const (
	DeploymentKind  = "Deployment"
	StatefulsetKind = "Statefulset"
	CronJobKind     = "CronJob"
)

// GenerateDeployObject returns new object of deployment or Statefulset
func GenerateDeployObject(opt *WithCtrlOption) runtime.Object {
	var containers = make([]corev1.Container, len(opt.Containers))
	var initContainers = make([]corev1.Container, len(opt.InitContainers))

	for i, initCnOpt := range opt.InitContainers {
		initContainers[i] = corev1.Container{
			Name:            initCnOpt.ContainerName,
			Image:           initCnOpt.Image,
			ImagePullPolicy: initCnOpt.ImagePullPolicy,
			Command:         initCnOpt.Command,
			Args:            initCnOpt.Args,
			Env:             append(DefaultCommonEnvs(), initCnOpt.Envs...),
			Resources:       initCnOpt.Resources,
			Ports:           initCnOpt.Ports,
			VolumeMounts:    initCnOpt.VolumeMounts,
			Lifecycle:       initCnOpt.LifeCycle,
		}
	}

	for j, CnOpt := range opt.Containers {
		containers[j] = corev1.Container{
			Name:            CnOpt.ContainerName,
			Image:           CnOpt.Image,
			ImagePullPolicy: CnOpt.ImagePullPolicy,
			Command:         CnOpt.Command,
			Args:            CnOpt.Args,
			Env:             append(DefaultCommonEnvs(), CnOpt.Envs...),
			Resources:       CnOpt.Resources,
			Ports:           CnOpt.Ports,
			VolumeMounts:    CnOpt.VolumeMounts,
			Lifecycle:       CnOpt.LifeCycle,
		}
	}

	switch opt.ObjectKind {
	case DeploymentKind:
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:            opt.Name + DeploymentSuffix,
				Namespace:       opt.Namespace,
				OwnerReferences: opt.OwnerReferences,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: opt.Replicas,
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": opt.LabelSelector}},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      map[string]string{"app": opt.LabelSelector},
						Annotations: opt.Annotations,
					},
					Spec: corev1.PodSpec{
						Affinity:       opt.Spec.Affinity,
						NodeSelector:   opt.Spec.NodeSelector,
						DNSPolicy:      corev1.DNSClusterFirst,
						InitContainers: initContainers,
						Containers:     containers,
						Volumes:        opt.Volumes,
					},
				},
			},
		}

	case StatefulsetKind:
		return &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:            opt.Name + StatefulsetSuffix,
				Namespace:       opt.Namespace,
				OwnerReferences: opt.OwnerReferences,
			},
			Spec: appsv1.StatefulSetSpec{
				ServiceName: opt.Name + ServiceSuffix,
				Replicas:    opt.Replicas,
				Selector:    &metav1.LabelSelector{MatchLabels: map[string]string{"app": opt.LabelSelector}},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      map[string]string{"app": opt.LabelSelector},
						Annotations: opt.Annotations,
					},
					Spec: corev1.PodSpec{
						Affinity:       opt.Spec.Affinity,
						NodeSelector:   opt.Spec.NodeSelector,
						DNSPolicy:      corev1.DNSClusterFirst,
						SchedulerName:  opt.Spec.Scheduler,
						InitContainers: initContainers,
						Containers:     containers,
						Volumes:        opt.Volumes,
					},
				},
			},
		}

	case CronJobKind:
		return &batchv1beta1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:            opt.Name + ServerRdbBackupSuffix,
				Namespace:       opt.Namespace,
				OwnerReferences: opt.OwnerReferences,
			},
			Spec: batchv1beta1.CronJobSpec{
				ConcurrencyPolicy:          opt.BackupCronJob.ConcurrencyPolicy,
				Schedule:                   opt.BackupCronJob.Schedule,
				StartingDeadlineSeconds:    opt.BackupCronJob.StartingDeadlineSeconds,
				SuccessfulJobsHistoryLimit: opt.BackupCronJob.SuccessfulJobsHistoryLimit,
				JobTemplate: batchv1beta1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Completions: opt.BackupCronJob.Completions,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								RestartPolicy:  corev1.RestartPolicyNever,
								Volumes:        opt.Volumes,
								InitContainers: initContainers,
								Containers:     containers,
							},
						},
					},
				},
			},
		}

	default:
		panic("请输入正确的ObjectKind")
	}
}

// InitCheckDashboardCommand provide default startup command of the dashboard started-check
func InitCheckDashboardCommand() []string {
	return []string{
		"/bin/sh",
		"-c",
		"curl -s http://$DASHBOARD_ADDR/api/topom/model",
	}
}

// CodisServerBackupCommand provide default startup command of codis-server backup
func CodisServerBackupCommand() []string {
	return []string{
		"python",
		"/data/service/codis-backup/backup.py",
		">>",
		"/codis/log/$(POD_NAMESPACE)_$(POD_NAME).log",
		"2>&1",
	}
}

// InitCheckZookeeperCommand provide default startup command of the zookeeper started-check
func InitCheckZookeeperCommand() []string {
	return []string{
		"/bin/sh",
		"-c",
		"telnet $ZK_ADDR 2181",
	}
}

// Dashboard service name
const DashboardSrvName = "codis-dashboard-service.default.svc.cluster.local"

// GenerateDashboardAddr returns the service address of the cluster dashboard
func GenerateDashboardAddr(name string, port int32) string {
	return fmt.Sprintf("%s-%s:%v", name, DashboardSrvName, port)
}

var lastAppliedConfigKey = infrav1.GroupVersion.String() + "-last-applied-config"

// SetStsLastAppliedConfig save last StatefulSet configuration
// When an update event is received, the Reconcile function compares the configuration changes accordingly
func SetStsLastAppliedConfig(sts *appsv1.StatefulSet) error {
	stsSpec, err := jsoniter.MarshalToString(sts.Spec)
	if err != nil {
		return err
	}
	if sts.Annotations == nil {
		sts.Annotations = map[string]string{}
	}
	sts.Annotations[lastAppliedConfigKey] = stsSpec
	return nil
}

// SetServiceLastAppliedConfig save last Service configuration
// When an update event is received, the Reconcile function compares the configuration changes accordingly
func SetServiceLastAppliedConfig(svc *corev1.Service) error {

	for k := range svc.Annotations {
		if k != "service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id" &&
			k != "service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners" &&
			k != "service.beta.kubernetes.io/backend-type" &&
			k != "service.beta.kubernetes.io/alibaba-cloud-loadbalancer-protocol-port" {
			delete(svc.Annotations, k)
		}
	}

	svcData, err := jsoniter.MarshalToString(*svc)
	if err != nil {
		return err
	}
	if svc.Annotations == nil {
		svc.Annotations = map[string]string{}
	}
	svc.Annotations[lastAppliedConfigKey] = svcData
	return nil
}

// SetDeploymentLastAppliedConfig save last Deployment configuration
// When an update event is received, the Reconcile function compares the configuration changes accordingly
func SetDeploymentLastAppliedConfig(deploy *appsv1.Deployment) error {
	deploySpec, err := jsoniter.MarshalToString(deploy.Spec)
	if err != nil {
		return err
	}
	if deploy.Annotations == nil {
		deploy.Annotations = map[string]string{}
	}
	deploy.Annotations[lastAppliedConfigKey] = deploySpec
	return nil
}

// SetConfigMapAppliedConfig save last ConfigMap configuration
// When an update event is received, the Reconcile function compares the configuration changes accordingly
func SetConfigMapAppliedConfig(deploy *corev1.ConfigMap) error {
	deploySpec, err := jsoniter.MarshalToString(deploy.Data)
	if err != nil {
		return err
	}
	if deploy.Annotations == nil {
		deploy.Annotations = map[string]string{}
	}
	deploy.Annotations[lastAppliedConfigKey] = deploySpec
	return nil
}

// SetPVCAppliedConfig save last PersistentVolumeClaim configuration
// When an update event is received, the Reconcile function compares the configuration changes accordingly
func SetPVCAppliedConfig(pvc *corev1.PersistentVolumeClaim) error {
	pvcSpec, err := jsoniter.MarshalToString(pvc.Spec)
	if err != nil {
		return err
	}
	if pvc.Annotations == nil {
		pvc.Annotations = map[string]string{}
	}
	pvc.Annotations[lastAppliedConfigKey] = pvcSpec
	return nil
}

// SetCronJobAppliedConfig save last CronJob configuration
// When an update event is received, the Reconcile function compares the configuration changes accordingly
func SetCronJobAppliedConfig(pvc *batchv1beta1.CronJob) error {
	jobSpec, err := jsoniter.MarshalToString(pvc.Spec)
	if err != nil {
		return err
	}
	if pvc.Annotations == nil {
		pvc.Annotations = map[string]string{}
	}
	pvc.Annotations[lastAppliedConfigKey] = jobSpec
	return nil
}

// StatefulSetEqual is Used to compare whether the new and old StatefulSet are equal
// If the comparison result is false, the update operation will be triggered
func StatefulSetEqual(new, old *appsv1.StatefulSet) (bool, error) {
	oldSpec := appsv1.StatefulSetSpec{}
	if lastAppliedConfig, ok := old.Annotations[lastAppliedConfigKey]; ok {
		err := jsoniter.Unmarshal([]byte(lastAppliedConfig), &oldSpec)
		if err != nil {
			return false, fmt.Errorf("ns:%s,name:%s,unmarshal statefulset err is %v", old.GetNamespace(), old.GetName(), err)
		}
		return apiequality.Semantic.DeepEqual(oldSpec, new.Spec), nil
	}
	return false, nil
}

// ServiceEqual is Used to compare whether the new and old Service are equal
// If the comparison result is false, the update operation will be triggered
func ServiceEqual(new, old *corev1.Service) (bool, error) {
	oldSvc := corev1.Service{}

	if lastAppliedConfig, ok := old.Annotations[lastAppliedConfigKey]; ok {
		err := jsoniter.Unmarshal([]byte(lastAppliedConfig), &oldSvc)
		if err != nil {
			return false, fmt.Errorf("ns:%s,name:%s,unmarshal deployment err is %v", old.GetNamespace(), old.GetName(), err)
		}
		return apiequality.Semantic.DeepEqual(oldSvc.Spec, new.Spec) &&
			apiequality.Semantic.DeepEqual(oldSvc.Annotations, new.Annotations), nil
	}
	return false, nil
}

// DeploymentEqual is Used to compare whether the new and old Deployment are equal
// If the comparison result is false, the update operation will be triggered
func DeploymentEqual(new, old *appsv1.Deployment) (bool, error) {
	oldSpec := appsv1.DeploymentSpec{}
	if lastAppliedConfig, ok := old.Annotations[lastAppliedConfigKey]; ok {
		err := jsoniter.Unmarshal([]byte(lastAppliedConfig), &oldSpec)
		if err != nil {
			return false, err
		}
		return apiequality.Semantic.DeepEqual(oldSpec, new.Spec), nil
	}
	return false, nil
}

// ConfigMapEqual is Used to compare whether the new and old ConfigMap are equal
// If the comparison result is false, the update operation will be triggered
func ConfigMapEqual(new, old *corev1.ConfigMap) (bool, error) {
	oldData := corev1.ConfigMap{}
	if lastAppliedConfig, ok := old.Annotations[lastAppliedConfigKey]; ok {
		err := jsoniter.Unmarshal([]byte(lastAppliedConfig), &oldData)
		if err != nil {
			return false, err
		}
		return apiequality.Semantic.DeepEqual(oldData, new.Data), nil
	}
	return false, nil
}

// PVCEqual is Used to compare whether the new and old PersistentVolumeClaim are equal
// If the comparison result is false, the update operation will be triggered
func PVCEqual(new, old *corev1.PersistentVolumeClaim) (bool, error) {
	oldSpec := corev1.PersistentVolumeClaimSpec{}
	if lastAppliedConfig, ok := old.Annotations[lastAppliedConfigKey]; ok {
		err := jsoniter.Unmarshal([]byte(lastAppliedConfig), &oldSpec)
		if err != nil {
			return false, err
		}
		return apiequality.Semantic.DeepEqual(oldSpec, new.Spec), nil
	}
	return false, nil
}

// CronJobEqual is Used to compare whether the new and old CronJob are equal
// If the comparison result is false, the update operation will be triggered
func CronJobEqual(new, old *batchv1beta1.CronJob) (bool, error) {
	oldSpec := batchv1beta1.CronJobSpec{}
	if lastAppliedConfig, ok := old.Annotations[lastAppliedConfigKey]; ok {
		err := jsoniter.Unmarshal([]byte(lastAppliedConfig), &oldSpec)
		if err != nil {
			return false, err
		}
		return apiequality.Semantic.DeepEqual(oldSpec, new.Spec), nil
	}
	return false, nil
}
