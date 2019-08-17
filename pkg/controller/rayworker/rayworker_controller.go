package rayworker

import (
	"context"
	"k8s.io/apimachinery/pkg/labels"
	"reflect"

	rayoperatorv1alpha1 "github.com/silveryfu/ray-operator/pkg/apis/rayoperator/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_rayworker")

// Add creates a new RayWorker Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRayWorker{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rayworker-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource RayWorker
	err = c.Watch(&source.Kind{Type: &rayoperatorv1alpha1.RayWorker{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner RayWorker
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rayoperatorv1alpha1.RayWorker{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileRayWorker implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRayWorker{}

// ReconcileRayWorker reconciles a RayWorker object
type ReconcileRayWorker struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a RayWorker object and makes changes based on the state read
// and what is in the RayWorker.Spec
func (r *ReconcileRayWorker) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling RayWorker")

	// Fetch the RayWorker instance
	rayworker := &rayoperatorv1alpha1.RayWorker{}
	err := r.client.Get(context.TODO(), request.NamespacedName, rayworker)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("RayWorker resource not found.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get RayWorker")
		return reconcile.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: rayworker.Name, Namespace: rayworker.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForRayWorker(rayworker)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := rayworker.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	// Update the ray worker status with the pod names
	// List the pods for this rayworker's deployment
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForRayWorker(rayworker.Name))
	listOps := &client.ListOptions{
		Namespace:     rayworker.Namespace,
		LabelSelector: labelSelector,
	}
	err = r.client.List(context.TODO(), listOps, podList)

	if err != nil {
		reqLogger.Error(err, "Failed to list pods", "RayWorker.Namespace", rayworker.Namespace, "RayWorker.Name", rayworker.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, rayworker.Status.Nodes) {
		rayworker.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), rayworker)
		if err != nil {
			reqLogger.Error(err, "Failed to update RayWorker status")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// deploymentForRayWorker returns a RayWorker Deployment object
func (r *ReconcileRayWorker) deploymentForRayWorker(ro *rayoperatorv1alpha1.RayWorker) *appsv1.Deployment {
	ls := labelsForRayWorker(ro.Name)
	replicas := ro.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ro.Name,
			Namespace: ro.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   "silveryfu/ray-examples:latest",
						Name:    "ray-worker",
						Command: []string{"/bin/bash", "-c", "--"},
						Args:    []string{"ray start --node-ip-address=$MY_POD_IP --redis-address=$(python -c 'import socket;import sys; sys.stdout.write(socket.gethostbyname(\"ray-head\"));sys.stdout.flush()'):6379 --object-manager-port=12345 --node-manager-port=12346 --block"},
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 12345,
								Name:          "port-1",
							},
							{
								ContainerPort: 12346,
								Name:          "port-2",
							},
						},
						Env: []corev1.EnvVar{
							{
								Name: "MY_POD_IP",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										FieldPath: "status.podIP",
									},
								},
							},
						},
						// TODO: make resource request tunable
						//Resources: corev1.ResourceRequirements{
						//	Requests: corev1.ResourceList{
						//	},
						//},
					}},
				},
			},
		},
	}
	// Set ray worker instance as the owner and controller
	if err := controllerutil.SetControllerReference(ro, dep, r.scheme); err != nil {
		reqLogger := log.WithValues("Request.Namespace")
		reqLogger.Error(err, "unable to set ray worker controller")
	}
	return dep
}

// labelsForRayWorker returns the labels for selecting the resources
// belonging to the given ray worker CR name.
func labelsForRayWorker(name string) map[string]string {
	return map[string]string{"type": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
