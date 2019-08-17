package rayhead

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

var log = logf.Log.WithName("controller_rayhead")

// Add creates a new RayHead Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRayHead{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rayhead-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource RayHead
	err = c.Watch(&source.Kind{Type: &rayoperatorv1alpha1.RayHead{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner RayHead
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rayoperatorv1alpha1.RayHead{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileRayHead implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRayHead{}

// ReconcileRayHead reconciles a RayHead object
type ReconcileRayHead struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a RayHead object and makes changes based on the state read
// and what is in the RayHead.Spec
func (r *ReconcileRayHead) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling RayHead")

	// Fetch the RayHead instance
	rayhead := &rayoperatorv1alpha1.RayHead{}
	err := r.client.Get(context.TODO(), request.NamespacedName, rayhead)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("RayHead resource not found.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get RayHead")
		return reconcile.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: rayhead.Name, Namespace: rayhead.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForRayHead(rayhead)
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
	size := rayhead.Spec.Size
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

	// Update the ray head status with the pod names
	// List the pods for this rayhead's deployment
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForRayHead(rayhead.Name))
	listOps := &client.ListOptions{
		Namespace:     rayhead.Namespace,
		LabelSelector: labelSelector,
	}
	err = r.client.List(context.TODO(), listOps, podList)

	if err != nil {
		reqLogger.Error(err, "Failed to list pods", "RayHead.Namespace", rayhead.Namespace, "RayHead.Name", rayhead.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, rayhead.Status.Nodes) {
		rayhead.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), rayhead)
		if err != nil {
			reqLogger.Error(err, "Failed to update RayHead status")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// deploymentForMemcached returns a RayHead Deployment object
func (r *ReconcileRayHead) deploymentForRayHead(ro *rayoperatorv1alpha1.RayHead) *appsv1.Deployment {
	ls := labelsForRayHead(ro.Name)
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
						Name:    "ray-head",
						Command: []string{"/bin/bash", "-c", "--"},
						Args:    []string{"ray start --head --redis-port=6379 --redis-shard-ports=6380,6381 --object-manager-port=12345 --node-manager-port=12346 --node-ip-address=$MY_POD_IP --block"},
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 6379,
								Name:          "port-1",
							},
							{
								ContainerPort: 6380,
								Name:          "port-2",
							},
							{
								ContainerPort: 6381,
								Name:          "port-3",
							},
							{
								ContainerPort: 12346,
								Name:          "port-4",
							},
							{
								ContainerPort: 12345,
								Name:          "port-5",
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
	// Set ray head instance as the owner and controller
	if err := controllerutil.SetControllerReference(ro, dep, r.scheme); err != nil {
		reqLogger := log.WithValues("Request.Namespace")
		reqLogger.Error(err, "unable to set ray head controller")
	}
	return dep
}

// labelsForRayHead returns the labels for selecting the resources
// belonging to the given ray head CR name.
func labelsForRayHead(name string) map[string]string {
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
