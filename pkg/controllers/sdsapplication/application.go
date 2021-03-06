package sdsapplication

import (
	"github.com/samsung-cnct/cma-operator/pkg/util/cma"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	api "github.com/samsung-cnct/cma-operator/pkg/apis/cma/v1alpha1"
	"github.com/samsung-cnct/cma-operator/pkg/util/helmutil"
	"github.com/samsung-cnct/cma-operator/pkg/util/k8sutil"
	"k8s.io/client-go/kubernetes"
)

func (c *SDSApplicationController) deployApplication(application *api.SDSApplication) (bool, error) {
	config, err := c.getRestConfigForRemoteCluster(application.Spec.PackageManager.Name, application.Namespace, nil)
	if err != nil {
		return false, err
	}
	packageManager, err := cma.GetSDSPackageManager(application.Spec.PackageManager.Name, "default", nil)

	k8sutil.CreateJob(helmutil.GenerateHelmInstallJob(application.Spec), packageManager.Spec.Namespace, config)

	application.Status.Phase = api.ApplicationPhaseInstalling
	_, err = c.client.CmaV1alpha1().SDSApplications(application.Namespace).Update(application)
	if err == nil {
		logger.Infof("Deployed helm install job for -->%s<--", application.Spec.Name)
	} else {
		logger.Infof("Could not update the status error was %s", err)
	}

	return true, nil
}

func (c *SDSApplicationController) waitForApplication(application *api.SDSApplication) (result bool, err error) {
	config, err := c.getRestConfigForRemoteCluster(application.Spec.PackageManager.Name, application.Namespace, nil)
	if err != nil {
		return false, err
	}

	packageManager, err := cma.GetSDSPackageManager(application.Spec.PackageManager.Name, "default", nil)
	if err != nil {
		logger.Infof("Cannot retrieve package manager for application %s", application.Spec.Name)
		return false, err
	}

	clientset, _ := kubernetes.NewForConfig(config)
	timeout := 0
	for timeout < 2000 {
		job, err := clientset.BatchV1().Jobs(packageManager.Spec.Namespace).Get("app-install-"+application.Spec.Name, v1.GetOptions{})
		if err == nil {
			if job.Status.Succeeded > 0 {
				application.Status.Phase = api.ApplicationPhaseImplemented
				application.Status.Ready = true
				_, err = c.client.CmaV1alpha1().SDSApplications(application.Namespace).Update(application)
				if err == nil {
					logger.Infof("Helm installed app -->%s<--", application.Spec.Name)
					c.updateSDSCluster(application.Spec.PackageManager.Name)
				} else {
					logger.Infof("Could not update the status error was %s", err)
				}
				return true, nil
			}
		}
		time.Sleep(5 * time.Second)
		timeout++
	}
	return false, nil
}

func (c *SDSApplicationController) updateSDSCluster(clusterName string) (result bool, err error) {
	// TODO This is dubious, but for the PoC, good enough
	sdsCluster, err := cma.GetSDSCluster(clusterName, "default", nil)
	if err != nil {
		logger.Infof("Failed to get SDSCluster for SDSApplication %s, error was: ", clusterName, err)
		return false, err
	}

	changes := false
	if sdsCluster.Status.AppsInstalled == false {
		changes = true
		sdsCluster.Status.AppsInstalled = true
	}
	switch sdsCluster.Status.Phase {
	case api.ClusterPhaseHavePackageManager, api.ClusterPhaseDeployingApplications:
		changes = true
		sdsCluster.Status.Phase = api.ClusterPhaseReady
	}

	if changes {
		_, err = cma.UpdateSDSCluster(sdsCluster, sdsCluster.Namespace, nil)
		if err != nil {
			logger.Infof("Could not update SDSCluster for KrakenCluster %s, error was: ", sdsCluster.Name, err)
		}
	}

	return true, nil
}
