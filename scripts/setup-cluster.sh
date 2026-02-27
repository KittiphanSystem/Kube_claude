#!/usr/bin/env bash
# สร้าง Kind cluster และติดตั้ง ArgoCD, ingress-nginx, cert-manager
# ใช้: ./scripts/setup-cluster.sh
set -e

CLUSTER_NAME="${CLUSTER_NAME:-kube-saas}"

echo "==> Creating Kind cluster: $CLUSTER_NAME"
kind create cluster --name "$CLUSTER_NAME" --wait 2m || true

echo "==> Installing ArgoCD"
kubectl create namespace argocd 2>/dev/null || true
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl wait --for=condition=available deployment/argocd-server -n argocd --timeout=180s 2>/dev/null || true

echo "==> Installing ingress-nginx (Helm)"
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx 2>/dev/null || true
helm repo update
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.ingressClassResource.name=nginx \
  --set controller.ingressClass=nginx \
  --wait --timeout 3m 2>/dev/null || true

echo "==> Installing cert-manager"
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml
kubectl wait --for=condition=available deployment/cert-manager -n cert-manager --timeout=120s 2>/dev/null || true
kubectl wait --for=condition=available deployment/cert-manager-webhook -n cert-manager --timeout=120s 2>/dev/null || true

echo ""
echo "✅ Cluster ready. Next steps:"
echo "  1. Set your Git repo URL: export REPO_URL=https://github.com/YOUR_ORG/YOUR_REPO.git"
echo "  2. Replace placeholder in repo: make replace-repo"
echo "  3. Commit and push platform-gitops"
echo "  4. Apply bootstrap: make bootstrap"
echo "  5. Get ArgoCD admin password: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d"
echo "  6. Port-forward ArgoCD: kubectl port-forward svc/argocd-server -n argocd 8080:443"
