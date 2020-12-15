# Global-Scheduler Code Process Analysis

## 1. Scheduling Process

### 1.1 Flow Chart

![flowchart](/images/global_scheduler_allocations_flowchart.jpg)

### 1.2 Mind Map

![mindmap](/images/global_scheduler_allocations_mindmap.jpg)



## 2. Differences from Kube-Scheduler

Unlike kube-scheduler, which assigns pods on the appropriate nodes, global-scheduler allocates VMs to particular clusters. The differences between the global-scheduler and the kube-scheduler will be explained from two aspects: “Resources” (what the scheduling according to) and "Scheduling Framework" (how does the scheduling work).

### 2.1 Cache(Resources)

Kube-Scheduler uses list-watch mode to sense resource changes and update the cache; Global-Scheduler retrieves the resource information by periodically requesting resource API and then updates it’s cache.

### 2.2 Scheduling Framework

The scheduling framework of global-scheduler is plugin framework learned from kube-scheduler. The difference is that each scheduling request of kube-scheduler is split into two phases, the **scheduling cycle** and the **binding cycle**, while global-scheduler only has the scheduling cycle stage and returns the scheduling results to the upper application after resource reservation. Global-Scheduler don't perform the specific bind operation.

In the scheduling cycle, the process is roughly the same, with 2-step operation including **"Predicates"** (equivalent to "Filter") and **"Priority"** (equivalent to "Score"). It needs to be explained that the bind extension of the global-scheduler corresponds to the reserve extension of the kube-scheduler. As for the specific implementation of the plugin in each extension, it may be different or similar, which is not discussed here.

![kube-scheduler&global-scheduler](/images/kube-scheduler&global-scheduler.jpg)



