package wait

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/redhat-openshift-ecosystem/opct/pkg"
)

const (
	aggregatorLabelSelector = "component=sonobuoy,sonobuoy-component=aggregator"
	aggregatorReadyTimeout  = 3 * time.Minute
	retryLogInterval        = 30 * time.Second
)

// WaitForRequiredResources will wait for the sonobuoy pod in the sonobuoy namespace to go into
// a Running/Ready state and then return nil.
func WaitForRequiredResources(kclient kubernetes.Interface) error {
	var obj kruntime.Object

	restClient := kclient.CoreV1().RESTClient()

	lw := cache.NewFilteredListWatchFromClient(restClient, "pods", pkg.CertificationNamespace, func(options *metav1.ListOptions) {
		options.LabelSelector = aggregatorLabelSelector
	})

	// Wait for Sonobuoy aggregator pod to become Ready
	ctx, cancel := context.WithTimeout(context.TODO(), time.Minute*10)
	defer cancel()
	_, err := watchtools.UntilWithSync(ctx, lw, obj, nil, func(event watch.Event) (bool, error) {
		switch event.Type {
		case watch.Error:
			return false, fmt.Errorf("error waiting for sonobuoy to start: %w", event.Object.(error))
		case watch.Deleted:
			return false, errors.New("sonobuoy pod deleted while waiting to become ready")
		}

		pod, isPod := event.Object.(*v1.Pod)
		if !isPod {
			return false, errors.New("type error watching for sononbuoy to start")
		}

		if pod.Status.Phase == v1.PodRunning && podIsReady(pod) {
			return true, nil
		}

		// Loop again
		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// WaitForAggregatorReady polls until the sonobuoy aggregator pod is Running/Ready.
// Uses List (no watch/reflector) with exponential backoff for transient API errors.
func WaitForAggregatorReady(parentCtx context.Context, kclient kubernetes.Interface) error {
	ctx, cancel := context.WithTimeout(parentCtx, aggregatorReadyTimeout)
	defer cancel()

	backoff := wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2.0,
		Cap:      30 * time.Second,
		Steps:    30,
		Jitter:   0.1,
	}

	startedAt := time.Now()
	var lastWarnAt time.Time
	warnIfDue := func(msg string, err error) {
		now := time.Now()
		if now.Sub(startedAt) < retryLogInterval {
			return
		}
		if !lastWarnAt.IsZero() && now.Sub(lastWarnAt) < retryLogInterval {
			return
		}
		lastWarnAt = now
		if err != nil {
			log.WithError(err).Warn(msg)
			return
		}
		log.Warn(msg)
	}

	err := wait.ExponentialBackoffWithContext(ctx, backoff, func(ctx context.Context) (bool, error) {
		pods, err := kclient.CoreV1().Pods(pkg.CertificationNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: aggregatorLabelSelector,
		})
		if err != nil {
			if apierrors.IsForbidden(err) || apierrors.IsUnauthorized(err) || apierrors.IsBadRequest(err) {
				return false, fmt.Errorf("listing sonobuoy aggregator pods: %w", err)
			}
			warnIfDue("error listing sonobuoy aggregator pods, retrying", err)
			return false, nil
		}

		if len(pods.Items) == 0 {
			warnIfDue("sonobuoy aggregator pod not found yet, retrying", nil)
			return false, nil
		}

		for i := range pods.Items {
			pod := &pods.Items[i]
			if pod.Status.Phase == v1.PodRunning && podIsReady(pod) {
				return true, nil
			}
		}

		warnIfDue("sonobuoy aggregator pod is not ready yet, retrying", nil)
		return false, nil
	})
	if err == nil {
		return nil
	}

	if wait.Interrupted(err) {
		return fmt.Errorf("timeout waiting for sonobuoy aggregator pod to become ready after %s: %w", aggregatorReadyTimeout, err)
	}
	return err
}

func podIsReady(pod *v1.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == v1.PodReady && cond.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}
