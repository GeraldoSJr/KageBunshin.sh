#!/bin/bash

minikube start --memory 1800mb
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
kubectl apply -f ./conf/deployment.yaml
