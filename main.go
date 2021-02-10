package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Masterminds/semver"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

func handleMutate(w http.ResponseWriter, r *http.Request) {
	log.Printf("Mutate called!")

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

	versions := []string{
		"1.15",
		"1.15.6",
		"1.17.4",
		"1.18.0",
		"1.19.6",
	}

	patchMap := []map[string]string{}
	for idx, container := range pod.Spec.Containers {

		// nginx: >= 1.78
		// nginx: '>= 1.2 <= 1.4.5
		split := strings.SplitN(container.Image, ":", 2)
		imageName := split[0]
		versionConstraint := split[1]
		constraint, err := semver.NewConstraint(versionConstraint)

		if err != nil {
			sendErr(w, fmt.Errorf("could not parse constraint: %v", err))
			return
		}

		var bestVersion string
		for _, version := range versions {
			v, err := semver.NewVersion(version)
			if err != nil {
				sendErr(w, fmt.Errorf("could not parse version: %v", err))
				return
			}

			if constraint.Check(v) {
				bestVersion = version
			}
		}

		patchMap = append(patchMap, map[string]string{
			"op":    "replace",
			"path":  fmt.Sprintf("/spec/containers/%d/image", idx),
			"value": fmt.Sprintf("%s:%s", imageName, bestVersion),
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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", handleMutate)
	srv := &http.Server{Addr: ":443", Handler: mux}
	log.Fatal(srv.ListenAndServeTLS("/certs/webhook.crt", "/certs/webhook-key.pem"))
}
