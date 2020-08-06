# step-observe-controller  
[WIP] Observer plugin to report CloudEvents at step execution level for Tekton  

## Concept  

This plugin creates a controller to watch TaskRun/Pod(TBD), then send CloudEvents to specified `default-cloud-event-sink` when it detects a change of state of each step.  

Supported CloudEvents event type are:
```
dev.tekton.event.plugin.step.started.v1  
dev.tekton.event.plugin.step.failed.v1
dev.tekton.event.plugin.step.succeeded.v1
dev.tekton.event.plugin.step.skipped.v1
```

The example of CloudEvent is:
```json
 ☁️  cloudevents.Event
 Validation: valid
 Context Attributes,
   specversion: 1.0
   type: dev.tekton.events.plugin.step.succeeded.v1
   source: /api/v1/namespaces/tekton-pipelines/deployments/tekton-step-observer
   id: 4706aaf2-7432-47ce-8d44-e51ac692a9c5
   time: 2020-08-03T07:37:47Z
   datacontenttype: application/json
 Data,
   {
     "podRef": {
       "kind": "Pod",
       "namespace": "default",
       "name": "hello1-run-bhx2s-pod-8hsvm",
       "apiVersion": "v1"
     },
     "log": "Hello step 1\n",
     "step": {
       "name": "hello1",
       "image": "ubuntu",
       "command": [
         "echo"
       ],
       "args": [
         "Hello step 1"
       ],
       "resources": {}
     },
     "stepState": {
       "terminated": {
         "exitCode": 0,
         "reason": "Completed",
         "startedAt": "2020-08-03T07:37:47Z",
         "finishedAt": "2020-08-03T07:37:47Z",
         "containerID": "containerd://9e6bdf350ef6c6744b7b920771a2475b4a595210dc0c6b6f5f0e50bb57413308"
       },
       "name": "hello1",
       "container": "step-hello1",
       "imageID": "docker.io/library/ubuntu@sha256:5d1d5407f353843ecf8b16524bc5565aa332e9e6a1297c73a92d3e754b8a636d"
```

More info will come soon.  
