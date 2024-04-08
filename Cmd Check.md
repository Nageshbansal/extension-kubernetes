# [Proposal]: Addtion of Command Execution Check

### Summary

Command Execution checks can be a choice for executing/ running shell commands and verification for experiments.

The idea of this check arises from the "Command Probes" in litmuschaos, an open source chaos engineering Platform

## Intent 

As Checks allows users to run their verifications criterias before, during and after the experiments. Steadybit have plenty number of checks , and specifically for Kubernetes, but it lacks the check which can be used to run bash shell commands within experiment pods or the defined images to check the specific data within a database, parse the value out of a JSON blob that is dumped into a certain path, or check for the existence of a particular string in the service logs.

## Implementation and Design

#### Align Mode
The Align Mode  s executed from within the Extension pod. It is preferred for simple shell commands. It is default mode,and it can be tuned by omitting source field.

![[alignMOde.png]]

**Note: I've already Implemented this Mode, and will raise the PR if community agrees on the Proposal**

##### Demo:
Logs:
![[Pasted image 20240408201256.png]]

Platform: 
![[Pasted image 20240408201228.png]]
### Source Mode

In source mode, the command execution is carried out from within a new pod whose image can be specified. It can be used when application-specific binaries are required.

![[source Mode.png]]

Reference:
- Litmuschaos Docs: https://docs.litmuschaos.io/docs/concepts/probes#cmdprobe