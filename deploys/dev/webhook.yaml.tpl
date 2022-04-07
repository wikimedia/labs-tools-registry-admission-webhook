apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: registry-admission
webhooks:
  - name: registry-admission.tools.wmflabs.org
    clientConfig:
      service:
        name: registry-admission
        namespace: registry-admission
        path: "/"
      caBundle: @@CA_BUNDLE@@
    rules:
      - operations: ["CREATE","UPDATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
      - operations: ["CREATE","UPDATE"]
        apiGroups: ["apps"]
        apiVersions: ["v1"]
        resources: ["deployments"]
    failurePolicy: Ignore
    admissionReviewVersions: ["v1"]
    sideEffects: None
