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

package manager

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github/oarfah/codis-operater/api/v1"
	"github/oarfah/codis-operater/controllers/util"
)

// CodisServerReconciler reconciles a CodisServer object
type CodisServerReconciler struct {
	ctx    context.Context
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	C      *infrav1.CodisCluster
}

func NewCodisServerMgt(c client.Client, log logr.Logger, scheme *runtime.Scheme) *CodisServerReconciler {
	return &CodisServerReconciler{Client: c, Log: log, Scheme: scheme}
}

func (r *CodisServerReconciler) Reconcile(cc *infrav1.CodisCluster) (ctrl.Result, error) {
	r.ctx = context.Background()
	r.C = cc

	if err := r.createOrUpdateService(); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.createOrUpdatePVC(); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.mustBackupRedisRdbCrontab(); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.createOrUpdateRedisConf(); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.createOrUpdateStatefulSet(); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CodisServerReconciler) createOrUpdateService() error {
	newSvc := util.GenerateService(r.defaultSvcOption())

	// fetch latest stat, update
	var found corev1.Service
	if err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.C.Name + "-" + util.CodisRedis + util.ServiceSuffix, Namespace: r.C.Namespace}, &found); err != nil && errors.IsNotFound(err) {
		r.Log.Info("creating service of codisServer...")
		err = util.SetServiceLastAppliedConfig(newSvc)
		if err != nil {
			return err
		}
		return r.Client.Create(r.ctx, newSvc)
	} else if err != nil {
		return err
	}

	if ok, err := util.ServiceEqual(newSvc, &found); err != nil {
		return err
	} else if !ok {
		r.Log.Info("updating service of codisServer...")
		err := util.SetServiceLastAppliedConfig(newSvc)
		if err != nil {
			return err
		}
		newSvc.Spec.ClusterIP = found.Spec.ClusterIP
		newSvc.ResourceVersion = found.ResourceVersion
		return r.Client.Update(r.ctx, newSvc)
	}

	return nil
}

func (r *CodisServerReconciler) createOrUpdateStatefulSet() error {
	newDeploy := util.GenerateDeployObject(r.defaultCtrlOption())

	// fetch latest status
	var found appsv1.StatefulSet
	if err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.C.Name + "-" + util.CodisRedis + util.StatefulsetSuffix, Namespace: r.C.Namespace}, &found); err != nil && errors.IsNotFound(err) {
		r.Log.Info("creating statefulset of codisServer...")
		err := util.SetStsLastAppliedConfig(newDeploy.(*appsv1.StatefulSet))
		if err != nil {
			return err
		}
		return r.Client.Create(r.ctx, newDeploy)
	} else if err != nil {
		return err
	}

	if ok, err := util.StatefulSetEqual(newDeploy.(*appsv1.StatefulSet), &found); err != nil {
		return err
	} else if !ok {
		r.Log.Info("updating service of codisDashboard...")
		err := util.SetStsLastAppliedConfig(newDeploy.(*appsv1.StatefulSet))
		if err != nil {
			return err
		}
		return r.Client.Update(r.ctx, newDeploy)
	}

	return nil
}

func (r *CodisServerReconciler) createOrUpdatePVC() error {
	if r.C.Spec.RedisSpec.StorageClassName == "" {
		return nil
	}
	var pvc = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.C.Name + "-" + util.CodisRedis + util.PVCSuffix,
			Namespace: r.C.Namespace,
			Annotations: map[string]string{
				"pv-name-created": r.C.Name + util.CodisRedis + util.PVCSuffix,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			StorageClassName: &r.C.Spec.RedisSpec.StorageClassName,
			Resources:        r.C.Spec.RedisSpec.StorageRequires,
		},
	}

	// fetch latest status
	var found corev1.PersistentVolumeClaim
	if err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.C.Name + "-" + util.CodisRedis + util.PVCSuffix, Namespace: r.C.Namespace}, &found); err != nil && errors.IsNotFound(err) {
		r.Log.Info("creating pvc of codisServer...")
		err := util.SetPVCAppliedConfig(pvc)
		if err != nil {
			return err
		}
		return r.Client.Create(r.ctx, pvc)
	} else if err != nil {
		return err
	}

	// update pvc
	if ok, err := util.PVCEqual(pvc, &found); err != nil {
		return err
	} else if !ok {
		r.Log.Info("updating storageClass of codisServer...")
		err := util.SetPVCAppliedConfig(pvc)
		if err != nil {
			return err
		}
		return r.Client.Update(r.ctx, pvc)
	}

	return nil
}

