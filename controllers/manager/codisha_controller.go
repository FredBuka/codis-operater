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

// CodisHaReconciler reconciles a CodisHa object
type CodisHaReconciler struct {
	ctx    context.Context
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	C      *infrav1.CodisCluster
}

// NewCodisHaMgt function generate reconciler of codis-ha
func NewCodisHaMgt(c client.Client, log logr.Logger, scheme *runtime.Scheme) *CodisHaReconciler {
	return &CodisHaReconciler{Client: c, Log: log, Scheme: scheme}
}

func (r *CodisHaReconciler) Reconcile(cc *infrav1.CodisCluster) (ctrl.Result, error) {
	r.ctx = context.Background()
	r.C = cc

	if err := r.createOrUpdateDeployment(); err != nil {
		r.Log.Error(err, "createOrUpdate deployment fail")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CodisHaReconciler) createOrUpdateDeployment() error {
	newDeploy := util.GenerateDeployObject(r.defaultCtrlOption())

	// fetch latest status
	var found appsv1.Deployment
	if err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.C.Name + "-" + util.CodisHa + util.DeploymentSuffix, Namespace: r.C.Namespace}, &found); err != nil && errors.IsNotFound(err) {
		r.Log.Info("creating deployment of codisHa...")
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
		r.Log.Info("updating service of codisHa...")
		err := util.SetDeploymentLastAppliedConfig(newDeploy.(*appsv1.Deployment))
		if err != nil {
			return err
		}
		return r.Client.Update(r.ctx, newDeploy)
	}

	return nil
}

func (r *CodisHaReconciler) defaultCtrlOption() *util.WithCtrlOption {
	var lifeCycle *corev1.Lifecycle
	if r.C.Spec.HaSpec.LifeCycle == nil {
		lifeCycle = util.DefaultHaLifeCycle()
	}

	return &util.WithCtrlOption{
		ObjectKind: util.DeploymentKind,
		Meta: util.Meta{
			Name:      r.C.Name + "-" + util.CodisHa,
			Namespace: r.C.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(r.C, infrav1.GroupVersion.WithKind(r.C.Kind)),
			},
		},
		Spec: util.Spec{
			Replicas:      r.C.Spec.HaSpec.Replicas,
			Affinity:      r.C.Spec.HaSpec.Affinity,
			NodeSelector:  r.C.Spec.HaSpec.NodeSelector,
			LabelSelector: r.C.Namespace + "-" + r.C.Name + "-" + util.CodisHa,
			Volumes: []corev1.Volume{
				{Name: util.CodisHa, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: r.C.Spec.LogPath}}},
			},
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
					ContainerName:   util.CodisHa,
					Image:           r.C.Spec.CodisImage,
					ImagePullPolicy: corev1.PullIfNotPresent,
					LifeCycle:       lifeCycle,
					Command:         util.DefaultHaCommand(),
					Args:            util.DefaultHaArgs(),
					Envs: []corev1.EnvVar{
						{Name: "DASHBOARD_ADDR", Value: util.GenerateDashboardAddr(r.C.Name, r.C.Spec.DashboardSpec.ServicePort)},
					},
					Resources: r.C.Spec.HaSpec.Requires,
					VolumeMounts: []corev1.VolumeMount{
						{Name: util.CodisHa, MountPath: util.CommonLogPath, SubPath: util.CodisHa},
					},
				},
			},
		},
	}
}
