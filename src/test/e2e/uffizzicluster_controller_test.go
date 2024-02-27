package e2e

import (
	"context"
	"github.com/UffizziCloud/uffizzi-cluster-operator/src/api/v1alpha1"
	"github.com/UffizziCloud/uffizzi-cluster-operator/src/test/util/resources"
	. "github.com/onsi/ginkgo/v2"
)

type TestDefinition struct {
	Name string
	Spec v1alpha1.UffizziClusterSpec
}

func (td *TestDefinition) ExecLifecycleTest(ctx context.Context) {
	ns := resources.CreateTestNamespace(td.Name)
	uc := resources.CreateTestUffizziCluster(td.Name, ns.Name)
	wrapUffizziClusterLifecycleTest(ctx, ns, uc)
}

const (
	timeout        = "1m"
	pollingTimeout = "100ms"
)

var _ = Describe("Basic UffizziCluster Lifecycle", func() {
	ctx := context.Background()
	testUffizziCluster := TestDefinition{
		Name: "basic",
		Spec: v1alpha1.UffizziClusterSpec{},
	}
	// run the testUffizziCluster
	testUffizziCluster.ExecLifecycleTest(ctx)
})
