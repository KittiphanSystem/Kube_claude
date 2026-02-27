# Kube SaaS Platform — Makefile
REPO_PATH ?= $(shell pwd)/platform-gitops
REPO_URL ?= __REPO_URL__
BLUEPRINTCTL := $(shell pwd)/blueprintctl/blueprintctl

.PHONY: build build-blueprintctl replace-repo bootstrap tenant-list tenant-create cluster-setup help

help:
	@echo "Kube SaaS — คำสั่งที่ใช้บ่อย"
	@echo ""
	@echo "  make build              Build blueprintctl"
	@echo "  make replace-repo        แทน __REPO_URL__ ด้วย REPO_URL ใน platform-gitops (ใช้ก่อน push ครั้งแรก)"
	@echo "  make bootstrap          Apply bootstrap ลง cluster (ต้อง set REPO_URL)"
	@echo "  make cluster-setup      สร้าง Kind cluster + ติดตั้ง ArgoCD, ingress-nginx, cert-manager"
	@echo "  make tenant-list        แสดง tenant ทั้งหมด"
	@echo "  make tenant-create      สร้าง tenant (ใช้กับ name= plan= repo= domain= email=)"
	@echo ""
	@echo "ตัวอย่าง:"
	@echo "  export REPO_URL=https://github.com/yourorg/platform-gitops.git"
	@echo "  make replace-repo"
	@echo "  make cluster-setup"
	@echo "  make bootstrap"
	@echo "  make tenant-create name=acme plan=starter repo=https://github.com/acme/apps.git domain=app.acme.com email=dev@acme.com"

build: build-blueprintctl

build-blueprintctl:
	cd blueprintctl && go build -o blueprintctl ./cmd/blueprintctl
	@echo "Built: blueprintctl/blueprintctl"

replace-repo:
	@if [ "$(REPO_URL)" = "__REPO_URL__" ] || [ -z "$(REPO_URL)" ]; then \
		echo "Error: กำหนด REPO_URL ก่อน เช่น export REPO_URL=https://github.com/you/repo.git"; exit 1; \
	fi
	@echo "Replacing __REPO_URL__ with $(REPO_URL) in platform-gitops..."
	@for f in platform-gitops/clusters/prod/bootstrap.yaml \
		platform-gitops/argocd/projects/platform.yaml \
		platform-gitops/argocd/applications/root-app.yaml \
		platform-gitops/argocd/applications/platform-services.yaml \
		platform-gitops/argocd/applications/tenants-appset.yaml; do \
		sed "s|__REPO_URL__|$(REPO_URL)|g" "$$f" > "$$f.tmp" && mv "$$f.tmp" "$$f"; \
	done
	@echo "Done. กรุณา commit และ push platform-gitops"

bootstrap:
	@if [ "$(REPO_URL)" = "__REPO_URL__" ] || [ -z "$(REPO_URL)" ]; then \
		echo "Error: กำหนด REPO_URL ก่อน เช่น export REPO_URL=https://github.com/you/repo.git"; exit 1; \
	fi
	sed "s|__REPO_URL__|$(REPO_URL)|g" platform-gitops/clusters/prod/bootstrap.yaml | kubectl apply -f -
	@echo "Bootstrap applied. ArgoCD will sync from $(REPO_URL)"

cluster-setup:
	chmod +x scripts/setup-cluster.sh
	./scripts/setup-cluster.sh

tenant-list: build-blueprintctl
	$(BLUEPRINTCTL) tenant list --repo-path "$(REPO_PATH)"

tenant-create: build-blueprintctl
	@if [ -z "$(name)" ] || [ -z "$(repo)" ] || [ -z "$(domain)" ] || [ -z "$(email)" ]; then \
		echo "Usage: make tenant-create name=xxx plan=starter repo=https://... domain=app.xxx.com email=dev@xxx.com"; exit 1; \
	fi
	$(BLUEPRINTCTL) tenant create --name "$(name)" --plan "$(plan)" --repo "$(repo)" --domain "$(domain)" --email "$(email)" --repo-path "$(REPO_PATH)"

plan-list: build-blueprintctl
	$(BLUEPRINTCTL) plan list
