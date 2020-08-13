# tekton-step-observer  
Observer plugin to emit [CloudEvents](https://cloudevents.io/) at step execution level for Tekton  

## Caution
```
The purpose of this controller is to explore/share  
more use cases to consume CloudEvents at step execution level.
The author does not assume any production/development environment nor pushing to the upstream.
```

## Installation
`tekton-step-observer` requires TektonCD Pipeline on your Kubernetes cluster.
Plus, you need a build tool [ko](https://github.com/google/ko).

For installation, [after TektonCD Pipeline installed](https://tekton.dev/docs/getting-started/), run:
```shell script
ko apply -f config
```
then `ko` starts build and deploy the resources to the cluster.

For uninstallation, run:
```shell script
kubectl delete -f config
```
then all resources gets removed from the cluster.
Note that this is a "plugin", hence no effect on TektonCD Pipelines.


## Concept  

This plugin creates a controller to watch TaskRun, 
then send CloudEvents to the read [`default-cloud-event-sink`](https://github.com/tektoncd/pipeline/blob/50ed02b4c2b96656355548acea878a0d20e89750/config/config-defaults.yaml#L59) 
when it detects a change of state of each step.  

Supported CloudEvents event type are:
```
dev.tekton.event.plugin.step.started.v1  
dev.tekton.event.plugin.step.failed.v1
dev.tekton.event.plugin.step.succeeded.v1
dev.tekton.event.plugin.step.skipped.v1
```

When it detects the state for any emission, it gathers its corresponding defined information such as 
[v1beta1.Step](https://github.com/tektoncd/pipeline/blob/50ed02b4c2b96656355548acea878a0d20e89750/pkg/apis/pipeline/v1beta1/task_types.go#L119), 
[v1beta1.StepState](https://github.com/tektoncd/pipeline/blob/50ed02b4c2b96656355548acea878a0d20e89750/pkg/apis/pipeline/v1beta1/taskrun_types.go#L258), 
PodReference and 
log of the step. 

The controller updates TaskRun resource to save its emission state for each event in `metadata.annotation`.  

Although identity of each CloudEvent is always guaranteed in CloudEvent ID attribute,  
sometimes duplication of event emission occurs due to known issue#8.


---

### Appendix
`tekton-step-observer` uses knative/pkg for its infrastructure to compose a basis of kubernetes controller.  
They provide several metrics such as reconciliation count, time, etc.
For measuring any affection on reconciliation, the [config/monitoring](./config/monitoring) directory contains Prometheus Scrape config.


The example of CloudEvent is:
```
 ☁️  cloudevents.Event
 Validation: valid
 Context Attributes,
   specversion: 1.0
   type: dev.tekton.event.plugin.step.succeeded.v1
   source: github.com/tom24d/step-observe-controller
   id: 50df8d70-8c11-4cd4-8fdf-3aa4cb287408-unnamed-0-dev.tekton.event.plugin.step.succeeded.v1
   time: 2020-08-12T12:07:54Z
   datacontenttype: application/json
 Extensions,
   knativearrivaltime: 2020-08-12T12:07:55.3392277Z
   knativehistory: inmemorychannel-kne-trigger-kn-channel.test-event-assertion-success-fail-skip-skip-in-memory-chan7f5q7.svc.cluster.local
 Data,
   {
     "podRef": {
       "kind": "Pod",
       "namespace": "test-event-assertion-success-fail-skip-skip-in-memory-chan7f5q7",
       "name": "e2e-test-step-observed-run-pod-wrpkl",
       "apiVersion": "v1"
     },
     "log": "hello1\n",
     "step": {
       "name": "",
       "image": "busybox@sha256:895ab622e92e18d6b461d671081757af7dbaa3b00e3e28e12505af7817f73649",
       "command": [
         "/bin/sh"
       ],
       "args": [
         "-c",
         "echo hello1"
       ],
       "resources": {}
     },
     "stepState": {
       "terminated": {
         "exitCode": 0,
         "reason": "Completed",
         "startedAt": "2020-08-12T12:07:54Z",
         "finishedAt": "2020-08-12T12:07:54Z",
         "containerID": "containerd://3548dcaf6287d647cbd775d9de6488e2024cbb8f436f71f78a1e4e8bab0eece5"
       },
       "name": "unnamed-0",
       "container": "step-unnamed-0",
       "imageID": "docker.io/library/busybox@sha256:895ab622e92e18d6b461d671081757af7dbaa3b00e3e28e12505af7817f73649"
     }
   }
```
