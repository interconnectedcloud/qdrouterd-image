package router

import (
	"fmt"
	"os"

	"github.com/interconnectedcloud/qdr-image/test/k8s/utils"
	"github.com/interconnectedcloud/qdr-image/test/k8s/utils/constants"
	"github.com/interconnectedcloud/qdr-image/test/k8s/utils/k8s"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const QpidDispatchImageEnvVar = "QPID_DISPATCH_IMAGE"

type QpidDispatchDeploymentOpts struct {
	k8s.DeploymentOpts
	ConfigMap *QpidDispatchConfigMap
}

func NewDeployment(namespace string, qpidDispatch QpidDispatch, opts QpidDispatchDeploymentOpts) (*apps.Deployment, error) {
	var d *apps.Deployment
	var err error

	// Validating mandatory fields
	if utils.StrEmpty(namespace) {
		err := fmt.Errorf("namespace is required")
		return nil, err
	}
	if utils.StrEmpty(qpidDispatch.Id) {
		err := fmt.Errorf("QpidDispatch.Id is required")
		return nil, err
	}

	// Default values - use constant by default or image from opts or from env if any provided
	image := utils.StrDefault(constants.QpidDispatchImage, opts.Image, os.Getenv(QpidDispatchImageEnvVar))

	// Static definitions for ActiveMQ Artemis Deployment
	replicas := int32(1)
	terminationSecs := int64(60)
	pullPolicy := core.PullAlways
	restartPolicy := core.RestartPolicyAlways

	// Preparing the Deployment
	envVars := []core.EnvVar{
		{Name: "QDROUTERD_CONF", Value: "/opt/router/qdrouterd.conf"},
	}
	if len(opts.EnvVars) > 0 {
		envVars = append(envVars, opts.EnvVars...)
	}
	d = &apps.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      qpidDispatch.Id,
			Namespace: namespace,
			Labels:    opts.Labels,
		},
		Spec: apps.DeploymentSpec{
			Replicas: &replicas,
			Selector: &meta.LabelSelector{
				MatchLabels: opts.Labels,
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: opts.Labels,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{Name: qpidDispatch.Id, Image: image, ImagePullPolicy: pullPolicy,
							Env: envVars,
							VolumeMounts: []core.VolumeMount{
								{Name: "router-config", MountPath: "/opt/router", ReadOnly: true},
							}},
					},
					Volumes: []core.Volume{
						{Name: "router-config", VolumeSource: core.VolumeSource{
							ConfigMap: &core.ConfigMapVolumeSource{
								LocalObjectReference: core.LocalObjectReference{
									Name: opts.ConfigMap.Name,
								},
							},
						}},
					},
					RestartPolicy:                 restartPolicy,
					TerminationGracePeriodSeconds: &terminationSecs,
				},
			},
		},
	}

	return d, err
}
