# คู่มือติดตั้งและใช้งาน Kube SaaS Platform

## สิ่งที่ต้องมีก่อนเริ่ม

| เครื่องมือ | Version | ใช้ทำอะไร |
|-----------|---------|-----------|
| kubectl | 1.28+ | คุม cluster |
| helm | 3.x | ติดตั้ง chart |
| git | any | GitOps |
| go | 1.21+ | build blueprintctl |
| domain | - | ชี้ DNS ไปที่ LoadBalancer IP |

---

## Step 1 — ติดตั้ง Platform Baseline

### 1.1 ติดตั้ง ArgoCD

```bash
kubectl create namespace argocd
kubectl apply -n argocd \
  -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# รอ ArgoCD พร้อม
kubectl wait --for=condition=available deployment/argocd-server \
  -n argocd --timeout=120s

# ดู initial admin password
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d
```

### 1.2 Bootstrap Platform via ArgoCD

```bash
# Clone repo
git clone https://git.example.com/platform-gitops.git
cd platform-gitops

# แก้ไข URL ใน clusters/prod/bootstrap.yaml ให้ตรงกับ repo จริง
sed -i 's|https://git.example.com/platform-gitops.git|YOUR_REPO_URL|g' \
  clusters/prod/bootstrap.yaml

# Apply root app — ArgoCD จะ sync ทุกอย่างให้อัตโนมัติ
kubectl apply -f clusters/prod/bootstrap.yaml
```

### 1.3 ตั้งค่า DNS

หลัง ingress-nginx ได้ LoadBalancer IP ให้ชี้ DNS:

```bash
# ดู External IP ของ admin ingress
kubectl get svc ingress-nginx-admin-lb -n ingress-nginx

# ดู External IP ของ tenant ingress
kubectl get svc ingress-nginx-tenant-lb -n ingress-nginx
```

ตั้งค่า DNS records:

```
# Admin tools → admin LoadBalancer IP
admin.example.com     A   <ADMIN_LB_IP>
argo.example.com      A   <ADMIN_LB_IP>
grafana.example.com   A   <ADMIN_LB_IP>

# Tenant portal + apps → tenant LoadBalancer IP
portal.example.com    A   <TENANT_LB_IP>
*.example.com         A   <TENANT_LB_IP>
```

### 1.4 ตั้งค่า cert-manager email

```bash
# แก้ email ใน ClusterIssuer ก่อน apply
vim platform/cert-manager/cert-manager.yaml
# เปลี่ยน admin@example.com เป็น email จริงของคุณ
```

---

## Step 2 — ติดตั้ง blueprintctl CLI

```bash
cd blueprintctl
go build -o blueprintctl ./cmd/blueprintctl

# ทดสอบ
./blueprintctl --help
./blueprintctl plan list
```

---

## Step 3 — Onboard Tenant ใหม่

### คำสั่งเดียว จบ:

```bash
./blueprintctl tenant create \
  --name acme-corp \
  --plan growth \
  --repo https://github.com/acme-corp/k8s-apps.git \
  --domain app.acme-corp.com \
  --email devops@acme-corp.com \
  --push
```

### หรือ Dry Run ดูก่อน:

```bash
./blueprintctl tenant create \
  --name acme-corp \
  --plan growth \
  --repo https://github.com/acme-corp/k8s-apps.git \
  --domain app.acme-corp.com \
  --email devops@acme-corp.com \
  --dry-run
```

### ดู tenant ทั้งหมด:

```bash
./blueprintctl tenant list

# Output:
# TENANT               PLAN         DOMAIN                         PORTAL URL
# ─────────────────────────────────────────────────────────────────
# acme-corp            growth        app.acme-corp.com              https://portal.example.com/tenant-acme-corp
# startup-x            starter       app.startup-x.io               https://portal.example.com/tenant-startup-x
```

### ดู plans:

