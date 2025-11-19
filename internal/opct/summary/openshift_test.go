package summary

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExtractOpenShiftConfigMap(t *testing.T) {
	t.Run("should_extract_invoker_from_valid_configmap", func(t *testing.T) {
		configMapList := &v1.ConfigMapList{
			Items: []v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "openshift-install-manifests",
					},
					Data: map[string]string{
						"invoker": "user",
					},
				},
			},
		}

		osSummary := NewOpenShiftSummary()
		err := osSummary.ExtractOpenShiftConfigMap(configMapList)

		assert.NoError(t, err)
		assert.Equal(t, "user", osSummary.InstallInvoker)
	})

	t.Run("should_not_set_invoker_if_configmap_name_does_not_match", func(t *testing.T) {
		configMapList := &v1.ConfigMapList{
			Items: []v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "other-configmap",
					},
					Data: map[string]string{
						"invoker": "test-invoker",
					},
				},
			},
		}

		osSummary := NewOpenShiftSummary()
		err := osSummary.ExtractOpenShiftConfigMap(configMapList)

		assert.NoError(t, err)
		assert.Empty(t, osSummary.InstallInvoker)
	})

	t.Run("should_handle_empty_configmap_list", func(t *testing.T) {
		configMapList := &v1.ConfigMapList{}

		osSummary := NewOpenShiftSummary()
		err := osSummary.ExtractOpenShiftConfigMap(configMapList)

		assert.NoError(t, err)
		assert.Empty(t, osSummary.InstallInvoker)
	})

	t.Run("should_handle_nil_configmap_list", func(t *testing.T) {
		osSummary := NewOpenShiftSummary()
		err := osSummary.ExtractOpenShiftConfigMap(nil)

		assert.NoError(t, err)
		assert.Empty(t, osSummary.InstallInvoker)
	})

	t.Run("should_handle_configmap_without_invoker_key", func(t *testing.T) {
		configMapList := &v1.ConfigMapList{
			Items: []v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "openshift-install-manifests",
					},
					Data: map[string]string{
						"some-other-key": "some-value",
					},
				},
			},
		}

		osSummary := NewOpenShiftSummary()
		err := osSummary.ExtractOpenShiftConfigMap(configMapList)

		assert.NoError(t, err)
		assert.Empty(t, osSummary.InstallInvoker)
	})
}
