package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RegistryAdmission type is where the registry is defined and the handler method is linked
type RegistryAdmission struct {
	Registries []string
}

// HandleAdmission is the logic of the whole webhook, really.  This is where
// the decision to allow a Kubernetes pod update or create or not takes place.
func (r *RegistryAdmission) HandleAdmission(review *admissionv1.AdmissionReview) error {
	req := review.Request

	logrus.Debugf("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)

	if req.Kind.Kind == "Pod" {
		var pod corev1.Pod
		if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
			logrus.Errorf("Could not unmarshal raw object: %v", err)
			review.Response = &admissionv1.AdmissionResponse{
				UID:     req.UID,
				Allowed: false,
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
			return nil
		}

		return r.handleForPodSpec(pod.Spec, review)
	} else if req.Kind.Kind == "Deployment" {
		var deployment appsv1.Deployment
		if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
			logrus.Errorf("Could not unmarshal raw object: %v", err)
			review.Response = &admissionv1.AdmissionResponse{
				UID:     req.UID,
				Allowed: false,
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
			return nil
		}

		return r.handleForPodSpec(deployment.Spec.Template.Spec, review)
	}

	logrus.Warningf("Encountered unsupported Kind=%v", req.Kind)
	review.Response = &admissionv1.AdmissionResponse{
		UID:     req.UID,
		Allowed: false,
		Result: &metav1.Status{
			Message: fmt.Sprintf("Unsupported Kind=%v", req.Kind),
		},
	}

	return nil
}

func (r *RegistryAdmission) handleForPodSpec(podSpec corev1.PodSpec, review *admissionv1.AdmissionReview) error {
	allPassed := true
	var nonCompliantImages []string
	for i := 0; i < len(podSpec.Containers); i++ {
		container := &podSpec.Containers[i]
		matchesAnyRegistry := false
		logrus.Debugf("Checking container image: %v", container.Image)
		for _, registry := range r.Registries {
			logrus.Debugf("Checking container image %v against registry %v", container.Image, registry)
			if strings.HasPrefix(container.Image, registry+"/") || review.Request.Namespace == "kube-system" {
				logrus.Debugf("Image %v matches registry %v", container.Image, registry)
				matchesAnyRegistry = true
			}
		}
		if !matchesAnyRegistry {
			logrus.Debugf("Image %v not in one of the known registries %v", container.Image, r.Registries)
			nonCompliantImages = append(
				nonCompliantImages,
				fmt.Sprintf(
					"Kind=%v, Namespace=%v Name=%v Image=%v",
					review.Request.Kind,
					review.Request.Namespace,
					review.Request.Name,
					container.Image,
				),
			)
			allPassed = false
		}
	}

	if allPassed {
		review.Response = &admissionv1.AdmissionResponse{
			UID:     review.Request.UID,
			Allowed: true,
			Result: &metav1.Status{
				Message: "Welcome to the fantasy zone!",
			},
		}
		return nil
	}

	errorMessage := fmt.Sprintf("The following container images did not match any of the allowed registries (%v): %v",
		r.Registries,
		nonCompliantImages,
	)
	logrus.Errorf(errorMessage)
	review.Response = &admissionv1.AdmissionResponse{
		UID:     review.Request.UID,
		Allowed: false,
		Result: &metav1.Status{
			Message: errorMessage,
		},
	}
	return nil
}
