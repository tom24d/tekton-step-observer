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

More info will come soon.  
