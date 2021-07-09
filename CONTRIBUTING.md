# Contributing

Love to accept your patches and contributions to this project. If you have a change in mind, please [fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo) the repository and work in your local fork. Once tested, submit the changes via a [pull request](https://docs.github.com/en/github/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request) and wait for the review.

## Issues
### Reporting an Issue
* Make sure to test against the latest version of Kubernetes. It is possible that the bug you're experiencing might have been fixed already.
* Provide steps to reproduce the issue, and if possible please include the expected results as well as the actual results.
* Stale issues will be closed periodically.

## Testing
A Kubernetes cluster is needed either remote or minikube works too. When testing from local machine, the K8s-Discovery reads the KubeConfig from local and authenticates and authorizes as the user in the local KubeConfig. 

I normally use this [testing project](https://github.com/gkarthiks/testingproj-k8s-discovery) for validating. To run from within the cluster, add this module to a container and validate the output from logs.