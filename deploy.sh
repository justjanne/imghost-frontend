#!/bin/sh
IMAGE=k8r.eu/justjanne/imghost-frontend
TAGS=$(git describe --always --tags HEAD)
DEPLOYMENT=imghost-frontend
POD=imghost-frontend

kubectl -n imghost set image deployment/$DEPLOYMENT $POD=$IMAGE:$TAGS