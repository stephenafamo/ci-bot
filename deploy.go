package main

import (
	"net/url"

	"github.com/spf13/viper"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// deploy() takes a Build, deploys and return the URL
// We have to generate an ID and the appropriate URL first
func deploy(build Build) (URL string, err error) {

	u, err := url.Parse(build.Project.URL)
	if err != nil {
		return
	}
	u.Host = build.Target + "." + build.Type + "." + u.Host
	URL = u.String()

	Id := build.Project.ID + "-" + build.Type + "-" + build.Target
	labels := map[string]string{
		"project":     build.Project.ID,
		"target":      build.Target,
		"type":        build.Type,
		"environment": "qa",
	}

	err = deployToUrl(build.Image, Id, URL, labels)
	return
}

// deployToProd() is to depoly a build to production
// Unlike deploy(), it does not add any sepcial identifiers
// to the url or ID.
func deployToProd(build Build) (URL string, err error) {

	u, err := url.Parse(build.Project.URL)
	if err != nil {
		return
	}
	URL = u.String()

	Id := build.Project.ID
	labels := map[string]string{
		"project":     build.Project.ID,
		"environment": "production",
	}

	err = deployToUrl(build.Image, Id, URL, labels)
	return
}

// deployToUrl() is the generic deploy function.
// Improvements to be made:
//     Allow flixibility in defining ports, resources and replicas
func deployToUrl(Image, Id, URL string,
	labels map[string]string) (err error) {

	config, err := clientcmd.BuildConfigFromFlags(
		"",
		viper.GetString("KubeConfigPath"),
	)

	if err != nil {
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}

	svcClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: Id,
			Annotations: map[string]string{
				"getambassador.io/config": ` |
				      ---
				      apiVersion: ambassador/v0
				      kind:  Mapping
				      name:  ` + Id + `
				      host: ` + URL + `
				      service: ` + Id,
			},
		},
		Spec: apiv1.ServiceSpec{
			Selector: labels,
			Ports: []apiv1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
		},
	}

	depClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: Id,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  Id,
							Image: Image,
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	_, getErr := svcClient.Get(Id, metav1.GetOptions{})
	if getErr != nil {
		switch getErr.(type) {
		case *errors.StatusError:
			if getErr.(*errors.StatusError).Status().Code == 404 {
				_, err = svcClient.Create(service)
				if err != nil {
					return
				}
			}

		default:
			err = getErr
			return
		}
	}

	_, getErr = depClient.Get(Id, metav1.GetOptions{})
	if getErr != nil {
		switch getErr.(type) {
		case *errors.StatusError:
			if getErr.(*errors.StatusError).Status().Code == 404 {
				_, err = depClient.Create(deployment)
				if err != nil {
					return
				}
			}

		default:
			err = getErr
			return
		}
	} else {
		_, updateErr := depClient.Update(deployment)
		if updateErr != nil {
			err = updateErr
			return
		}
	}

	return
}

func int32Ptr(i int32) *int32 { return &i }
