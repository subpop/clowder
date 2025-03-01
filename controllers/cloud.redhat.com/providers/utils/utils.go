package providers

import (
	"github.com/RedHatInsights/clowder/apis/cloud.redhat.com/v1alpha1/common"
	"github.com/RedHatInsights/clowder/controllers/cloud.redhat.com/config"
	obj "github.com/RedHatInsights/clowder/controllers/cloud.redhat.com/object"
	"github.com/RedHatInsights/clowder/controllers/cloud.redhat.com/providers"
	"github.com/RedHatInsights/clowder/controllers/cloud.redhat.com/providers/sizing"
	"github.com/RedHatInsights/clowder/controllers/cloud.redhat.com/utils"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// MakeLocalDB populates the given deployment object with the local DB struct.
func MakeLocalDB(dd *apps.Deployment, nn types.NamespacedName, baseResource obj.ClowdObject, extraLabels *map[string]string, cfg *config.DatabaseConfig, image string, usePVC bool, dbName string, res *core.ResourceRequirements) {
	labels := baseResource.GetLabels()
	labels["service"] = "db"

	for k, v := range *extraLabels {
		labels[k] = v
	}

	labler := utils.MakeLabeler(nn, labels, baseResource)
	labler(dd)

	var volSource core.VolumeSource
	if usePVC {
		volSource = core.VolumeSource{
			PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
				ClaimName: nn.Name,
			},
		}
		dd.Spec.Strategy.Type = apps.RecreateDeploymentStrategyType
		dd.Spec.Strategy.RollingUpdate = nil
	} else {
		volSource = core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		}
		dd.Spec.Strategy.Type = apps.RollingUpdateDeploymentStrategyType
		dd.Spec.Strategy.RollingUpdate = &apps.RollingUpdateDeployment{
			MaxUnavailable: &intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "25%",
			},
			MaxSurge: &intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "25%",
			},
		}
	}

	dd.Spec.Replicas = common.Int32Ptr(1)
	dd.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	dd.Spec.Template.Spec.Volumes = []core.Volume{
		{
			Name:         nn.Name,
			VolumeSource: volSource,
		},
	}

	dd.Spec.Template.ObjectMeta.Labels = labels

	envVars := []core.EnvVar{
		{Name: "POSTGRESQL_USER", Value: cfg.Username},
		{Name: "POSTGRESQL_PASSWORD", Value: cfg.Password},
		{Name: "PGPASSWORD", Value: cfg.AdminPassword}, // Legacy for old db images can likely be removed soon
		{Name: "POSTGRESQL_MASTER_USER", Value: cfg.AdminUsername},
		{Name: "POSTGRESQL_MASTER_PASSWORD", Value: cfg.AdminPassword},
		{Name: "POSTGRESQL_DATABASE", Value: dbName},
	}
	ports := []core.ContainerPort{{
		Name:          "database",
		ContainerPort: 5432,
		Protocol:      core.ProtocolTCP,
	}}

	probeHandler := core.ProbeHandler{
		Exec: &core.ExecAction{
			Command: []string{
				"psql",
				"-U",
				"$(POSTGRESQL_USER)",
				"-d",
				"$(POSTGRESQL_DATABASE)",
				"-c",
				"SELECT 1",
			},
		},
	}

	livenessProbe := core.Probe{
		ProbeHandler:        probeHandler,
		InitialDelaySeconds: 15,
		TimeoutSeconds:      2,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	readinessProbe := core.Probe{
		ProbeHandler:        probeHandler,
		InitialDelaySeconds: 45,
		TimeoutSeconds:      2,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}

	requestResource := sizing.GetDefaultResourceRequirements()

	if res != nil {
		requestResource = *res
	}

	c := core.Container{
		Name:           nn.Name,
		Image:          image,
		Env:            envVars,
		LivenessProbe:  &livenessProbe,
		ReadinessProbe: &readinessProbe,
		Ports:          ports,
		Resources:      requestResource,
		VolumeMounts: []core.VolumeMount{{
			Name:      nn.Name,
			MountPath: "/var/lib/pgsql/data",
		}},
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: core.TerminationMessageReadFile,
		ImagePullPolicy:          core.PullIfNotPresent,
	}

	dd.Spec.Template.Spec.Containers = []core.Container{c}
}

// MakeLocalDBService populates the given service object with the local DB struct.
func MakeLocalDBService(s *core.Service, nn types.NamespacedName, baseResource obj.ClowdObject, extraLabels *map[string]string) {
	servicePorts := []core.ServicePort{{
		Name:       "database",
		Port:       5432,
		Protocol:   "TCP",
		TargetPort: intstr.FromInt(5432),
	}}
	labels := providers.Labels{"service": "db", "app": baseResource.GetClowdName()}
	for k, v := range *extraLabels {
		labels[k] = v
	}
	utils.MakeService(s, nn, labels, servicePorts, baseResource, false)
}

// MakeLocalDBPVC populates the given PVC object with the local DB struct.
func MakeLocalDBPVC(pvc *core.PersistentVolumeClaim, nn types.NamespacedName, baseResource obj.ClowdObject, capacity string) {
	utils.MakePVC(pvc, nn, providers.Labels{"service": "db", "app": baseResource.GetClowdName()}, capacity, baseResource)
}
