package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
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
	req := review.Request

	logrus.Debugf("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)

	if req.Kind.Kind == "Pod" {
		var pod corev1.Pod
		if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
			logrus.Errorf("Could not unmarshal raw object: %v", err)
			review.Response = &admissionv1.AdmissionResponse{
				UID: req.UID,
				Allowed: false,
				Result: &v1.Status{
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
				UID: req.UID,
				Allowed: false,
				Result: &v1.Status{
					Message: err.Error(),
				},
			}
			return nil
		}

		return r.handleForPodSpec(deployment.Spec.Template.Spec, review)
	}

	logrus.Warningf("Encountered unsupported Kind=%v", req.Kind)
	review.Response = &admissionv1.AdmissionResponse{
		UID: req.UID,
		Allowed: false,
		Result: &v1.Status{
			Message: fmt.Sprintf("Unsupported Kind=%v", req.Kind),
		},
	}

	return nil
}

func (r *RegistryAdmission) handleForPodSpec(podSpec corev1.PodSpec, review *admissionv1.AdmissionReview) error {
	for i := 0; i < len(podSpec.Containers); i++ {
		container := &podSpec.Containers[i]
		if !strings.HasPrefix(container.Image, r.Registry+"/") && review.Request.Namespace != "kube-system" {
			logrus.Errorf(
				"Attempt to use docker image not in approved registry: %v, namespace: %v, object: %v/%v",
				container.Image,
				review.Request.Namespace,
				review.Request.Kind.Kind,
				review.Request.Name,
			)
			review.Response = &admissionv1.AdmissionResponse{
				UID: review.Request.UID,
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
		UID: review.Request.UID,
		Allowed: true,
		Result: &v1.Status{
			Message: "Welcome to the fantasy zone!",
		},
	}
	return nil
}
