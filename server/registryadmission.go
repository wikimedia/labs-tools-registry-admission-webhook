package server

import (
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RegistryAdmission type is where the registry is defined and the handler method is linked
type RegistryAdmission struct {
	Registry string
}

// HandleAdmission is the logic of the whole webhook, really.  This is where
// the decision to allow a Kubernetes pod update or create or not takes place.
func (r *RegistryAdmission) HandleAdmission(review *admissionv1.AdmissionReview) error {
	// logrus.Debugln(review.Request)
	req := review.Request
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		logrus.Errorf("Could not unmarshal raw object: %v", err)
		review.Response = &admissionv1.AdmissionResponse{
			UID: req.UID,
			Result: &v1.Status{
				Message: err.Error(),
			},
		}
		return nil
	}
	logrus.Debugf("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, pod.Name, req.UID, req.Operation, req.UserInfo)
	for i := 0; i < len(pod.Spec.Containers); i++ {
		container := &pod.Spec.Containers[i]
		if !strings.HasPrefix(container.Image, r.Registry+"/") &&
			req.Namespace != "kube-system" {
			logrus.Errorf(
				"Attempt to use docker image not in approved registry: %v, namespace: %v",
				container.Image,
				req.Namespace,
			)
			review.Response = &admissionv1.AdmissionResponse{
				UID: req.UID,
				Allowed: false,
				Result: &v1.Status{
					Message: "Only WMCS-approved docker registry allowed",
				},
			}
			return nil
		}
		logrus.Debugf("Found registry image: %v", container.Image)
	}

	review.Response = &admissionv1.AdmissionResponse{
		UID: req.UID,
		Allowed: true,
		Result: &v1.Status{
			Message: "Welcome to the fantasy zone!",
		},
	}
	return nil
}
