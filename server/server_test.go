package server

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gerrit.wikimedia.org/labs/tools/registry-admission-webhook/config"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	AdmissionRequestFailPod = admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind: "AdmissionReview",
		},
		Request: &admissionv1.AdmissionRequest{
			UID: "e911857d-c318-11e8-bbad-025000000001",
			Kind: metav1.GroupVersionKind{
				Kind: "Pod",
			},
			Operation: "CREATE",
			Object: runtime.RawExtension{
				Raw: []byte(`{
					"metadata": {
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
					}
				}`),
			},
		},
	}
	AdmissionRequestPassPod = admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind: "AdmissionReview",
		},
		Request: &admissionv1.AdmissionRequest{
			UID: "e911857d-c318-11e8-bbad-025000000001",
			Kind: metav1.GroupVersionKind{
				Kind: "Pod",
			},
			Operation: "CREATE",
			Object: runtime.RawExtension{
				Raw: []byte(`{
					"metadata": {
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
					}
				}`),
			},
		},
	}
	AdmissionRequestFailDeployment = admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind: "AdmissionReview",
		},
		Request: &admissionv1.AdmissionRequest{
			UID: "e911857d-c318-11e8-bbad-025000000001",
			Kind: metav1.GroupVersionKind{
				Group:   "core",
				Version: "v1",
				Kind:    "Deployment",
			},
			Operation: "CREATE",
			Object: runtime.RawExtension{
				Raw: []byte(`{
					"apiVersion": "apps/v1",
					"kind": "Deployment",
					"metadata": {
						"annotations": {
							"deployment.kubernetes.io/revision": "1"
						},
						"creationTimestamp": "2021-10-21T15:56:13Z",
						"generation": 1,
						"labels": {
							"app.kubernetes.io/component": "web",
							"app.kubernetes.io/managed-by": "webservice",
							"name": "openstack-browser",
							"toolforge": "tool"
						},
						"name": "openstack-browser",
						"namespace": "tool-openstack-browser",
						"resourceVersion": "509633552",
						"uid": "50e8a190-f762-40f7-bef9-e3ec3384e3b2"
					},
					"spec": {
						"progressDeadlineSeconds": 600,
						"replicas": 1,
						"revisionHistoryLimit": 10,
						"selector": {
							"matchLabels": {
								"app.kubernetes.io/component": "web",
								"app.kubernetes.io/managed-by": "webservice",
								"name": "openstack-browser",
								"toolforge": "tool"
							}
						},
						"strategy": {
							"rollingUpdate": {
								"maxSurge": "25%",
								"maxUnavailable": "25%"
							},
							"type": "RollingUpdate"
						},
						"template": {
							"metadata": {
								"creationTimestamp": null,
								"labels": {
									"app.kubernetes.io/component": "web",
									"app.kubernetes.io/managed-by": "webservice",
									"name": "openstack-browser",
									"toolforge": "tool"
								}
							},
							"spec": {
								"containers": [
									{
										"command": [
											"/usr/bin/webservice-runner",
											"--type",
											"uwsgi-python",
											"--port",
											"8000"
										],
										"image": "hub.docker.io/example:foobar",
										"imagePullPolicy": "Always",
										"name": "webservice",
										"ports": [
											{
												"containerPort": 8000,
												"name": "http",
												"protocol": "TCP"
											}
										],
										"resources": {},
										"terminationMessagePath": "/dev/termination-log",
										"terminationMessagePolicy": "File",
										"workingDir": "/data/project/openstack-browser/"
									}
								],
								"dnsPolicy": "ClusterFirst",
								"restartPolicy": "Always",
								"schedulerName": "default-scheduler",
								"securityContext": {},
								"terminationGracePeriodSeconds": 30
							}
						}
					}
				}`),
			},
		},
	}
	AdmissionRequestPassDeployment = admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind: "AdmissionReview",
		},
		Request: &admissionv1.AdmissionRequest{
			UID: "e911857d-c318-11e8-bbad-025000000001",
			Kind: metav1.GroupVersionKind{
				Group:   "core",
				Version: "v1",
				Kind:    "Deployment",
			},
			Operation: "CREATE",
			Object: runtime.RawExtension{
				Raw: []byte(`{
					"apiVersion": "apps/v1",
					"kind": "Deployment",
					"metadata": {
						"annotations": {
							"deployment.kubernetes.io/revision": "1"
						},
						"creationTimestamp": "2021-10-21T15:56:13Z",
						"generation": 1,
						"labels": {
							"app.kubernetes.io/component": "web",
							"app.kubernetes.io/managed-by": "webservice",
							"name": "openstack-browser",
							"toolforge": "tool"
						},
						"name": "openstack-browser",
						"namespace": "tool-openstack-browser",
						"resourceVersion": "509633552",
						"uid": "50e8a190-f762-40f7-bef9-e3ec3384e3b2"
					},
					"spec": {
						"progressDeadlineSeconds": 600,
						"replicas": 1,
						"revisionHistoryLimit": 10,
						"selector": {
							"matchLabels": {
								"app.kubernetes.io/component": "web",
								"app.kubernetes.io/managed-by": "webservice",
								"name": "openstack-browser",
								"toolforge": "tool"
							}
						},
						"strategy": {
							"rollingUpdate": {
								"maxSurge": "25%",
								"maxUnavailable": "25%"
							},
							"type": "RollingUpdate"
						},
						"template": {
							"metadata": {
								"creationTimestamp": null,
								"labels": {
									"app.kubernetes.io/component": "web",
									"app.kubernetes.io/managed-by": "webservice",
									"name": "openstack-browser",
									"toolforge": "tool"
								}
							},
							"spec": {
								"containers": [
									{
										"command": [
											"/usr/bin/webservice-runner",
											"--type",
											"uwsgi-python",
											"--port",
											"8000"
										],
										"image": "docker-registry.tools.wmflabs.org/toolforge-python39-sssd-web:latest",
										"imagePullPolicy": "Always",
										"name": "webservice",
										"ports": [
											{
												"containerPort": 8000,
												"name": "http",
												"protocol": "TCP"
											}
										],
										"resources": {},
										"terminationMessagePath": "/dev/termination-log",
										"terminationMessagePolicy": "File",
										"workingDir": "/data/project/openstack-browser/"
									}
								],
								"dnsPolicy": "ClusterFirst",
								"restartPolicy": "Always",
								"schedulerName": "default-scheduler",
								"securityContext": {},
								"terminationGracePeriodSeconds": 30
							}
						}
					}
				}`),
			},
		},
	}
)

