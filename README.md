# Kube SaaS Platform

ระบบ Multi-tenant Kubernetes ตาม Blueprint ในโฟลเดอร์ `Doc/` — ประกอบด้วย **blueprintctl** (CLI) และ **platform-gitops** (GitOps manifests). ใช้ Makefile และสคริปต์ให้รันได้ทันทีบน Kind cluster.

## โครงสร้างโปรเจกต์

```
Kube_claude/
├── Makefile                # build, replace-repo, bootstrap, cluster-setup, tenant-*
├── scripts/setup-cluster.sh  # สร้าง Kind + ArgoCD + ingress-nginx + cert-manager
├── blueprintctl/           # CLI สำหรับ onboard จัดการ tenant (Go)
├── platform-gitops/        # GitOps repo (ใช้กับ ArgoCD)
│   ├── clusters/prod/      # bootstrap.yaml
│   ├── argocd/             # projects, applications
│   ├── platform/           # ingress-nginx, cert-manager
│   └── tenants/            # tenant-<name> (สร้างด้วย blueprintctl)
└── Doc/                    # เอกสารและ reference
```

## ใช้งานได้เลย (Quick Start)

### ขั้นตอนที่ 1: กำหนด Git Repo ของ platform-gitops

โปรเจกต์นี้ต้อง push ขึ้น Git (GitHub/GitLab) เพื่อให้ ArgoCD ดึงได้

```bash
# สร้าง repo บน GitHub/GitLab แล้ว
export REPO_URL=https://github.com/YOUR_ORG/YOUR_REPO.git

# แทนที่ __REPO_URL__ ในทุกไฟล์ของ platform-gitops (ทำครั้งเดียว ก่อน push ครั้งแรก)
make replace-repo

# จากนั้น commit และ push ทั้งโปรเจกต์ (รวม platform-gitops)
git add .
git commit -m "chore: set platform-gitops repo URL"
git remote add origin "$REPO_URL"   # ถ้ายังไม่มี
git push -u origin main
```

### ขั้นตอนที่ 2: สร้าง Cluster และติดตั้ง ArgoCD

```bash
# ต้องมี: kind, kubectl, helm
make cluster-setup
```

จะได้: Kind cluster, ArgoCD, ingress-nginx (Helm), cert-manager

### ขั้นตอนที่ 3: Apply Bootstrap

```bash
export REPO_URL=https://github.com/YOUR_ORG/YOUR_REPO.git
make bootstrap
```

ArgoCD จะ sync จาก repo ของคุณ (root-app → platform-services, tenants-appset)

### ขั้นตอนที่ 4: ใช้ blueprintctl จัดการ tenant

```bash
# Build CLI (ครั้งแรก)
make build

# ดู plans
make plan-list

# สร้าง tenant ใหม่
make tenant-create name=acme plan=starter \
  repo=https://github.com/acme/k8s-apps.git \
  domain=app.acme.com email=dev@acme.com

# ดู tenant ทั้งหมด
make tenant-list
```

จากนั้น `git add platform-gitops/tenants/tenant-acme` แล้ว commit & push — ApplicationSet จะสร้าง Application สำหรับ tenant นั้นให้อัตโนมัติ

### (ถ้าต้องการ) เข้า ArgoCD UI

```bash
# รหัส admin ครั้งแรก
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d

# Port-forward
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

เปิด https://localhost:8080 (รับรอง self-signed ได้)

---

## คำสั่ง Make ที่ใช้บ่อย

| คำสั่ง | คำอธิบาย |
|--------|----------|
| `make help` | แสดงคำสั่งทั้งหมด |
| `make build` | Build blueprintctl |
| `make replace-repo` | แทน __REPO_URL__ ด้วย REPO_URL (ต้อง export REPO_URL ก่อน) |
| `make bootstrap` | Apply bootstrap ลง cluster (ต้อง export REPO_URL ก่อน) |
| `make cluster-setup` | สร้าง Kind cluster + ArgoCD + ingress-nginx + cert-manager |
| `make tenant-list` | แสดง tenant ทั้งหมด |
| `make tenant-create name=... plan=... repo=... domain=... email=...` | สร้าง tenant |
| `make plan-list` | แสดง plans (starter, growth, enterprise) |

---

## คำสั่ง blueprintctl โดยตรง

ถ้าไม่ใช้ Make ให้ส่ง `--repo-path` ไปที่โฟลเดอร์ platform-gitops:

```bash
./blueprintctl/blueprintctl plan list
./blueprintctl/blueprintctl tenant list --repo-path platform-gitops
./blueprintctl/blueprintctl tenant create --name myco --plan starter \
  --repo https://github.com/myco/apps.git --domain app.myco.com --email dev@myco.com \
  --repo-path platform-gitops
```

---

## แหล่งอ้างอิง

- สถาปัตยกรรมและขั้นตอนติดตั้งละเอียด: `Doc/Kube_Saas/README.md`, `Doc/Kube_Saas/SETUP.md`
- UI mockups: `Doc/admin-portal.html`, `Doc/tenant-portal.html`
