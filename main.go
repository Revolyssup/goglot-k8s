/*
Copyright 2021.

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

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	goglotdevv1alpha1 "revolyssup/goglot-k8s/api/v1alpha1"
	"revolyssup/goglot-k8s/controllers"

	//+kubebuilder:scaffold:imports
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(goglotdevv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "fe4d0c6f.github.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.GlotpodReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Glotpod")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}
	// go listenForJobs()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func listenForPods(key string, value string, cs *kubernetes.Clientset) {
	infact := informers.NewSharedInformerFactory(cs, 30*time.Second)
	podInformer := infact.Core().V1().Pods()
	stopchan := make(chan struct{})
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// go func(obj interface{}) {
			lol, ok := obj.(*v1.Pod)
			if !ok {
				fmt.Println("SHIT ")
			}
			if lol.Labels[key] == value {
				stopchan <- struct{}{}
				go getLogs(lol, cs)
			}
			// }(obj)
		},
	})
	fmt.Println("LISETNIG FOR PODS")
	infact.Start(stopchan)
	fmt.Println("STOPPING LISTENG PODS")
	infact.WaitForCacheSync(wait.NeverStop)
}
func getLogs(pod *v1.Pod, cs *kubernetes.Clientset) {
	req := cs.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{
		Follow: true,
	})
	podLogs, er := req.Stream(context.Background())
	if er != nil {
		if errors.IsNotFound(er) {
			return
		}
		fmt.Println("SHIT could not get logs", er)
		getLogs(pod, cs)
		return
	}
	defer podLogs.Close()
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, podLogs)
	if err != nil {
		fmt.Println("SHIT could not get logs", err)
		return
	}
	str := buf.String()
	fmt.Println("LOGS ARE ", str)
	if str == "" {
		getLogs(pod, cs)
		return
	}
	err = cs.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("Could not delete pod because ", err.Error())
	}
	return
}

func listenForJobs(hashed string) {
	kubeconfig := flag.String("kc", "/home/ashish/.kube/config", "")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Println("Coulnd start pod watcher")
		return
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Coulnd start pod watcher 2")
		return
	}
	infact := informers.NewSharedInformerFactory(cs, 30*time.Second)
	jobInformer := infact.Batch().V1().Jobs()
	jobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// go func(obj interface{}) {
			lol, ok := obj.(*batchv1.Job)
			if !ok {
				fmt.Println("SHIT ")
			}

			go getJobLogs(lol, cs, hashed)
			// }(obj)
		},
	})
	fmt.Println("LISETNIG FOR Jobs")
	infact.Start(wait.NeverStop)
	infact.WaitForCacheSync(wait.NeverStop)
}

func getJobLogs(job *batchv1.Job, cs *kubernetes.Clientset, hashed string) {
	listenForPods("name", hashed, cs)
	err := cs.BatchV1().Jobs(job.Namespace).Delete(context.Background(), hashed, metav1.DeleteOptions{})
	if err != nil {
		log.Fatal(err)
	}
	return
}

// func applyGlotpodJob(cs *kubernetes.Clientset, code string, language string, input string, namespace string) error {
// 	sha := sha1.New()
// 	sha.Write([]byte(code + language + input))
// 	hashed := sha.Sum(nil)
// 	fmt.Println("Hashed name is ", hashed)
// 	job := &batchv1.Job{
// 		TypeMeta: metav1.TypeMeta{
// 			APIVersion: "goglot.dev.github.com/v1alpha1",
// 			Kind:       "Glotpod",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      string(hashed),
// 			Namespace: namespace,
// 		},
// 		Spec: ,
// 	}
// 	job, err := cs.BatchV1().Jobs(namespace).Create(context.Background(), job, metav1.CreateOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
