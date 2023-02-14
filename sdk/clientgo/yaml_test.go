package clientgo

import "testing"

func Test_InstallYaml(t *testing.T) {
	const deploymentYAML = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: default
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
`
	err := InstallYaml([]byte(deploymentYAML))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}

func Test_UnInstallYaml(t *testing.T) {
	const deploymentYAML = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: default
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
`
	err := UnInstallYaml([]byte(deploymentYAML))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
