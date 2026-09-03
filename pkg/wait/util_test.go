package wait

import (
	"context"
	"errors"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/fake"

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
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, utilwait.ErrWaitTimeout) {
		t.Fatalf("WaitForAggregatorReady() error = %v, want deadline or wait timeout", err)
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
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, utilwait.ErrWaitTimeout) {
		t.Fatalf("WaitForAggregatorReady() error = %v, want deadline or wait timeout", err)
	}
}
