# Registry Admission Validation Webhook for Toolforge

This is a simple [Kubernetes Admission Validation Webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#what-are-admission-webhooks) that is intended to replace
a custom validating admission controller previously compiled into older versions of
Kubernetes.  Since the webhook will compile with the versions it uses, it will work
as long as the admission v1 API is valid.

## Use and development

This is pending adaption to Toolforge.  Currently it depends on local docker images and it
can be built and deployed on Kubernetes by insuring any node it is expected to run on
has access to the image it uses.  The image will need to be in a registry most likely when deployed.

It was developed using [Go Modules](https://github.com/golang/go/wiki/Modules), which will
validate the hash of every imported library during build.  At this time, it depends on
these external go libraries:

	* github.com/kelseyhightower/envconfig
	* github.com/sirupsen/logrus
	* k8s.io/api
	* k8s.io/apimachinery

To build on minikube (current supported k8s version is 1.21) and launch, just run:
* `deploy.sh -b dev`

## Testing

At the top level, run `go test ./...` to capture all tests.  If you need to see output
or want to examine things more, use `go test -test.v ./...`

## Deploying

You can deploy it using the `deploy.sh` script.
For the dev environment rebuilding the container (minikube running locally):
* `deploy.sh -ib dev`

If you don't want to rebuild the image, skip the flag:
* `deploy.sh -i dev`

The `-i` flag will ask before changing anything (skip it for non-interactive deployment).

Since this was designed for use in [Toolforge](https://wikitech.wikimedia.org/wiki/Portal:Toolforge "Toolforge Portal"), so the instructions here focus on that.

The version of docker on the builder host is very old, so the builder/scratch pattern in
the Dockerfile won't work.

* Build the container image on the docker-builder host (currently tools-docker-imagebuilder-01.tools.eqiad1.wikimedia.cloud). `$ docker build . -t docker-registry.tools.wmflabs.org/registry-admission:latest`
* Push the image to the internal repo: `root@tools-docker-imagebuilder-01:~# docker push docker-registry.tools.wmflabs.org/registry-admission:latest`
* On a control plane node as root (or as a cluster-admin user), with a checkout of the repo there somewhere (in a home directory is probably great), as root or admin user on Kubernetes, run `root@tools-k8s-control-1:# ./deploy.sh toolsbeta`


## Updating the certs

Certificates created with the Kubernetes API are valid for one year. When upgrading Kubernetes (or whenever necessary)
it is wise to rotate the certs for this service. To do so, you can do it at deploy time with (change `tools` with the env you are refreshing the certs for):

* `./deploy.sh -ic tools`

Or any time by simply running (as cluster admin or root@control host) `root@tools-k8s-control-1:# ./deploy/utils/get-cert.sh`. That will recreate the cert secret. Then delete the existing pods to ensure
that the golang web services are serving the new cert.
