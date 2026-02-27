# Kube SaaS Platform — Production Blueprint

## สถาปัตยกรรมภาพรวม

```
Internet
   │
   ▼
ingress-nginx (LoadBalancer)
   ├── admin.example.com  → Devtron Dashboard (Admin Portal)
   ├── argo.example.com   → ArgoCD UI (Admin only)
   ├── grafana.example.com → Grafana (Admin + per-tenant)
   └── *.example.com      → Tenant Apps (แยก namespace)

Platform Stack (namespace: platform tools)
   ├── cert-manager       → TLS อัตโนมัติ (Let's Encrypt)
   ├── ingress-nginx      → Reverse Proxy / Load Balancer
   ├── ArgoCD             → GitOps Engine (App-of-Apps)
   ├── Devtron            → Admin Portal (RBAC + Multi-cluster)
   └── kube-prometheus    → Monitoring (Prometheus + Grafana)

Per-Tenant (namespace: tenant-<n>)
   ├── ResourceQuota      → จำกัด CPU/RAM/Pods
   ├── LimitRange         → Default requests/limits
   ├── NetworkPolicy      → Isolate + allow DNS/Ingress
   ├── RBAC               → Role/RoleBinding แยก tenant
   └── ArgoCD Application → Deploy จาก repo ของ tenant
```

## Portal Links (แยกชัดเจน)

| Portal | URL | สำหรับใคร | ใช้ทำอะไร |
|--------|-----|-----------|-----------|
| **Admin Portal** | `https://admin.example.com` | Platform Team | คุม cluster ทั้งหมด, RBAC, onboard tenant |
| **ArgoCD UI** | `https://argo.example.com` | Admin | ดู/คุม GitOps sync, apps |
| **Grafana (Admin)** | `https://grafana.example.com` | Admin | ดู metrics ทุก tenant |
| **Tenant Portal** | `https://portal.example.com/tenant-<n>` | Tenant User | ดูเฉพาะของตัวเอง (logs/deploy status) |
| **Grafana (Tenant)** | `https://grafana.example.com/d/tenant-<n>` | Tenant User | metrics เฉพาะ namespace ตัวเอง |

## ขั้นตอนติดตั้ง (Quick Start)

```bash
# 1. Clone repo นี้
git clone https://git.example.com/platform-gitops.git
cd platform-gitops

# 2. ติดตั้ง Platform Baseline
kubectl apply -f clusters/prod/bootstrap.yaml

# 3. ArgoCD จะ sync ทุกอย่างให้อัตโนมัติ

# 4. เพิ่ม Tenant ด้วย CLI
cd blueprintctl
go build -o blueprintctl ./cmd/blueprintctl
./blueprintctl tenant create \
  --name tenant-a \
  --plan starter \
  --repo https://git.example.com/tenant-a-apps.git \
  --domain a.example.com \
  --email admin@tenant-a.com
```

## โครงสร้างโฟลเดอร์

```
kube-saas/
├── platform-gitops/          # Repo หลัก (platform team)
│   ├── argocd/               # ArgoCD projects + apps
│   ├── platform/             # Shared services (nginx, cert, monitor)
│   ├── tenants/              # Per-tenant configs
│   └── clusters/             # Cluster-level bootstrap
├── blueprintctl/             # CLI สำหรับ onboard tenant
└── docs/                     # เอกสารเพิ่มเติม
```
