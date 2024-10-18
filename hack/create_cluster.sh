#!/bin/bash

minikube start --memory 1800mb
kubectl apply -f ./conf/deployment.yaml
