#### Qiniu Certificate Sync ####

For whatever reason you want to transfer TLS certificates in kubernetes to Qiniu CDN

This app will upload provided TLS secrets to Qiniu, and update the certificates of provided domains

WARNING: This program currently won't verify the certificates in any way, this may or may not change in the future but use it at your own risk for now.

## USAGE
-----

### locally

Change your configuration like what config.toml suggested

```
go run . --kubeconfig somewhere \ # default in ~/.kube/config
        --config somewhere-else # default in /etc/qiniu-cert-sync/config.toml
```

### in cluster

Fill in the configs under the configmap section in k8s.yaml, and change the namespaces of all sections

```
kubectl apply -f ./k8s.yaml
```

