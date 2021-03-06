---
title: Weave Flux FAQ
menu_order: 60
---

## General questions

Also see [the introduction](/site/introduction.md).

### What does Flux do?

Flux automates the process of deploying new configuration and
container images to Kubernetes.

### How does it automate deployment?

It synchronises all manifests in a repository with a Kubernetes cluster.
It also monitors container registries for new images and updates the
manifests accordingly.

### How is that different from a bash script?

The amount of functionality contained within Flux warrants a dedicated
application/service. An equivalent script could easily get too large
to maintain and reuse.

Anyway, we've already done it for you!

### Why should I automate deployment?

Automation is a principle of lean development. It reduces waste, to
provide efficiency gains. It empowers employees by removing dull
tasks. It mitigates against failure by avoiding silly mistakes.

### I thought Flux was about service routing?

That's [where we started a while
ago](https://www.weave.works/blog/flux-service-routing/). But we
discovered that automating deployments was more urgent for our own
purposes.

Staging deployments with clever routing is useful, but it's a later
level of operational maturity.

There are some pretty good solutions for service routing:
[Envoy](https://www.envoyproxy.io/), [Istio](https://istio.io) for
example. We may return to the matter of staged deployments.

## Technical questions

### Why does Flux need a deploy key?

Flux needs a deploy key to be allowed to push to the version control
system in order to read and update the manifests.

### How do I give Flux access to a private registry?

Flux transparently looks at the image pull secret that you give for a
controller, and thereby uses the same credentials that Kubernetes uses
for pulling each image. If your pods are running, Kubernetes has
pulled the images, and Flux should be able to access them.

There are exceptions: in some environments, authorisation provided by
the platform is used instead of image pull secrets. Google Container
Registry works this way, for example (and we have introduced a special
case for it so Flux will work there too).

### How often does Flux check for new images?

Short answer: every five minutes.

You can set this to be more often, to try and be responsive in
deploying new images. Be aware that there are some other things going
on:

 * Requests to image registries are rate limited, and _discovering_
   new images has some lag in itself, which depends on how many images
   you are using in total (since we go and check them one by one as
   fast as rate limiting allows).

 * Operations on the git repo are more or less
   serialised, so while you are checking for new images to deploy, you
   are not doing something else (like syncing).

Having said that, the defaults are pretty conservative, so try it and
see. Please don't increase the rate limiting numbers (`--registry-rps`
and `--registry-burst`) -- it's possible to get blacklisted by image
registries if you spam them with requests.

### How do I use my own deploy key?

Flux uses a k8s secret to hold the git ssh deploy key. It is possible to
provide your own.

First delete the secret (if it exists):

`kubectl delete secret flux-git-deploy`

Then create a new secret named `flux-git-deploy`, using your key as the content of the secret:

`kubectl create secret generic flux-git-deploy --from-file=identity=/full/path/to/key`

Now restart fluxd to re-read the k8s secret (if it is running):

`kubectl delete $(kubectl get pod -o name -l name=flux)`

### How do I use a private docker registry?

Create a Kubernetes Secret with your docker credentials then add the
name of this secret to your Pod manifest under the `imagePullSecrets`
setting. Flux will read this value and parse the Kubernetes secret.

For a guide showing how to do this, see the
[Kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).
