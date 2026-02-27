package tenant

import "fmt"

// Plan กำหนด resource quota สำหรับแต่ละ tier
type Plan struct {
	Name string

	// ResourceQuota
	CPURequests string
	CPULimits   string
	MemRequests string
	MemLimits   string
	MaxPods     string
	MaxPVC      string
	MaxStorage  string

	// LimitRange per-container defaults
	DefaultCPURequest    string
	DefaultCPULimit      string
	DefaultMemRequest    string
	DefaultMemLimit      string
	MaxCPUPerContainer   string
	MaxMemPerContainer   string
	MaxCPUPerPod         string
	MaxMemPerPod         string
	MaxPVCSize           string
}

// Plans คือแผนทั้งหมดที่มีในระบบ
var Plans = map[string]Plan{
	"starter": {
		Name:        "starter",
		CPURequests: "1",
		CPULimits:   "2",
		MemRequests: "2Gi",
		MemLimits:   "4Gi",
		MaxPods:     "20",
		MaxPVC:      "5",
		MaxStorage:  "50Gi",

		DefaultCPURequest:    "100m",
		DefaultCPULimit:      "500m",
		DefaultMemRequest:    "128Mi",
		DefaultMemLimit:      "512Mi",
		MaxCPUPerContainer:   "1",
		MaxMemPerContainer:   "2Gi",
		MaxCPUPerPod:         "2",
		MaxMemPerPod:         "4Gi",
		MaxPVCSize:           "20Gi",
	},
	"growth": {
		Name:        "growth",
		CPURequests: "2",
		CPULimits:   "4",
		MemRequests: "4Gi",
		MemLimits:   "8Gi",
		MaxPods:     "40",
		MaxPVC:      "10",
		MaxStorage:  "200Gi",

		DefaultCPURequest:    "200m",
		DefaultCPULimit:      "1",
		DefaultMemRequest:    "256Mi",
		DefaultMemLimit:      "1Gi",
		MaxCPUPerContainer:   "2",
		MaxMemPerContainer:   "4Gi",
		MaxCPUPerPod:         "4",
		MaxMemPerPod:         "8Gi",
		MaxPVCSize:           "50Gi",
	},
	"enterprise": {
		Name:        "enterprise",
		CPURequests: "4",
		CPULimits:   "8",
		MemRequests: "8Gi",
		MemLimits:   "16Gi",
		MaxPods:     "80",
		MaxPVC:      "20",
		MaxStorage:  "1Ti",

		DefaultCPURequest:    "500m",
		DefaultCPULimit:      "2",
		DefaultMemRequest:    "512Mi",
		DefaultMemLimit:      "2Gi",
		MaxCPUPerContainer:   "4",
		MaxMemPerContainer:   "8Gi",
		MaxCPUPerPod:         "8",
		MaxMemPerPod:         "16Gi",
		MaxPVCSize:           "100Gi",
	},
}

// GetPlan คืน plan หรือ error ถ้า plan ไม่มีอยู่
func GetPlan(name string) (Plan, error) {
	plan, ok := Plans[name]
	if !ok {
		return Plan{}, fmt.Errorf("plan %q ไม่มีในระบบ — ใช้ได้: starter, growth, enterprise", name)
	}
	return plan, nil
}

// ListPlans คืนรายชื่อ plan ทั้งหมด
func ListPlans() []string {
	return []string{"starter", "growth", "enterprise"}
}
