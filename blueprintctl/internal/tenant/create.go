package tenant

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CreateOptions ตัวเลือกสำหรับสร้าง tenant ใหม่
type CreateOptions struct {
	Name      string
	PlanName  string
	RepoURL   string
	Domain    string
	Email     string
	OutputDir string // path ของ platform-gitops repo
	DryRun    bool   // แค่ preview ไม่สร้างจริง
}

// Validate ตรวจสอบ input ก่อนสร้าง
func (o *CreateOptions) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("--name ต้องระบุชื่อ tenant")
	}
	if !isValidName(o.Name) {
		return fmt.Errorf("ชื่อ tenant %q ไม่ถูกต้อง — ใช้ได้เฉพาะ lowercase, ตัวเลข, และ hyphen", o.Name)
	}
	if o.RepoURL == "" {
		return fmt.Errorf("--repo ต้องระบุ Git repo URL")
	}
	if o.Domain == "" {
		return fmt.Errorf("--domain ต้องระบุ domain name")
	}
	if o.Email == "" {
		return fmt.Errorf("--email ต้องระบุ contact email")
	}
	if _, err := GetPlan(o.PlanName); err != nil {
		return err
	}
	return nil
}

// isValidName ตรวจสอบว่าชื่อ tenant ถูกต้อง (kubernetes naming convention)
func isValidName(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return name[0] != '-' && name[len(name)-1] != '-'
}

// Create สร้าง tenant bootstrap files ทั้งหมด
func Create(opts *CreateOptions) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("validation ล้มเหลว: %w", err)
	}

	cfg, err := NewTenantConfig(opts.Name, opts.PlanName, opts.RepoURL, opts.Domain, opts.Email)
	if err != nil {
		return fmt.Errorf("สร้าง config ล้มเหลว: %w", err)
	}

	tenantDir := filepath.Join(opts.OutputDir, "tenants", "tenant-"+opts.Name)

	// ตรวจสอบว่า tenant มีอยู่แล้วหรือไม่
	if _, err := os.Stat(tenantDir); !os.IsNotExist(err) {
		return fmt.Errorf("tenant %q มีอยู่แล้วที่ %s — ใช้ 'tenant sync' เพื่ออัปเดต", opts.Name, tenantDir)
	}

	if opts.DryRun {
		fmt.Printf("=== DRY RUN: สร้าง tenant %q (plan: %s) ===\n", opts.Name, opts.PlanName)
		fmt.Printf("Output dir: %s\n", tenantDir)
		fmt.Printf("Domain: %s\n", opts.Domain)
		fmt.Printf("Repo: %s\n", opts.RepoURL)
		fmt.Printf("Contact: %s\n", opts.Email)
		fmt.Println("\nFiles ที่จะสร้าง:")
		for _, f := range []string{"tenant.yaml", "bootstrap/namespace.yaml", "bootstrap/rbac.yaml",
			"bootstrap/quota.yaml", "bootstrap/limitrange.yaml", "bootstrap/networkpolicy.yaml",
			"bootstrap/argocd.yaml", "bootstrap/ingress.yaml"} {
			fmt.Printf("  %s/%s\n", tenantDir, f)
		}
		return nil
	}

	// สร้างไฟล์ทั้งหมด
	fmt.Printf("กำลังสร้าง tenant %q...\n", opts.Name)
	if err := RenderAllTemplates(cfg, tenantDir); err != nil {
		// Cleanup หากสร้างไม่สำเร็จ
		_ = os.RemoveAll(tenantDir)
		return fmt.Errorf("render templates ล้มเหลว: %w", err)
	}

	fmt.Printf("✅ สร้าง tenant %q เสร็จสมบูรณ์\n", opts.Name)
	fmt.Printf("   → %s\n", tenantDir)
	fmt.Printf("\nขั้นตอนต่อไป:\n")
	fmt.Printf("  1. git add tenants/tenant-%s\n", opts.Name)
	fmt.Printf("  2. git commit -m 'feat: onboard tenant %s'\n", opts.Name)
	fmt.Printf("  3. git push\n")
	fmt.Printf("  4. ArgoCD จะ sync อัตโนมัติภายใน 3 นาที\n")
	fmt.Printf("  5. Tenant portal: https://portal.example.com/tenant-%s\n", opts.Name)

	return nil
}

// Delete ลบ tenant bootstrap files (ต้อง confirm)
func Delete(name, outputDir string, force bool) error {
	if !isValidName(name) {
		return fmt.Errorf("ชื่อ tenant ไม่ถูกต้อง: %q", name)
	}

	tenantDir := filepath.Join(outputDir, "tenants", "tenant-"+name)

	if _, err := os.Stat(tenantDir); os.IsNotExist(err) {
		return fmt.Errorf("tenant %q ไม่พบที่ %s", name, tenantDir)
	}

	if !force {
		return fmt.Errorf("ต้องใช้ --force เพื่อยืนยันการลบ tenant %q", name)
	}

	if err := os.RemoveAll(tenantDir); err != nil {
		return fmt.Errorf("ลบ tenant dir ล้มเหลว: %w", err)
	}

	fmt.Printf("✅ ลบ tenant %q เสร็จแล้ว\n", name)
	fmt.Printf("   กรุณา commit และ push เพื่อให้ ArgoCD ลบ resources จาก cluster\n")
	return nil
}

// List แสดง tenant ทั้งหมด
func List(outputDir string) ([]string, error) {
	tenantsDir := filepath.Join(outputDir, "tenants")
	entries, err := os.ReadDir(tenantsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("อ่าน tenants dir ล้มเหลว: %w", err)
	}

	var tenants []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "tenant-") {
			name := strings.TrimPrefix(e.Name(), "tenant-")
			if name != "template" { // ข้าม template
				tenants = append(tenants, name)
			}
		}
	}
	return tenants, nil
}
