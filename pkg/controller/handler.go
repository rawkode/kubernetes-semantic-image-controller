package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

var defaultResolver = &dockerHubResolver{}

type patchObject struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

// HandleMutate takes in our /mutate request and figures out any patches that
// that need to be made to resolve the image field for semantic versioning
func HandleMutate(w http.ResponseWriter, r *http.Request) {
	input := &admissionv1.AdmissionReview{}
	err := json.NewDecoder(r.Body).Decode(input)
	if err != nil {
		sendErr(w, fmt.Errorf("could not unmarshal review: %v", err))
		return
	}

	pod := &corev1.Pod{}
	err = json.Unmarshal(input.Request.Object.Raw, pod)
	if err != nil {
		sendErr(w, fmt.Errorf("could not unmarshal pod: %v", err))
		return
	}

	// Go through each of the containers and build up a list of patches. For
	// things where we can't resolve the image or it's already an absolute
	// path, we will mostly skip over
	patchMap := make([]patchObject, 0, len(pod.Spec.Containers))
	for idx, container := range pod.Spec.Containers {
		if !defaultResolver.ShouldResolve(container.Image) {
			continue
		}

		resolvedImage, err := defaultResolver.Resolve(container.Image)
		if err != nil {
			sendErr(w, fmt.Errorf("could not resolve image %s: %v", container.Image, err))
			return
		}

		log.Printf("- Resolved %s to %s", container.Image, resolvedImage)
		patchMap = append(patchMap, patchObject{
			Op:    "replace",
			Path:  fmt.Sprintf("/spec/containers/%d/image", idx),
			Value: resolvedImage,
		})
	}

	patchBytes, err := json.Marshal(patchMap)
	if err != nil {
		sendErr(w, fmt.Errorf("could not generate patch: %v", err))
		return
	}

	jsonPatch := admissionv1.PatchTypeJSONPatch
	respReview := &admissionv1.AdmissionReview{
		TypeMeta: input.TypeMeta,
		Response: &admissionv1.AdmissionResponse{
			UID:       input.Request.UID,
			Allowed:   true,
			Patch:     patchBytes,
			PatchType: &jsonPatch,
		},
	}

	respBytes, err := json.Marshal(respReview)
	if err != nil {
		sendErr(w, fmt.Errorf("could not generate response: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}

func sendErr(w http.ResponseWriter, err error) {
	out, err := json.Marshal(map[string]string{"Err": err.Error()})
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(out)
}
