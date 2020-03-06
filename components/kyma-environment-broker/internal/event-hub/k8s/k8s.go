package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SecretFrom(name, namespace, brokersEndpoint, password, eventHubNamespace string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"knativekafka.kyma-project.io/kafka-secret": "true",
				"installer":                    "overrides",
				"component":                    "knative-eventing-channel-kafka",
				"kyma-project.io/installation": "",
			},
		},
		StringData: map[string]string{
			"kafka.brokers":             brokersEndpoint,
			"kafka.namespace":           eventHubNamespace,
			"kafka.password":            password,
			"kafka.username":            "${K8S_SECRET_USERNAME}",
			"kafka.secretName":          "knative-kafka",
			"environment.kafkaProvider": "azure",
		},
		Type: "Opaque",
	}
}
