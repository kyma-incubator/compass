package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SecretFrom(name, namespace, brokersEndpoint, username, password, eventHubNamespace string) *corev1.Secret {
	return &corev1.Secret{
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
			"kafka.username":            username,
			"kafka.password":            password,
			"kafka.namespace":           eventHubNamespace,
			"kafka.secretName":          "knative-kafka",
			"environment.kafkaProvider": "azure",
		},
		Type: "Opaque",
	}
}
