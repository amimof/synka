# synka
[![Build Status](https://travis-ci.org/amimof/synka.svg?branch=master)](https://travis-ci.org/amimof/synka) [![codecov](https://codecov.io/gh/amimof/synka/branch/master/graph/badge.svg)](https://codecov.io/gh/amimof/synka) [![synka](https://godoc.org/github.com/amimof/synka?status.svg)](https://godoc.org/github.com/amimof/synka) [![Go Report Card](https://goreportcard.com/badge/github.com/amimof/synka)](https://goreportcard.com/report/github.com/amimof/synka)


Synka is a Kubernetes controller that syncs resources from the cluster it runs in to clusters you configure. 

__♥️ This project is currently in it's early stages of development. Please provide your feedback by opening an issue or a pull request if you want to contribute ♥️__

## Run
```
kubectl apply -f https://raw.githubusercontent.com/amimof/synka/master/deploy/k8s.yaml
```

## Use 
Synka will watch for changes on a default set of cluster resources. You can define your own using the `--informer` flag on the command line. Synka will however not sync anything until a resource contains the `synka.io/sync: true` annotation. 