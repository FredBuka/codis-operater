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

// CodisDashboardReconciler reconciles a CodisDashboard object
type CodisDashboardReconciler struct {
	ctx    context.Context
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	C      *infrav1.CodisCluster
}

// NewCodisDashboardMgt function generate reconciler of codis-dashboard
func NewCodisDashboardMgt(c client.Client, log logr.Logger, scheme *runtime.Scheme) *CodisDashboardReconciler {
	return &CodisDashboardReconciler{Client: c, Log: log, Scheme: scheme}
}

func (r *CodisDashboardReconciler) Reconcile(cc *infrav1.CodisCluster) (ctrl.Result, error) {
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

func (r *CodisDashboardReconciler) createOrUpdateService() error {
	newSvc := util.GenerateService(r.defaultSvcOption())

	// fetch latest status
	var found corev1.Service
	if err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.C.Name + "-" + util.CodisDashboard + util.ServiceSuffix, Namespace: r.C.Namespace}, &found); err != nil && errors.IsNotFound(err) {
		r.Log.Info("creating service of codisDashboard...")
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
		r.Log.Info("updating service of codisDashboard...")
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

func (r *CodisDashboardReconciler) createOrUpdateDeployment() error {
	newDeploy := util.GenerateDeployObject(r.defaultCtrlOption())

	// fetch latest status
	var found appsv1.Deployment
	if err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.C.Name + "-" + util.CodisDashboard + util.DeploymentSuffix, Namespace: r.C.Namespace}, &found); err != nil && errors.IsNotFound(err) {
		r.Log.Info("creating deployment of codisDashboard...")
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
		r.Log.Info("updating service of codisDashboard...")
		err := util.SetDeploymentLastAppliedConfig(newDeploy.(*appsv1.Deployment))
		if err != nil {
			return err
		}
		return r.Client.Update(r.ctx, newDeploy)
	}

	return nil
}

func (r *CodisDashboardReconciler) defaultSvcOption() *util.WithSvcOptions {
	return &util.WithSvcOptions{
		Name:        r.C.Name + "-" + util.CodisDashboard,
		Namespace:   r.C.Namespace,
		TargetPort:  util.DefaultDashboardPort,
		ServicePort: r.C.Spec.DashboardSpec.ServicePort,
		ServiceType: r.C.Spec.DashboardSpec.ServiceType,
		Instanceid:  r.C.Spec.DashboardSpec.SlbInstanceId,
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(r.C, infrav1.GroupVersion.WithKind(r.C.Kind)),
		},
	}
}

func (r *CodisDashboardReconciler) defaultCtrlOption() *util.WithCtrlOption {
	var lifeCycle *corev1.Lifecycle
	if r.C.Spec.DashboardSpec.LifeCycle == nil {
		lifeCycle = util.DefaultDashboardLifeCycle()
	}

	return &util.WithCtrlOption{
		ObjectKind: util.DeploymentKind,
		Meta: util.Meta{
			Name:      r.C.Name + "-" + util.CodisDashboard,
			Namespace: r.C.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(r.C, infrav1.GroupVersion.WithKind(r.C.Kind)),
			},
		},
		Spec: util.Spec{
			Replicas:      r.C.Spec.DashboardSpec.Replicas,
			Affinity:      r.C.Spec.DashboardSpec.Affinity,
			NodeSelector:  r.C.Spec.DashboardSpec.NodeSelector,
			LabelSelector: r.C.Namespace + "-" + r.C.Name + "-" + util.CodisDashboard,
			Volumes: []corev1.Volume{
				{Name: util.CodisDashboard, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: r.C.Spec.LogPath}}},
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
					ContainerName:   util.CodisDashboard,
					Image:           r.C.Spec.CodisImage,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         util.DefaultDashboardCommand(),
					Args:            util.DefaultDashboardArgs(),
					LifeCycle:       lifeCycle,
					Envs: []corev1.EnvVar{
						{Name: "ZK_ADDR", Value: r.C.Spec.ZkAddr},
						{Name: "PRODUCT_NAME", Value: r.C.Spec.ProductName},
					},
					Resources: r.C.Spec.DashboardSpec.Requires,
					Ports: []corev1.ContainerPort{
						{ContainerPort: util.DefaultDashboardPort},
					},
					VolumeMounts: []corev1.VolumeMount{
						{Name: util.CodisDashboard, MountPath: util.CommonLogPath, SubPath: util.CodisDashboard},
					},
				},
			},
		},
	}
}
