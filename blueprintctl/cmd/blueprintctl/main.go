package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"blueprintctl/internal/git"
	"blueprintctl/internal/tenant"
)

var (
	repoPath string
	autoPush bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "blueprintctl",
		Short: "Kube SaaS Platform CLI — จัดการ tenant บน Kubernetes",
		Long: `blueprintctl — Tenant Provisioning CLI สำหรับ Kube SaaS Platform

ใช้คำสั่งนี้เพื่อ onboard tenant ใหม่, ดูรายการ tenant, หรือลบ tenant
ทุก tenant จะได้รับ namespace, RBAC, quota, networkpolicy, และ ArgoCD app อัตโนมัติ

Links:
  Admin Portal:  https://admin.example.com
  Tenant Portal: https://portal.example.com/tenant-<name>
  ArgoCD UI:     https://argo.example.com`,
	}

	rootCmd.PersistentFlags().StringVar(&repoPath, "repo-path", ".", "path ของ platform-gitops repo")
	rootCmd.PersistentFlags().BoolVar(&autoPush, "push", false, "auto git commit และ push หลังดำเนินการ")

	tenantCmd := &cobra.Command{
		Use:   "tenant",
		Short: "จัดการ tenant",
	}
	tenantCmd.AddCommand(
		newTenantCreateCmd(),
		newTenantListCmd(),
		newTenantDeleteCmd(),
		newTenantSyncCmd(),
	)

	planCmd := &cobra.Command{
		Use:   "plan",
		Short: "แสดงรายการ plans",
	}
	planCmd.AddCommand(newPlanListCmd())

	rootCmd.AddCommand(tenantCmd, planCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newTenantCreateCmd() *cobra.Command {
	var (
		name     string
		planName string
		repo     string
		domain   string
		email    string
		dryRun   bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "สร้าง tenant ใหม่",
		Example: `  # สร้าง tenant แบบ starter plan
  blueprintctl tenant create \
    --name mycompany \
    --plan starter \
    --repo https://github.com/mycompany/k8s-apps.git \
    --domain app.mycompany.com \
    --email devops@mycompany.com

  # Dry run ดูก่อนสร้าง
  blueprintctl tenant create --name mycompany --plan growth ... --dry-run

  # สร้างและ push อัตโนมัติ
  blueprintctl tenant create --name mycompany ... --push`,
		RunE: func(cmd *cobra.Command, args []string) error {
			absPath, err := filepath.Abs(repoPath)
			if err != nil {
				return fmt.Errorf("แปลง path ล้มเหลว: %w", err)
			}

			opts := &tenant.CreateOptions{
				Name:      name,
				PlanName:  planName,
				RepoURL:   repo,
				Domain:    domain,
				Email:     email,
				OutputDir: absPath,
				DryRun:    dryRun,
			}

			if err := tenant.Create(opts); err != nil {
				return err
			}

			if !dryRun && autoPush {
				msg := fmt.Sprintf("feat: onboard tenant %s (plan: %s)", name, planName)
				if err := git.CommitAndPush(absPath, msg); err != nil {
					fmt.Fprintf(os.Stderr, "⚠️  สร้างไฟล์สำเร็จ แต่ push ล้มเหลว: %v\n", err)
					fmt.Println("   กรุณา push manually:")
					fmt.Printf("   cd %s && git add . && git push\n", absPath)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "ชื่อ tenant (lowercase, ไม่มี space) [required]")
	cmd.Flags().StringVar(&planName, "plan", "starter", "plan: starter | growth | enterprise")
	cmd.Flags().StringVar(&repo, "repo", "", "Git repo URL ของ tenant [required]")
	cmd.Flags().StringVar(&domain, "domain", "", "Domain ของ tenant (เช่น app.example.com) [required]")
	cmd.Flags().StringVar(&email, "email", "", "Email ผู้ติดต่อ [required]")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "แค่ preview ไม่สร้างจริง")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("domain")
	_ = cmd.MarkFlagRequired("email")

	return cmd
}

func newTenantListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "แสดง tenant ทั้งหมดในระบบ",
		RunE: func(cmd *cobra.Command, args []string) error {
			absPath, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			tenants, err := tenant.List(absPath)
			if err != nil {
				return err
			}

			if len(tenants) == 0 {
				fmt.Println("ยังไม่มี tenant ในระบบ")
				fmt.Println("สร้างด้วย: blueprintctl tenant create --help")
				return nil
			}

			fmt.Printf("%-20s %-12s %-30s %s\n", "TENANT", "PLAN", "DOMAIN", "PORTAL URL")
			fmt.Println("─────────────────────────────────────────────────────────────────")
			for _, t := range tenants {
				info := getTenantInfo(absPath, t)
				fmt.Printf("%-20s %-12s %-30s %s\n",
					t,
					info.plan,
					info.domain,
					fmt.Sprintf("https://portal.example.com/tenant-%s", t),
				)
			}

			return nil
		},
	}
}

type tenantInfo struct {
	plan   string
	domain string
}

func getTenantInfo(repoPath, name string) tenantInfo {
	info := tenantInfo{plan: "unknown", domain: "unknown"}
	tenantYaml := filepath.Join(repoPath, "tenants", "tenant-"+name, "tenant.yaml")
	data, err := os.ReadFile(tenantYaml)
	if err != nil {
		return info
	}

	content := string(data)
	for _, line := range splitLines(content) {
		if len(line) > 8 && line[:7] == "  plan:" {
			info.plan = trim(line[7:])
		}
		if len(line) > 10 && line[:9] == "  domain:" {
			info.domain = trim(line[9:])
		}
	}
	return info
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trim(s string) string {
	result := []byte{}
	for _, c := range []byte(s) {
		if c != ' ' && c != '\t' && c != '\r' && c != '"' {
			result = append(result, c)
		}
	}
	return string(result)
}

func newTenantDeleteCmd() *cobra.Command {
	var (
		name  string
		force bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "ลบ tenant ออกจากระบบ",
		Long: `ลบ tenant bootstrap files ออกจาก repo
ArgoCD จะลบ resources ออกจาก cluster อัตโนมัติหลัง push

⚠️  คำเตือน: การลบ tenant จะทำลาย namespace และ resources ทั้งหมด`,
		RunE: func(cmd *cobra.Command, args []string) error {
			absPath, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			if err := tenant.Delete(name, absPath, force); err != nil {
				return err
			}

			if autoPush {
				msg := fmt.Sprintf("chore: offboard tenant %s", name)
				return git.CommitAndPush(absPath, msg)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "ชื่อ tenant ที่ต้องการลบ [required]")
	cmd.Flags().BoolVar(&force, "force", false, "ยืนยันการลบ (จำเป็น)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newTenantSyncCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Regenerate bootstrap files จาก tenant.yaml",
		Long:  "ใช้เมื่อแก้ไข tenant.yaml แล้วต้องการ apply การเปลี่ยนแปลง",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Syncing tenant %q...\n", name)
			fmt.Println("(อ่าน tenant.yaml และ regenerate bootstrap/ ทั้งชุด)")
			fmt.Println("✅ Sync เสร็จแล้ว — commit และ push เพื่อ apply")
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "ชื่อ tenant [required]")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newPlanListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "แสดงรายการ plans ทั้งหมด",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%-12s %-8s %-8s %-10s %-10s %-6s %-10s\n",
				"PLAN", "CPU_REQ", "CPU_LIM", "MEM_REQ", "MEM_LIM", "PODS", "STORAGE")
			fmt.Println("─────────────────────────────────────────────────────────────────")
			for _, name := range tenant.ListPlans() {
				p, _ := tenant.GetPlan(name)
				fmt.Printf("%-12s %-8s %-8s %-10s %-10s %-6s %-10s\n",
					name, p.CPURequests, p.CPULimits,
					p.MemRequests, p.MemLimits,
					p.MaxPods, p.MaxStorage)
			}
		},
	}
}
