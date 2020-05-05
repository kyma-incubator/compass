package gardener

import (
	"io/ioutil"
	"os"
	"testing"
)

var (
	goodKubeconfig = []byte(`apiVersion: v1
kind: Config
clusters:
  - name: test
    cluster:
      server: 'http://127.0.0.1:8080'
users:
  - name: test
    user:
      token: NEUzOTNCNUItRThFQS00QzdBLTlEMUYtQzM5ODRDQUQ4MUE4LTdERTNDNjc5LTI2NDktNDg3NS05MzQ1LUZCNTIzNEY5RkE3Qi0xOTk2N0Y3OC0zQTg4LTRBOUUtODY0Mi0zNUMwQkVEQjBDNTY=
contexts:
  - name: test
    context:
      cluster: test
      user: test
      namespace: test
current-context: test`)

	badKubeconfig = []byte(`apiVersion: v1
kind: Config
clusters:
  - name: test
users:
  - name: test
contexts:
  - name: test`)
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name       string
		kubeconfig *[]byte
		wantErr    bool
	}{
		{
			name:       "good kubeconfig",
			kubeconfig: &goodKubeconfig,
			wantErr:    false,
		},
		{
			name:       "bad kubeconfig",
			kubeconfig: &badKubeconfig,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		tt := tt // pin!

		t.Run(tt.name, func(t *testing.T) {
			kubec, err := ioutil.TempFile("", "config")
			if err != nil {
				t.Fatal(err)
			}

			defer func() {
				err = os.Remove(kubec.Name())
				if err != nil {
					t.Errorf("error removing kubeconfig file %v", err)
					return
				}
			}()

			if _, err = kubec.Write(*tt.kubeconfig); err != nil {
				t.Errorf("error writing kubeconfig file %v", err)
				return
			}

			_, err = NewClient(kubec.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