func (r *CodisServerReconciler) createOrUpdateRedisConf() error {
	if r.C.Spec.RedisSpec.RedisConf == nil {
		r.C.Spec.RedisSpec.RedisConf = util.DefaultRedisConfiguration()
	}

	var conf = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.C.Name + "-" + util.CodisRedis + util.ServerConfSuffix,
			Namespace: r.C.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(r.C, infrav1.GroupVersion.WithKind(r.C.Kind)),
			},
		},
		Data: r.C.Spec.RedisSpec.RedisConf,
	}

	// fetch latest status
	var found corev1.ConfigMap
	if err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.C.Name + "-" + util.CodisRedis + util.ServerConfSuffix, Namespace: r.C.Namespace}, &found); err != nil && errors.IsNotFound(err) {
		r.Log.Info("creating configmap of codisServer...")
		err := util.SetConfigMapAppliedConfig(conf)
		if err != nil {
			return err
		}
		return r.Client.Create(r.ctx, conf)
	} else if err != nil {
		return err
	}

	if ok, err := util.ConfigMapEqual(conf, &found); err != nil {
		return err
	} else if !ok {
		r.Log.Info("updating storageClass of codisServer...")
		err := util.SetConfigMapAppliedConfig(conf)
		if err != nil {
			return err
		}
		return r.Client.Update(r.ctx, conf)
	}

	return nil
}

func (r *CodisServerReconciler) mustBackupRedisRdbCrontab() error {
	newJob := util.GenerateDeployObject(r.defaultCronJobOption())
	// fetch latest status
	var found batchv1beta1.CronJob
	if err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.C.Name + "-" + util.CodisRedis + util.ServerRdbBackupSuffix, Namespace: r.C.Namespace}, &found); err != nil && errors.IsNotFound(err) {
		r.Log.Info("creating backup job of codisServer...")
		err := util.SetCronJobAppliedConfig(newJob.(*batchv1beta1.CronJob))
		if err != nil {
			return err
		}
		return r.Client.Create(r.ctx, newJob)
	} else if err != nil {
		return err
	}

	if ok, err := util.CronJobEqual(newJob.(*batchv1beta1.CronJob), &found); err != nil {
		return err
	} else if !ok {
		r.Log.Info("updating backup job of codisServer...")
		err := util.SetCronJobAppliedConfig(newJob.(*batchv1beta1.CronJob))
		if err != nil {
			return err
		}
		return r.Client.Update(r.ctx, newJob)
	}

	return nil
}

