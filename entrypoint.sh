#!/bin/bash

if [ ! -z "$KUBERNETES_TOKEN" ]; then
	K8S_TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
fi

if [ ! -z "$KUBERNETES_SERVICE_HOST" ]; then
	K8S_API="https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT"
fi


kubectl config set-cluster k8s --server=$K8S_API --insecure-skip-tls-verify=true
kubectl config set-credentials k8s-mattermost --token=$K8S_TOKEN
kubectl config set-context default-context --cluster=k8s --user=k8s-mattermost
kubectl config use-context default-context
kubectl get pod 

sed -i "s/your-mattermost.org/$SERVER/g" config.toml.dist
sed -i "s/your-channel/$CHANNEL/g" config.toml.dist
sed -i "s/your-team/$TEAM/g" config.toml.dist
sed -i "s/bot@email.org/$LOGIN/g" config.toml.dist
sed -i "s/averystr0ngpassw0rd/$PASSWORD/g" config.toml.dist

echo ''
cat config.toml.dist

./k8s-mattermost --config config.toml.dist
