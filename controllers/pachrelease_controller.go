/*


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

	"github.com/go-logr/logr"
	opsv1 "github.com/pachyderm/pachyderm-operator/api/v1"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// PachReleaseReconciler reconciles a PachRelease object
type PachReleaseReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ops.pachyderm.io,resources=pachreleases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ops.pachyderm.io,resources=pachreleases/status,verbs=get;update;patch

func (r *PachReleaseReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pachrelease", req.NamespacedName)

	// your logic here

	var opts *assets.AssetOpts
	opts = &assets.AssetOpts{}
	opts.BlockCacheSize = "0G"
	opts.EtcdNodes = 1
	opts.Namespace = req.NamespacedName.Namespace
	opts.DashImage = "pachyderm/dash:0.5.48"

	var pachRelease opsv1.PachRelease
	if err := r.Get(ctx, req.NamespacedName, &pachRelease); err != nil {
		//log.Info(err, "unable to fetch PachRelease")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	//TODO Look into ctrl.CreateOrUpdate

	dashService := assets.DashService(opts)
	err := r.Get(ctx, types.NamespacedName{Name: dashService.Name, Namespace: req.NamespacedName.Namespace}, &v1.Service{})
	if err != nil && errors.IsNotFound(err) {

		if err := controllerutil.SetControllerReference(&pachRelease, dashService, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Creating Service: ", dashService.Namespace, dashService.Name)

		err = r.Create(ctx, dashService)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	}

	dashDeployment := assets.DashDeployment(opts)
	err = r.Get(ctx, types.NamespacedName{Name: dashDeployment.Name, Namespace: req.NamespacedName.Namespace}, &apps.Deployment{})
	if err != nil && errors.IsNotFound(err) {

		if err := controllerutil.SetControllerReference(&pachRelease, dashDeployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Creating Deployment: ", dashDeployment.Namespace, dashDeployment.Name)

		err = r.Create(ctx, dashDeployment)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PachReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.PachRelease{}).
		Complete(r)
}
