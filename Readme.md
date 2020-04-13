# Pachyderm Operator
The goal of this project

**Status**: Experimental

# Running the project locally

Set your kube context to local kube cluster

Install the manifests (including CRD) on your cluster
```
make install
```

Run the controller
```
make run ENABLE_WEBHOOKS=false
```

Install an instance of the CRD and watch the controller logs

```
kubectl apply -f config/samples/ops_v1_pachrelease.yaml
```
