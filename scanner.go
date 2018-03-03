package scanner

import (
  "github.com/golang/glog"
  "os"
  "path"
  "path/filepath"
  "runtime"
  "strings"
  "time"
)

type FileScanner struct {
  Path          string
  StableTimeout float64
  OutChannel    chan string
  Whitelist     []string
  Blacklist *map[string]bool
}

func (w FileScanner) Scan() {
  glog.V(2).Info(" ", w.Path)
  err := filepath.Walk(w.Path, process_files(w))
  if err != nil {
    glog.V(2).Info(err)
  }
  w.OutChannel <- "__EOF"
}

func (w FileScanner) isLock(pth string, f os.FileInfo) bool {
  if osType() == "windows" {
    err := os.Rename(pth, pth)
    if err != nil {
      glog.V(2).Info("File still locked", err)
      return false
    }
    return true
  } else {
    mod := f.ModTime()
    if time.Now().Sub(mod).Seconds() < w.StableTimeout {
      glog.V(2).Info("Locked", pth)
      return false
    }
    return true
  }
}

func osType() string {
  return runtime.GOOS
}

func (w FileScanner) isWhiteListed(basePath string, curFilePath string) bool {
  for _, folder := range w.Whitelist {
    if strings.Contains(curFilePath, path.Join(basePath, folder)) {
      if (*w.Blacklist)[curFilePath] {
        return false
      }
      return true
    }
  }
  return false
}

func process_files(w FileScanner) filepath.WalkFunc {
  return func(pth string, info os.FileInfo, err error) error {
    if !info.IsDir() {
      if w.isWhiteListed(w.Path, pth) && w.isLock(pth, info) {
        w.OutChannel <- pth
      } else {
        glog.V(2).Info("Path ", pth, " not in whitelist")
      }
    }
    return nil
  }
}
