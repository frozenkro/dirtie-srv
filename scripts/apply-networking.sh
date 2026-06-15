#!/usr/bin/env bash
set -euo pipefail

# Apply networking options for dirtie-srv on k3s
# Run this on rpic1 (or wherever you have kubectl configured for the cluster)

echo "=== Applying NodePort service (rpic1:30080) ==="
kubectl apply -f k8s/31-service-nodeport.yaml

echo ""
echo "=== Option A complete ==="
echo "  You can now hit the app directly at:"
echo "     http://rpic1:30080"
echo ""
echo "  Test it:"
echo "     curl http://rpic1:30080/"
echo ""

# Uncomment below if you want LoadBalancer (needed for Tailscale LB / MetalLB) ---
# echo "=== Applying LoadBalancer service ==="
# kubectl apply -f k8s/35-service-loadbalancer.yaml
# echo "  After an external LB gives out an IP, the service will appear on that IP:80"

# Uncomment below if you have Traefik CRDs working
# echo "=== Applying path-based Traefik IngressRoute ==="
# kubectl apply -f k8s/45-ingressroute-path.yaml
# echo "  You can now hit: http://rpic1/dirtie/"
