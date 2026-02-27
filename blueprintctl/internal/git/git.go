package git

import (
	"fmt"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CommitAndPush commit ไฟล์ทั้งหมดและ push ไปยัง remote
func CommitAndPush(repoPath, message string) error {
	r, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("เปิด git repo ที่ %s ล้มเหลว: %w", repoPath, err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("เปิด worktree ล้มเหลว: %w", err)
	}

	// Stage ทุกไฟล์ที่เปลี่ยน
	if err := w.AddGlob("."); err != nil {
		return fmt.Errorf("git add ล้มเหลว: %w", err)
	}

	// ตรวจสอบ status ก่อน commit
	status, err := w.Status()
	if err != nil {
		return fmt.Errorf("ดู git status ล้มเหลว: %w", err)
	}

	if status.IsClean() {
		fmt.Println("ℹ️  ไม่มีการเปลี่ยนแปลง — ข้าม commit")
		return nil
	}

	// Commit
	commit, err := w.Commit(message, &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "blueprintctl",
			Email: "blueprintctl@kube-saas.local",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("git commit ล้มเหลว: %w", err)
	}

	obj, err := r.CommitObject(commit)
	if err != nil {
		return fmt.Errorf("ดู commit object ล้มเหลว: %w", err)
	}
	fmt.Printf("✅ Committed: %s\n", obj.Hash)

	// Push
	if err := r.Push(&gogit.PushOptions{}); err != nil {
		if err == gogit.NoErrAlreadyUpToDate {
			fmt.Println("ℹ️  Already up to date")
			return nil
		}
		return fmt.Errorf("git push ล้มเหลว: %w\nTip: ตรวจสอบ SSH key / credentials ก่อน push", err)
	}

	fmt.Println("✅ Pushed to remote")
	return nil
}
