# Guidance

Using the deployment strategy <Code inline>RollingUpdate</Code> can reduce the risk of downtime. Rolling
updates allow deployments’ update to take place with zero downtime by incrementally updating pods instances
with new ones. The new pods will be scheduled on nodes with available resources.

[Kubernetes Documentation - Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
