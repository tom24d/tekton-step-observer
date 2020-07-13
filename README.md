# step-observe-controller  
[WIP] Observer plugin to report CloudEvents at step execution level for Tekton  

## Concept  

This plugin creates a controller to watch TaskRun/Pod(TBD), then send CloudEvents to specified `default-cloud-event-sink` when it detects a change of state of each step.  

Supported CloudEvents event type are:
```
dev.tekton.event.step.started.v1  
dev.tekton.event.step.failed.v1
dev.tekton.event.step.succeeded.v1
dev.tekton.event.step.skipped.v1
```

The shape of CloudEvent data is:
```json
{
    "pod": v1.Pod,
    "logs": "2020/07/11 10:13:31 Error executing command: exec: "exit": executable file not found in $PATH",
    "status: v1beta1.StepStatus,
    "details": v1beta1.Step,
}
```

More info will come soon.  
