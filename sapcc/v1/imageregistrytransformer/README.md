Image registry transformer
--------------------------

The image registry transformer ensures all images are used from the specified registry.

While it would also be possible to use the image transformer in each base and for every image that comes builtin with Kustomize,
this plugin provides a global mechanism to set registries.

# Example

Given the configuration 

```yaml
$ cat mirror-transformer.yaml 

apiVersion: sapcc/v1
kind: ImageRegistryTransformer
metadata:
  name: mirror

imageRegistries:
  - name: dockerhub
    newName: mirror.cloud.sap/dockerhub-mirror
  - name: quay.io
    newName: mirror.cloud.sap/quay-mirror
  - name: k8s.gcr.io
    newName: mirror.cloud.sap/k8sgcr-mirror
```

and

``` yaml
$ cat kustomization.yaml

transformers:
  - mirror-transformer.yaml

resources:
  - ...
```

any image registry in any resource will be replaced, e.g. 
```
k8s.gcr.io/autoscaling/vpa-updater:0.10.0 -> mirror.cloud.sap/ccloud-k8sgcr-mirror/autoscaling/vpa-updater:0.10.0
```

