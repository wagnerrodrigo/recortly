package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// fileLock evita dois workers ao mesmo tempo (causa renders duplicados).
type fileLock struct {
	file *os.File
}

func acquireLock(storageDir string) (*fileLock, error) {
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(storageDir, ".worker.lock")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("lock %s em uso", path)
	}
	_ = f.Truncate(0)
	_, _ = f.WriteAt([]byte(fmt.Sprintf("%d\n", os.Getpid())), 0)
	_ = os.WriteFile(filepath.Join(storageDir, ".worker.pid"), []byte(fmt.Sprintf("%d", os.Getpid())), 0o644)
	return &fileLock{file: f}, nil
}

func (l *fileLock) Release() {
	if l == nil || l.file == nil {
		return
	}
	name := l.file.Name()
	_ = syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	_ = l.file.Close()
	_ = os.Remove(filepath.Join(filepath.Dir(name), ".worker.pid"))
}
