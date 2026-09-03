package wait

import (
	"context"
	"errors"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	"github.com/redhat-openshift-ecosystem/opct/pkg"
)

func readyAggregatorPod(name string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: pkg.CertificationNamespace,
			Labels: map[string]string{
				"component":           "sonobuoy",
				"sonobuoy-component":  "aggregator",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			Conditions: []v1.PodCondition{
				{Type: v1.PodReady, Status: v1.ConditionTrue},
			},
		},
	}
}

func pendingAggregatorPod(name string) *v1.Pod {
	pod := readyAggregatorPod(name)
	pod.Status.Phase = v1.PodPending
	pod.Status.Conditions = nil
	return pod
}

func TestWaitForAggregatorReady_ready(t *testing.T) {
	kclient := fake.NewSimpleClientset(readyAggregatorPod("sonobuoy-aggregator"))

	if err := WaitForAggregatorReady(context.Background(), kclient); err != nil {
		t.Fatalf("WaitForAggregatorReady() error = %v, want nil", err)
	}
}

func TestWaitForAggregatorReady_noPodsTimesOut(t *testing.T) {
	kclient := fake.NewSimpleClientset()

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	err := WaitForAggregatorReady(ctx, kclient)
	if err == nil {
		t.Fatal("WaitForAggregatorReady() error = nil, want timeout error")
	}
	if !utilwait.Interrupted(err) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitForAggregatorReady() error = %v, want interrupted or deadline exceeded", err)
	}
}

func TestWaitForAggregatorReady_pendingTimesOut(t *testing.T) {
	kclient := fake.NewSimpleClientset(pendingAggregatorPod("sonobuoy-aggregator"))

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	err := WaitForAggregatorReady(ctx, kclient)
	if err == nil {
		t.Fatal("WaitForAggregatorReady() error = nil, want timeout error")
	}
	if !utilwait.Interrupted(err) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitForAggregatorReady() error = %v, want interrupted or deadline exceeded", err)
	}
}

func TestWaitForAggregatorReady_forbidden(t *testing.T) {
	kclient := fake.NewSimpleClientset()
	kclient.PrependReactor("list", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "pods"}, "", errors.New("denied"))
	})

	err := WaitForAggregatorReady(context.Background(), kclient)
	if err == nil {
		t.Fatal("WaitForAggregatorReady() error = nil, want forbidden error")
	}
	if !apierrors.IsForbidden(err) {
		t.Fatalf("WaitForAggregatorReady() error = %v, want forbidden error", err)
	}
}