func decodeResponse(body io.ReadCloser) *admissionv1.AdmissionReview {
	response, _ := ioutil.ReadAll(body)
	review := &admissionv1.AdmissionReview{}
	codecs.UniversalDeserializer().Decode(response, nil, review)
	return review
}

func encodeRequest(review *admissionv1.AdmissionReview) []byte {
	ret, err := json.Marshal(review)
	if err != nil {
		logrus.Errorln(err)
	}
	return ret
}

func TestServeReturnsCorrectJson(t *testing.T) {
	nsc := &RegistryAdmission{}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(&AdmissionRequestPassPod))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)
	t.Log(review.Response)
	if review.Request.UID != AdmissionRequestPassPod.Request.UID {
		t.Error("Request and response UID don't match")
	}
}

func TestServeReturnsCorrectJsonWithDefaultConfig(t *testing.T) {
	config, _ := config.GetConfigFromEnv()
	nsc := &RegistryAdmission{Registries: config.Registries}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(&AdmissionRequestPassPod))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)
	t.Log(review.Response)
	if review.Request.UID != AdmissionRequestPassPod.Request.UID {
		t.Error("Request and response UID don't match")
	}
}
func TestHookFailsOnBadRegistryPod(t *testing.T) {
	nsc := &RegistryAdmission{Registries: []string{"docker-registry.tools.wmflabs.org"}}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(&AdmissionRequestFailPod))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)
	t.Log(review.Response)
	if review.Response.Allowed {
		t.Error("Allowed pod that should not have been allowed!")
	}
}
func TestHookPassesOnRightRegistryPod(t *testing.T) {
	nsc := &RegistryAdmission{Registries: []string{"docker-registry.tools.wmflabs.org"}}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(&AdmissionRequestPassPod))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)
	t.Log(review.Response)
	if !review.Response.Allowed {
		t.Error("Failed to allow pod that should have been allowed!")
	}
}
func TestHookFailsOnBadRegistryDeployment(t *testing.T) {
	nsc := &RegistryAdmission{Registries: []string{"docker-registry.tools.wmflabs.org"}}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(&AdmissionRequestFailDeployment))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)
	t.Log(review.Response)
	if review.Response.Allowed {
		t.Error("Allowed deployment that should not have been allowed!")
	}
}

func TestHookFailsOnBadRegistryDeploymentWhenMoreThanOne(t *testing.T) {
	nsc := &RegistryAdmission{Registries: []string{"dummyregistry1", "docker-registry.tools.wmflabs.org"}}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(&AdmissionRequestFailDeployment))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)
	t.Log(review.Response)
	if review.Response.Allowed {
		t.Error("Allowed deployment that should not have been allowed!")
	}
}
func TestHookPassesOnRightRegistryDeployment(t *testing.T) {
	nsc := &RegistryAdmission{Registries: []string{"docker-registry.tools.wmflabs.org"}}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(&AdmissionRequestPassDeployment))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)
	t.Log(review.Response)
	if !review.Response.Allowed {
		t.Error("Failed to allow deployment that should have been allowed!")
	}
}

func TestHookPassesOnRightRegistryDeploymentWhenNotFirst(t *testing.T) {
	nsc := &RegistryAdmission{Registries: []string{"dummyregistry1", "docker-registry.tools.wmflabs.org"}}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(&AdmissionRequestPassDeployment))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)
	t.Log(review.Response)
	if !review.Response.Allowed {
		t.Error("Failed to allow deployment that should have been allowed!")
	}
}
