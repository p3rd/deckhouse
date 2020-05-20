package deckhouse

import (
	"encoding/base64"
	"github.com/flant/logboek"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	deckhouseRegistrySecretName = "deckhouse-registry"
	deckhouseRegistryVolumeName = "registrysecret"
)

//nolint:funlen
func generateDeckhouseDeployment(registry, logLevel, bundle string, isSecureRegistry bool) *appsv1.Deployment {
	var deckhouseDeployment = `
kind: Deployment
apiVersion: apps/v1
metadata:
  name: deckhouse
  namespace: d8-system
  labels:
    heritage: deckhouse
spec:
  replicas: 1
  selector:
    matchLabels:
      app: deckhouse
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: deckhouse
    spec:
      containers:
      - name: deckhouse
        image: PLACEHOLDER
        command:
        - /deckhouse/deckhouse
        imagePullPolicy: Always
        env:
        - name: LOG_LEVEL
          value: PLACEHOLDER
        - name: DECKHOUSE_BUNDLE
          value: PLACEHOLDER
        - name: DECKHOUSE_POD
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: HELM_HOST
          value: "127.0.0.1:44434"
        - name: ADDON_OPERATOR_CONFIG_MAP
          value: deckhouse
        - name: ADDON_OPERATOR_PROMETHEUS_METRICS_PREFIX
          value: deckhouse_
        - name: ADDON_OPERATOR_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: ADDON_OPERATOR_LISTEN_ADDRESS
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: KUBERNETES_DEPLOYED
          value: PLACEHOLDER
        ports:
        - containerPort: 9650
        readinessProbe:
          httpGet:
            path: /ready
            port: 9650
          initialDelaySeconds: 5
          # fail after 10 minutes
          periodSeconds: 5
          failureThreshold: 120
        resources:
          requests:
            cpu: 50m
            memory: 512Mi
        workingDir: /deckhouse
      hostNetwork: true
      dnsPolicy: Default
      serviceAccountName: deckhouse
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
      - operator: Exists
`

	var deployment appsv1.Deployment
	_ = yaml.Unmarshal([]byte(deckhouseDeployment), &deployment)

	deployment.Spec.Template.Spec.Containers[0].Image = registry

	for i, env := range deployment.Spec.Template.Spec.Containers[0].Env {
		switch env.Name {
		case "LOG_LEVEL":
			deployment.Spec.Template.Spec.Containers[0].Env[i].Value = logLevel
		case "DECKHOUSE_BUNDLE":
			deployment.Spec.Template.Spec.Containers[0].Env[i].Value = bundle
		case "KUBERNETES_DEPLOYED":
			deployment.Spec.Template.Spec.Containers[0].Env[i].Value = time.Unix(0, time.Now().Unix()).String()
		}
	}

	if isSecureRegistry {
		deployment.Spec.Template.Spec.ImagePullSecrets = []apiv1.LocalObjectReference{
			{Name: deckhouseRegistrySecretName},
		}

		deployment.Spec.Template.Spec.Volumes = []apiv1.Volume{
			{
				Name: deckhouseRegistryVolumeName,
				VolumeSource: apiv1.VolumeSource{
					Secret: &apiv1.SecretVolumeSource{SecretName: deckhouseRegistrySecretName},
				},
			},
		}

		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = []apiv1.VolumeMount{
			{
				Name:      deckhouseRegistryVolumeName,
				MountPath: "/etc/registrysecret",
				ReadOnly:  true,
			},
		}
	}

	return &deployment
}

func generateDeckhouseNamespace(name string) *apiv1.Namespace {
	return &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"heritage": "deckhouse",
			},
			Annotations: map[string]string{
				"extended-monitoring.flant.com/enabled": "",
			},
		},
		Spec: apiv1.NamespaceSpec{
			Finalizers: []apiv1.FinalizerName{
				apiv1.FinalizerKubernetes,
			},
		},
	}
}

func generateDeckhouseServiceAccount() *apiv1.ServiceAccount {
	return &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deckhouse",
			Labels: map[string]string{
				"heritage": "deckhouse",
			},
		},
	}
}

func generateDeckhouseAdminClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-admin",
			Labels: map[string]string{
				"heritage": "deckhouse",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{rbacv1.APIGroupAll},
				Resources: []string{rbacv1.ResourceAll},
				Verbs:     []string{rbacv1.VerbAll},
			},
			{
				NonResourceURLs: []string{rbacv1.NonResourceAll},
				Verbs:           []string{rbacv1.VerbAll},
			},
		},
	}
}

func generateDeckhouseAdminClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deckhouse",
			Labels: map[string]string{
				"heritage": "deckhouse",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      "deckhouse",
				Namespace: "d8-system",
			},
		},
	}
}

func generateDeckhouseRegistrySecret(dockerCfg string) *apiv1.Secret {
	data, _ := base64.StdEncoding.DecodeString(dockerCfg)
	return &apiv1.Secret{
		Type: apiv1.SecretTypeDockercfg,
		ObjectMeta: metav1.ObjectMeta{
			Name: deckhouseRegistrySecretName,
			Labels: map[string]string{
				"heritage": "deckhouse",
			},
		},
		Data: map[string][]byte{
			apiv1.DockerConfigKey: data,
		},
	}
}

func generateDeckhouseConfigMap(deckhouseConfig map[string]interface{}) *apiv1.ConfigMap {
	configMap := apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deckhouse",
			Labels: map[string]string{
				"heritage": "deckhouse",
			},
		},
	}

	configMapData := make(map[string]string, len(deckhouseConfig))
	for setting, data := range deckhouseConfig {
		if strings.HasSuffix(setting, "Enabled") {
			boolData, ok := data.(bool)
			if !ok {
				logboek.LogWarnF("deckhouse config map: %q must be boo\n", setting)
			}
			configMapData[setting] = strconv.FormatBool(boolData)

			continue
		}
		convertedData, err := yaml.Marshal(data)
		if err != nil {
			logboek.LogWarnF("preparing deckhouse config map error (probably validation bug): %v", err)
			continue
		}
		configMapData[setting] = string(convertedData)
	}
	configMap.Data = configMapData
	return &configMap
}

func generateSecret(name, namespace string, data map[string][]byte) *apiv1.Secret {
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"heritage": "deckhouse",
			},
		},
		Data: data,
	}
}

func generateSecretWithTerraformState(data []byte) *apiv1.Secret {
	return generateSecret(
		"d8-cluster-terraform-state",
		"kube-system",
		map[string][]byte{
			"cluster_terraform_state.json": data,
		},
	)
}

func generateSecretWithClusterConfig(data []byte) *apiv1.Secret {
	return generateSecret("d8-cluster-configuration", "kube-system",
		map[string][]byte{"cluster-configuration.yaml": data})
}

func generateSecretWithProviderClusterConfig(configData, discoveryData []byte) *apiv1.Secret {
	return generateSecret("d8-provider-cluster-configuration", "kube-system",
		map[string][]byte{
			"cloud-provider-cluster-configuration.yaml": configData,
			"cloud-provider-discovery-data.json":        discoveryData,
		})
}

func int32Ptr(i int32) *int32 { return &i }