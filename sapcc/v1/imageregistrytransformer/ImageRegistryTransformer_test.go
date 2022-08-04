// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestAnnotationsTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("ImageRegistryTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: sapcc/v1
kind: ImageRegistryTransformer
metadata:
  name: notImportantHere

imageRegistry: gcr.io/test
`, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:latest
        name: nginx
`, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: gcr.io/test/nginx:latest
        name: nginx
`)
}
