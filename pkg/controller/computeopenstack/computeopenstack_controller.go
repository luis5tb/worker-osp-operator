package computeopenstack

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"path/filepath"

	computev1alpha1 "github.com/luis5tb/worker-osp-operator/pkg/apis/compute/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	machinev1beta1 "github.com/openshift/cluster-api/pkg/apis/machine/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/luis5tb/worker-osp-operator/pkg/render"
	"github.com/luis5tb/worker-osp-operator/pkg/apply"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var log = logf.Log.WithName("controller_computeopenstack")

// ManifestPath is the path to the manifest templates
var ManifestPath = "./bindata"


// Add creates a new ComputeOpenStack Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileComputeOpenStack{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("computeopenstack-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ComputeOpenStack
	err = c.Watch(&source.Kind{Type: &computev1alpha1.ComputeOpenStack{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ComputeOpenStack
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &computev1alpha1.ComputeOpenStack{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileComputeOpenStack implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileComputeOpenStack{}

// ReconcileComputeOpenStack reconciles a ComputeOpenStack object
type ReconcileComputeOpenStack struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ComputeOpenStack object and makes changes based on the state read
// and what is in the ComputeOpenStack.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileComputeOpenStack) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ComputeOpenStack")

	// Fetch the ComputeOpenStack instance
	instance := &computev1alpha1.ComputeOpenStack{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Create worker-osp, includes: machineset, machineconfigpool, machineconfigs and user-data secret
	// Fill all defaults explicitly
	data := render.MakeRenderData()
	data.Data["WorkerOspRole"] = instance.Spec.RoleName

	// get it from openshift-machine-api secrets (assumes worker-user-data)
	// Check if this Pod already exists
	userData := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "worker-user-data", Namespace: "openshift-machine-api"}, userData)
	if err != nil {
		return reconcile.Result{}, err
	}
	modifiedUserData := strings.Replace(string(userData.Data["userData"]), "worker", instance.Spec.RoleName, 1)
	encodedModifiedUserData := base64.StdEncoding.EncodeToString([]byte(modifiedUserData))
	data.Data["WorkerOspUserData"] = encodedModifiedUserData
	
	data.Data["K8sServiceIp"] = instance.Spec.K8sServiceIp
	data.Data["ApiIntIp"] = instance.Spec.ApiIntIp

	if instance.Spec.CorePinning == "" {
		data.Data["Pinning"] = false
	} else {
		data.Data["Pinning"] = true
		data.Data["CorePinning"] = instance.Spec.CorePinning
	}

	// Cluster Name hardcoded, should it be part of the CR? or should it be read from a ConfigMap?
	data.Data["ClusterName"] = instance.Spec.ClusterName
	//// get it from openshift-machine-api machineset (from other workers, assuming ostest-worker-0)
	workerMachineSet := &machinev1beta1.MachineSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.BaseWorkerMachineSetName, Namespace: "openshift-machine-api"}, workerMachineSet)
	if err != nil {
		return reconcile.Result{}, err
	}
	providerSpec := workerMachineSet.Spec.Template.Spec.ProviderSpec.Value.Raw
	providerData := make(map[string]map[string]interface{})
	err = json.Unmarshal(providerSpec, &providerData)
	if err != nil {
		return reconcile.Result{}, err
	}
	// workerMachineSet.Spec.Template.Spec.ProviderSpec.Value.Image.Url
	data.Data["RhcosImageUrl"] = providerData["image"]["url"]
	data.Data["Workers"] = instance.Spec.Workers

	//// Generate the objects
	objs := []*uns.Unstructured{}
	manifests, err := render.RenderDir(filepath.Join(ManifestPath, "worker-osp"), &data)
	if err != nil {
		log.Error(err, "Failed to render manifests : %v")
		return reconcile.Result{}, err
	}

	objs = append(objs, manifests...)

	// Apply the objects to the cluster
	for _, obj := range objs {
		// Set ComputeOpenStack instance as the owner and controller
		oref := metav1.NewControllerRef(instance, instance.GroupVersionKind())
		obj.SetOwnerReferences([]metav1.OwnerReference{*oref})			

		// Open question: should an error here indicate we will never retry?
		if err := apply.ApplyObject(context.TODO(), r.client, obj); err != nil {
			log.Error(err, "Failed to apply objects")	
			return reconcile.Result{}, err
		}
	}

	// create node-exporter (to have monitoring information)
	// cerate machine-config-daemon (to allow reconfigurations)
	// create multus pod (to set the node to ready)
	return reconcile.Result{}, nil
}
