# Kustomize plugins

Collection of external plugins to extend [Kustomize](https://kustomize.io). Read more about [Kustomize Plugins](https://github.com/kubernetes-sigs/kustomize/tree/master/docs/plugins).

## Installation

Create a folder for external plugins, copy the entire `sapcc` directory and populate it via `KUSTOMIZE_PLUGIN_HOME=$pathToFolder` environment variable.  

```bash
# Ensure home for Kustomize plugins exists.
export KUSTOMIZE_PLUGIN_HOME=~/.kustomize/plugin && mkdir -p $KUSTOMIZE_PLUGIN_HOME

# Get the plugins.
wget -qO- https://github.com/sapcc/helm-outdated-dependencies/releases/download/$VERSION/kustomize-plugins_$VERSION_$OSTYPE_amd64.tar.gz | tar xvz - -C $KUSTOMIZE_PLUGIN_HOME
```

As of 03/2020 Kustomize external plugins are an alpha feature, so build needs to be invoked with the `--enable_alpha_plugins` flag.

## List of plugins

### ValueTransformer

Replaces variables used in resources incl. Kubernetes Secrets.  

Example:

The ValueTransformer 
```
apiVersion: sapcc/v2
kind: ValueTransformer
metadata:
  name: ignored
argsFromFile: cluster.yaml
```
with the `cluster.yaml` file containing multiple replacements
```
REGION: qa-de-1
CLUSTER: s-qa-de-1
```

will replace every occurence of `$REGION` and `$CLUSTER` in the given resources.


### SecretsMerger
