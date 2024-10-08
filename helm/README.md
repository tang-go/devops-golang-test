# mytest-replicaset

## Requirements

- Kubernetes v1.13+.
- helm v3.


## Create namespace

```cosole
apiVersion: v1
kind: Namespace
metadata:
  labels:
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/name: devops-test
    control-plane: controller-manager
  name: devops-test-system
```
```console
kubectl create -f helm/templates/namespace.yaml
```





## Installing
```console
helm install mytest-replicaset
```

To install the chart with the release name `mytest-replicaset`:

```console
helm install mytest-replicaset ./helm -n ${namespace}
```

The command deploys the mytest-replicaset chart to the namespace ${namespace} of the Kubernetes cluster with the default configuration. The configuration section lists the parameters that can be configured during installation.

## Uninstalling

To uninstall/delete the `my-release` deployment:

```console
helm delete my-release
```