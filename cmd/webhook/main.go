package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	admissionv1 "k8s.io/api/admission/v1"
)

func mutate(w http.ResponseWriter, r *http.Request) {
	var review admissionv1.AdmissionReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, fmt.Sprint("could not decode body:", err), http.StatusBadRequest)
		return
	}

	response := admissionv1.AdmissionReview{
		TypeMeta: review.TypeMeta,
		Response: &admissionv1.AdmissionResponse{
			Allowed: true,
			UID:     review.Request.UID,
		},
	}

	if review.Request.SubResource == "status" {
		log.Printf("mutate(%#v)", review.Request)

		pt := admissionv1.PatchTypeJSONPatch
		response.Response.PatchType = &pt
		response.Response.Patch = []byte(`[{
			"op": "replace",
			"path": "/status/allocatable",
			"value": {
				"cpu": "10m"
			}
		}]`)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprint("could not encode response:", err), http.StatusInternalServerError)
		return
	}
}

func validate(w http.ResponseWriter, r *http.Request) {
	var review admissionv1.AdmissionReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, fmt.Sprint("could not decode body:", err), http.StatusBadRequest)
		return
	}

	response := admissionv1.AdmissionReview{
		TypeMeta: review.TypeMeta,
		Response: &admissionv1.AdmissionResponse{
			Allowed: false,
			UID:     review.Request.UID,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprint("could not encode response:", err), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/validate", validate)
	http.HandleFunc("/mutate", mutate)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