func (r *CodisServerReconciler) defaultCtrlOption() *util.WithCtrlOption {
	var lifeCycle *corev1.Lifecycle
	if r.C.Spec.RedisSpec.LifeCycle == nil {
		lifeCycle = util.DefaultServerLifeCycle()
	}

	volumes := []corev1.Volume{
		{Name: util.CodisRedis, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: r.C.Spec.LogPath}}},
		{Name: util.CodisConf, VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: r.C.Name + "-" + util.CodisRedis + util.ServerConfSuffix}, Items: []corev1.KeyToPath{{Key: "redis.conf", Path: "redis.conf"}}}}},
	}

	volumesMounts := []corev1.VolumeMount{
		{Name: util.CodisRedis, MountPath: util.CommonLogPath, SubPathExpr: fmt.Sprintf("%v/$(%v)", util.CodisRedis, "POD_NAME")},
		{Name: util.CodisConf, MountPath: util.CodisServerConfPath},
	}

	if r.C.Spec.RedisSpec.StorageClassName != "" {
		volumes = append(volumes, corev1.Volume{Name: util.CodisRedisPVC, VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: r.C.Name + "-" + util.CodisRedis + util.PVCSuffix}}})
		volumesMounts = append(volumesMounts, corev1.VolumeMount{Name: util.CodisRedisPVC, MountPath: util.CodisServerDataPath, SubPathExpr: fmt.Sprintf("$(%v)", "POD_NAME")})
	}

	return &util.WithCtrlOption{
		ObjectKind: util.StatefulsetKind,
		Meta: util.Meta{
			Name:      r.C.Name + "-" + util.CodisRedis,
			Namespace: r.C.Namespace,
			Annotations: map[string]string{
				"prometheus.io/path":   "/metrics",
				"prometheus.io/port":   strconv.Itoa(int(util.DefaultServerExporterPort)),
				"prometheus.io/scrape": "true",
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(r.C, infrav1.GroupVersion.WithKind(r.C.Kind)),
			},
		},
		Spec: util.Spec{
			Replicas:      r.C.Spec.RedisSpec.Replicas,
			Affinity:      r.C.Spec.RedisSpec.Affinity,
			NodeSelector:  r.C.Spec.RedisSpec.NodeSelector,
			LabelSelector: r.C.Namespace + "-" + r.C.Name + "-" + util.CodisRedis,
			Scheduler:     r.C.Spec.RedisSpec.Scheduler,
			Volumes:       volumes,
			InitContainers: []util.ContainerSpec{
				{
					ContainerName:   "check-dashboard-ready",
					Image:           util.CurlImg,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         util.InitCheckDashboardCommand(),
					Envs: []corev1.EnvVar{
						{Name: "DASHBOARD_ADDR", Value: util.GenerateDashboardAddr(r.C.Name, r.C.Spec.DashboardSpec.ServicePort)},
					},
				},
			},
			Containers: []util.ContainerSpec{
				{
					ContainerName:   util.CodisRedis + "-exporter",
					Image:           r.C.Spec.RedisSpec.ExporterImage,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Envs: []corev1.EnvVar{
						{Name: "REDIS_ADDR", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}}},
					},
					Resources: r.C.Spec.RedisSpec.CodisServerExporterRequires,
					Ports: []corev1.ContainerPort{
						{ContainerPort: util.DefaultServerExporterPort, Name: "exporter"},
					},
				},
				{
					ContainerName:   util.CodisRedis,
					Image:           r.C.Spec.CodisImage,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         util.DefaultServerCommand(),
					Args:            util.DefaultServerArgs(),
					LifeCycle:       lifeCycle,
					Envs: []corev1.EnvVar{
						{Name: "SERVER_REPLICA", Value: strconv.Itoa(r.C.Spec.RedisSpec.CodisServerNumsPerGroup)},
						{Name: "DASHBOARD_ADDR", Value: util.GenerateDashboardAddr(r.C.Name, r.C.Spec.DashboardSpec.ServicePort)},
						{Name: "HOST_NAME", Value: fmt.Sprintf("$(POD_NAME).%s", r.C.Name+"-"+util.CodisRedis+util.ServiceSuffix)},
					},
					Resources: r.C.Spec.RedisSpec.CodisServerRequires,
					Ports: []corev1.ContainerPort{
						{ContainerPort: util.DefaultServerPort, Name: "codis-server"},
					},
					VolumeMounts: volumesMounts,
				},
			},
		},
	}
}

func (r *CodisServerReconciler) defaultCronJobOption() *util.WithCtrlOption {
	var Completions = new(int32)
	*Completions = 1
	return &util.WithCtrlOption{
		ObjectKind: util.CronJobKind,
		Meta: util.Meta{
			Name:      r.C.Name + "-" + util.CodisRedis,
			Namespace: r.C.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(r.C, infrav1.GroupVersion.WithKind(r.C.Kind)),
			},
		},
		Spec: util.Spec{
			BackupCronJob: util.BackupCronJobSpec{
				ConcurrencyPolicy:          r.C.Spec.RedisSpec.BackupCronJob.ConcurrencyPolicy,
				Schedule:                   r.C.Spec.RedisSpec.BackupCronJob.Schedule,
				StartingDeadlineSeconds:    r.C.Spec.RedisSpec.BackupCronJob.StartingDeadlineSeconds,
				SuccessfulJobsHistoryLimit: r.C.Spec.RedisSpec.BackupCronJob.SuccessfulJobsHistoryLimit,
				Completions:                Completions,
			},
			Volumes: []corev1.Volume{
				{Name: util.CodisBackup, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: r.C.Spec.LogPath}}},
			},
			Containers: []util.ContainerSpec{
				{
					ContainerName: r.C.Name,
					Image:         r.C.Spec.RedisSpec.BackupImage,
					Command:       util.CodisServerBackupCommand(),
					VolumeMounts: []corev1.VolumeMount{
						{Name: util.CodisBackup, MountPath: util.CommonLogPath, SubPath: util.CodisBackup},
					},
				},
			},
		},
	}
}

func (r *CodisServerReconciler) defaultSvcOption() *util.WithSvcOptions {
	return &util.WithSvcOptions{
		Name:        r.C.Name + "-" + util.CodisRedis,
		Namespace:   r.C.Namespace,
		ServicePort: util.DefaultServerPort,
		ServiceType: corev1.ServiceTypeClusterIP,
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(r.C, infrav1.GroupVersion.WithKind(r.C.Kind)),
		},
	}
}
