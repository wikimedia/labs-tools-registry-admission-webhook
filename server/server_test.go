package server

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	AdmissionRequestFail = v1beta1.AdmissionReview{
		TypeMeta: v1.TypeMeta{
			Kind: "AdmissionReview",
		},
		Request: &v1beta1.AdmissionRequest{
			UID: "e911857d-c318-11e8-bbad-025000000001",
			Kind: v1.GroupVersionKind{
				Kind: "Pod",
			},
			Operation: "CREATE",
			Object: runtime.RawExtension{
				Raw: []byte(`{"metadata": {
        						"name": "test",
        						"uid": "e911857d-c318-11e8-bbad-025000000001",
						        "creationTimestamp": "2018-09-28T12:20:39Z"
						      },

						"kind": "Pod",
						"apiVersion": "v1",
						"metadata": {
						"name": "busybox-sleep",
						"namespace": "default",
						"uid": "4b54be10-8d3c-11e9-8b7a-080027f5f85c",
						"creationTimestamp": "2019-06-12T18:02:51Z"
						},
						"spec": {
						"volumes": [
							{
							"name": "default-token-99n2k",
							"secret": {
								"secretName": "default-token-99n2k"
							}
							}
						],
						"containers": [
							{
							"name": "busybox",
							"image": "busybox",
							"args": [
								"sleep",
								"1000000"
							],
							"resources": {},
							"volumeMounts": [
								{
								"name": "default-token-99n2k",
								"readOnly": true,
								"mountPath": "/var/run/secrets/kubernetes.io/serviceaccount"
								}
							],
							"terminationMessagePath": "/dev/termination-log",
							"terminationMessagePolicy": "File",
							"imagePullPolicy": "Always"
							}
						],
						"restartPolicy": "Always",
						"terminationGracePeriodSeconds": 30,
						"dnsPolicy": "ClusterFirst",
						"serviceAccountName": "default",
						"serviceAccount": "default",
						"securityContext": {},
						"schedulerName": "default-scheduler",
						"tolerations": [
							{
							"key": "node.kubernetes.io/not-ready",
							"operator": "Exists",
							"effect": "NoExecute",
							"tolerationSeconds": 300
							},
							{
							"key": "node.kubernetes.io/unreachable",
							"operator": "Exists",
							"effect": "NoExecute",
							"tolerationSeconds": 300
							}
						],
						"priority": 0,
						"enableServiceLinks": true
						},
						"status": {
						"phase": "Pending",
						"qosClass": "BestEffort"
						}}`),
			},
		},
	}
	AdmissionRequestPass = v1beta1.AdmissionReview{
		TypeMeta: v1.TypeMeta{
			Kind: "AdmissionReview",
		},
		Request: &v1beta1.AdmissionRequest{
			UID: "e911857d-c318-11e8-bbad-025000000001",
			Kind: v1.GroupVersionKind{
				Kind: "Pod",
			},
			Operation: "CREATE",
			Object: runtime.RawExtension{
				Raw: []byte(`{"metadata": {
        						"name": "test",
        						"uid": "e911857d-c318-11e8-bbad-025000000001",
						        "creationTimestamp": "2018-09-28T12:20:39Z"
						      },

						"kind": "Pod",
						"apiVersion": "v1",
						"metadata": {
						"name": "busybox-sleep",
						"namespace": "default",
						"uid": "4b54be10-8d3c-11e9-8b7a-080027f5f85c",
						"creationTimestamp": "2019-06-12T18:02:51Z"
						},
						"spec": {
						"volumes": [
							{
							"name": "default-token-99n2k",
							"secret": {
								"secretName": "default-token-99n2k"
							}
							}
						],
						"containers": [
							{
							"name": "busybox-working",
							"image": "docker-registry.tools.wmflabs.org/busybox:1.2.3",
							"args": [
								"sleep",
								"1000000"
							],
							"resources": {},
							"volumeMounts": [
								{
								"name": "default-token-99n2k",
								"readOnly": true,
								"mountPath": "/var/run/secrets/kubernetes.io/serviceaccount"
								}
							],
							"terminationMessagePath": "/dev/termination-log",
							"terminationMessagePolicy": "File",
							"imagePullPolicy": "Always"
							}
						],
						"restartPolicy": "Always",
						"terminationGracePeriodSeconds": 30,
						"dnsPolicy": "ClusterFirst",
						"serviceAccountName": "default",
						"serviceAccount": "default",
						"securityContext": {},
						"schedulerName": "default-scheduler",
						"tolerations": [
							{
							"key": "node.kubernetes.io/not-ready",
							"operator": "Exists",
							"effect": "NoExecute",
							"tolerationSeconds": 300
							},
							{
							"key": "node.kubernetes.io/unreachable",
							"operator": "Exists",
							"effect": "NoExecute",
							"tolerationSeconds": 300
							}
						],
						"priority": 0,
						"enableServiceLinks": true
						},
						"status": {
						"phase": "Pending",
						"qosClass": "BestEffort"
						}}`),
			},
		},
	}
)

func decodeResponse(body io.ReadCloser) *v1beta1.AdmissionReview {
	response, _ := ioutil.ReadAll(body)
	review := &v1beta1.AdmissionReview{}
	codecs.UniversalDeserializer().Decode(response, nil, review)
	return review
}

func encodeRequest(review *v1beta1.AdmissionReview) []byte {
	ret, err := json.Marshal(review)
	if err != nil {
		logrus.Errorln(err)
	}
	return ret
}

func TestServeReturnsCorrectJson(t *testing.T) {
	nsc := &RegistryAdmission{}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(&AdmissionRequestPass))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)
	t.Log(review.Response)
	if review.Request.UID != AdmissionRequestPass.Request.UID {
		t.Error("Request and response UID don't match")
	}
}
func TestHookFailsOnBadRegistry(t *testing.T) {
	nsc := &RegistryAdmission{}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(&AdmissionRequestFail))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)
	t.Log(review.Response)
	if review.Response.Allowed {
		t.Error("Allowed pod that should not have been allowed!")
	}
}
func TestHookPassesOnRightRegistry(t *testing.T) {
	nsc := &RegistryAdmission{}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(&AdmissionRequestPass))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)
	t.Log(review.Response)
	if review.Response.Allowed {
		t.Error("Allowed pod that should not have been allowed!")
	}
}