```bash
./blueprintctl plan list

# Output:
# PLAN         CPU_REQ  CPU_LIM  MEM_REQ    MEM_LIM    PODS   STORAGE
# ─────────────────────────────────────────────────────────────────
# starter      1        2        2Gi        4Gi        20     50Gi
# growth       2        4        4Gi        8Gi        40     200Gi
# enterprise   4        8        8Gi        16Gi       80     1Ti
```

---

## Portal Links — แยกชัดเจน

### 🔐 Admin Portal (Platform Team เท่านั้น)

| URL | เข้าด้วย | ใช้ทำอะไร |
|-----|---------|-----------|
| `https://admin.example.com` | SSO / Devtron account | คุม cluster, RBAC, onboard tenant |
| `https://argo.example.com` | ArgoCD account | ดู GitOps sync, manage apps |
| `https://grafana.example.com` | Grafana admin account | ดู metrics ทุก tenant |

### 👤 Tenant Portal (Tenant User)

| URL | เข้าด้วย | ใช้ทำอะไร |
|-----|---------|-----------|
| `https://portal.example.com/tenant-<n>` | Token จาก admin | ดู deploy status, logs ของตัวเอง |
| `https://grafana.example.com/d/tenant-<n>` | Anonymous (embed) | metrics เฉพาะ namespace ตัวเอง |
| `https://app.acme-corp.com` | N/A | แอป tenant เอง |

### ข้อแตกต่างหลัก:

```
Admin Portal (admin.example.com)
├── Ingress Class: nginx-admin  (dedicated controller)
├── เห็น: ทุก namespace, ทุก cluster
├── สิทธิ์: สร้าง/ลบ resource ได้ทั้ง cluster
└── Access: เฉพาะ platform team (IP whitelist + SSO)

Tenant Portal (portal.example.com/tenant-<n>)
├── Ingress Class: nginx  (tenant controller)
├── เห็น: เฉพาะ namespace tenant-<n>
├── สิทธิ์: ดูได้ หรือ sync ArgoCD app ของตัวเองเท่านั้น
└── Access: token ที่ admin ออกให้
```

---

## Checklist ก่อน Production

- [ ] เปลี่ยน `example.com` เป็น domain จริงใน YAML ทั้งหมด
- [ ] ตั้ง email จริงใน ClusterIssuer (cert-manager)
- [ ] ตั้ง Git repo URL จริงใน bootstrap.yaml และ AppProject
- [ ] ตั้ง IP whitelist สำหรับ admin ingress (แทน `0.0.0.0/0`)
- [ ] ตั้ง SSO (Dex/OIDC) สำหรับ ArgoCD และ Grafana
- [ ] ทดสอบ NetworkPolicy: DNS ผ่าน, cross-namespace ไม่ผ่าน
- [ ] ทดสอบ ResourceQuota: tenant สร้าง pod เกิน quota ไม่ได้
- [ ] ทดสอบ AppProject: tenant deploy ไป namespace อื่นไม่ได้
- [ ] Backup etcd สม่ำเสมอ
- [ ] ตั้ง cluster autoscaler (ถ้าใช้ cloud)

---

## Troubleshooting

### TLS ไม่ออก certificate
```bash
kubectl describe certificate -n tenant-<n>
kubectl describe challenge -n tenant-<n>
# ตรวจสอบ HTTP-01 challenge ผ่าน ingress หรือไม่
```

### ArgoCD sync ล้มเหลว
```bash
argocd app get tenant-<n>-apps
argocd app sync tenant-<n>-apps --force
```

### NetworkPolicy block traffic
```bash
# ทดสอบ DNS จาก pod
kubectl exec -n tenant-<n> <pod> -- nslookup kubernetes.default.svc.cluster.local
# ทดสอบ internet
kubectl exec -n tenant-<n> <pod> -- curl -v https://httpbin.org/get
```

### blueprintctl build error
```bash
cd blueprintctl
go mod tidy
go build ./...
```
