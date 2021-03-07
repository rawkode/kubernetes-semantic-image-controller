package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestHandleMutate(t *testing.T) {
	// Override our resolver with a mock one that doesn't call the internet
	oldResolver := defaultResolver
	defaultResolver = &staticResolver{images: map[string][]string{
		"nginx": {"1.15.0", "1.15.1", "1.15.2", "1.15.3", "1.16.0", "1.16.1"},
	}}
	defer func() {
		defaultResolver = oldResolver
	}()

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleMutate)

	// Run our request and expect all to be okay!
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "nginx: >1.15.0, <1.16"},
				{Image: "nginx: <1.17"},
			},
		},
	}
	podBytes, err := json.Marshal(&pod)
	assert.NoError(t, err)

	review := &admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{Raw: podBytes},
		},
	}
	reviewBytes, err := json.Marshal(review)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/mutate", bytes.NewReader(reviewBytes))
	assert.NoError(t, err)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)

	respReview := &admissionv1.AdmissionReview{}
	assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), respReview))

	patches := make([]patchObject, 0)
	assert.NoError(t, json.Unmarshal(respReview.Response.Patch, &patches))
	assert.Equal(t, []patchObject{
		{Op: "replace", Path: "/spec/containers/0/image", Value: "nginx:1.15.3"},
		{Op: "replace", Path: "/spec/containers/1/image", Value: "nginx:1.16.1"},
	}, patches)
}

func TestHandleMutateBadRequest(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleMutate)

	req, err := http.NewRequest("POST", "/mutate", bytes.NewReader([]byte{}))
	assert.NoError(t, err)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "could not unmarshal review")
}

func TestHandleHealth(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleHealth)

	// Run our request and expect all to be okay!
	req, err := http.NewRequest("GET", "/health", nil)
	assert.NoError(t, err)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, `{"status": "ok"}`, recorder.Body.String())
}
