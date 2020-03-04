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

package controllers

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github/oarfah/codis-operater/api/v1"
	"github/oarfah/codis-operater/controllers/manager"
)

// CodisClusterReconciler reconciles a CodisCluster object
type CodisClusterReconciler struct {
	sync.Mutex
	ctx context.Context
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Dashboard *manager.CodisDashboardReconciler
	Fe        *manager.CodisFeReconciler
	Ha        *manager.CodisHaReconciler
	Proxy     *manager.CodisProxyReconciler
	Server    *manager.CodisServerReconciler
}

// +kubebuilder:rbac:groups=k8s.infra.cn,resources=codisclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=k8s.infra.cn,resources=codisclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *CodisClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.Lock()
	defer r.Unlock()

	r.ctx = context.Background()
	log := r.Log.WithValues("codiscluster", req.NamespacedName)

	var cluster infrav1.CodisCluster
	if err := r.Get(r.ctx, req.NamespacedName, &cluster); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "fetch latest status fail")
		return ctrl.Result{}, err
	}

	if _, err := r.Fe.Reconcile(&cluster); err != nil {
		log.Error(err, "operate fe")
	} else {
		log.Info("operate fe ok")
	}

	if _, err := r.Dashboard.Reconcile(&cluster); err != nil {
		log.Error(err, "operate dashboard")
	} else {
		log.Info("operate dashboard ok")
	}

	if _, err := r.Ha.Reconcile(&cluster); err != nil {
		log.Error(err, "operate ha")
	} else {
		log.Info("operate ha ok")
	}

	if _, err := r.Proxy.Reconcile(&cluster); err != nil {
		log.Error(err, "operate proxy")
	} else {
		log.Info("operate proxy ok")
	}

	if _, err := r.Server.Reconcile(&cluster); err != nil {
		log.Error(err, "operate server")
	} else {
		log.Info("operate server ok")
	}

	return ctrl.Result{}, nil
}

func (r *CodisClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.CodisCluster{}).
		Complete(r)
}
