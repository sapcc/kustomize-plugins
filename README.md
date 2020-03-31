# Kustomize plugins

Collection of external plugins to extend [Kustomize](https://kustomize.io). Read more about [Kustomize Plugins](https://github.com/kubernetes-sigs/kustomize/tree/master/docs/plugins).

## Installation

Create a folder for external plugins, copy the entire `sapcc` directory and populate it via `KUSTOMIZE_PLUGIN_HOME=$pathToFolder` environment variable.  

```bash
export KUSTOMIZE_PLUGIN_HOME=~/.kustomize/plugin && mkdir -p $KUSTOMIZE_PLUGIN_HOME && cp -r sapcc $KUSTOMIZE_PLUGIN_HOME
```

As of 03/2020 Kustomize external plugins are an alpha feature, so build needs to be invoked with the `--enable_alpha_plugins` flag.

## List of plugins

### ValueTransformer

Replaces variables used in resources. [Credits](https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/sedtransformer).

Example:
```
apiVersion: sapcc/v1
kind: ValueTransformer
metadata:
  name: ignored
argsOneLiner: s/$REGION/qa-de-1/g
#argsFromFile: cluster.globals
```

with an optional file containing multiple replacements:
```
s/$REGION/qa-de-1/g
s/$CLUSTER/s-qa-de-1/g
```
