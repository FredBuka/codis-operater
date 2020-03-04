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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
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

// CodisFeReconciler reconciles a CodisFe object
type CodisFeReconciler struct {
	ctx    context.Context
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	C      *infrav1.CodisCluster
}

// NewCodisFeMgt function generate reconciler of codis-fe
func NewCodisFeMgt(c client.Client, log logr.Logger, scheme *runtime.Scheme) *CodisFeReconciler {
	return &CodisFeReconciler{Client: c, Log: log, Scheme: scheme}
}

func (r *CodisFeReconciler) Reconcile(cc *infrav1.CodisCluster) (ctrl.Result, error) {
	r.ctx = context.Background()
	r.C = cc

	if err := r.createOrUpdateService(); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.createOrUpdateDeployment(); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CodisFeReconciler) createOrUpdateService() error {
	newSvc := util.GenerateService(r.defaultSvcOption())

	// fetch latest status
	var found corev1.Service
	if err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.C.Name + "-" + util.CodisFe + util.ServiceSuffix, Namespace: r.C.Namespace}, &found); err != nil && errors.IsNotFound(err) {
		r.Log.Info("creating service of codisFe...")
		err := util.SetServiceLastAppliedConfig(newSvc)
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
		r.Log.Info("updating service of codisFe...")
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

func (r *CodisFeReconciler) createOrUpdateDeployment() error {
	newDeploy := util.GenerateDeployObject(r.defaultCtrlOption())

	// fetch latest stats
	var found appsv1.Deployment
	if err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.C.Name + "-" + util.CodisFe + util.DeploymentSuffix, Namespace: r.C.Namespace}, &found); err != nil && errors.IsNotFound(err) {
		r.Log.Info("creating deployment of codisFe...")
		err := util.SetDeploymentLastAppliedConfig(newDeploy.(*appsv1.Deployment))
		if err != nil {
			return err
		}
		return r.Client.Create(r.ctx, newDeploy)
	} else if err != nil {
		return err
	}

	if ok, err := util.DeploymentEqual(newDeploy.(*appsv1.Deployment), &found); err != nil {
		return err
	} else if !ok {
		r.Log.Info("updating deployment of codisFe...")
		err := util.SetDeploymentLastAppliedConfig(newDeploy.(*appsv1.Deployment))
		if err != nil {
			return err
		}
		return r.Client.Update(r.ctx, newDeploy)
	}

	return nil
}

func (r *CodisFeReconciler) defaultCtrlOption() *util.WithCtrlOption {
	var lifeCycle *corev1.Lifecycle
	if r.C.Spec.FeSpec.LifeCycle == nil {
		lifeCycle = util.DefaultFeLifeCycle()
	}

	return &util.WithCtrlOption{
		ObjectKind: util.DeploymentKind,
		Meta: util.Meta{
			Name:      r.C.Name + "-" + util.CodisFe,
			Namespace: r.C.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(r.C, infrav1.GroupVersion.WithKind(r.C.Kind)),
			},
		},
		Spec: util.Spec{
			Replicas:      r.C.Spec.FeSpec.Replicas,
			Affinity:      r.C.Spec.FeSpec.Affinity,
			NodeSelector:  r.C.Spec.FeSpec.NodeSelector,
			LabelSelector: r.C.Namespace + "-" + r.C.Name + "-" + util.CodisFe,
			Volumes: []corev1.Volume{
				{Name: util.CodisFe, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: r.C.Spec.LogPath}}},
			},
			InitContainers: []util.ContainerSpec{
				{
					ContainerName:   "check-ready",
					Image:           util.TelnetImg,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         util.InitCheckZookeeperCommand(),
					Envs: []corev1.EnvVar{
						{Name: "ZK_ADDR", Value: r.C.Spec.ZkAddr},
					},
				},
			},
			Containers: []util.ContainerSpec{
				{
					ContainerName:   util.CodisFe,
					Image:           r.C.Spec.CodisImage,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         util.DefaultFeCommand(),
					Args:            util.DefaultFeArgs(),
					LifeCycle:       lifeCycle,
					Envs: []corev1.EnvVar{
						{Name: "ZK_ADDR", Value: r.C.Spec.ZkAddr},
					},
					Resources: r.C.Spec.FeSpec.Requires,
					Ports: []corev1.ContainerPort{
						{ContainerPort: util.DefaultFePort},
					},
					VolumeMounts: []corev1.VolumeMount{
						{Name: util.CodisFe, MountPath: util.CommonLogPath, SubPath: util.CodisFe},
					}},
			},
		},
	}
}

func (r *CodisFeReconciler) defaultSvcOption() *util.WithSvcOptions {
	return &util.WithSvcOptions{
		Name:        r.C.Name + "-" + util.CodisFe,
		Namespace:   r.C.Namespace,
		TargetPort:  util.DefaultFePort,
		ServicePort: r.C.Spec.FeSpec.ServicePort,
		ServiceType: r.C.Spec.FeSpec.ServiceType,
		Instanceid:  r.C.Spec.FeSpec.SlbInstanceId,
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(r.C, infrav1.GroupVersion.WithKind(r.C.Kind)),
		},
	}
}
