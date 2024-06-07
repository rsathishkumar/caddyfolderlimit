package caddyfolderlimit

import (
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "strconv"

    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
    caddy.RegisterModule(FolderLimit{})
}

// FolderLimit is a middleware that limits the size of a folder.
type FolderLimit struct {
    FolderPath string `json:"folder_path"`
    MaxSize    int64  `json:"max_size"`
}

// CaddyModule returns the Caddy module information.
func (FolderLimit) CaddyModule() caddy.ModuleInfo {
    return caddy.ModuleInfo{
        ID:  "http.handlers.folder_limit",
        New: func() caddy.Module { return new(FolderLimit) },
    }
}

// Provision sets up the module.
func (fl *FolderLimit) Provision(ctx caddy.Context) error {
    // Ensure the folder exists
    if _, err := os.Stat(fl.FolderPath); os.IsNotExist(err) {
        return fmt.Errorf("folder does not exist: %s", fl.FolderPath)
    }
    return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (fl FolderLimit) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    size, err := getFolderSize(fl.FolderPath)
    if err != nil {
        return err
    }
    if size > fl.MaxSize {
        http.Error(w, "Folder size limit exceeded", http.StatusForbidden)
        return nil
    }
    return next.ServeHTTP(w, r)
}

// getFolderSize calculates the total size of the folder.
func getFolderSize(path string) (int64, error) {
    var size int64
    err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        size += info.Size()
        return nil
    })
    return size, err
}

// UnmarshalCaddyfile sets up the handler from Caddyfile tokens. Syntax:
// folder_limit {
//     folder_path /path/to/folder
//     max_size 104857600
// }
func (fl *FolderLimit) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
    for d.Next() {
        for d.NextBlock(0) {
            switch d.Val() {
            case "folder_path":
                if !d.Args(&fl.FolderPath) {
                    return d.ArgErr()
                }
            case "max_size":
                var maxSizeStr string
                if !d.Args(&maxSizeStr) {
                    return d.ArgErr()
                }
                maxSize, err := parseSize(maxSizeStr)
                if err != nil {
                    return err
                }
                fl.MaxSize = maxSize
            }
        }
    }
    return nil
}

func parseSize(sizeStr string) (int64, error) {
    sizeStr = strings.ToLower(sizeStr)
    var multiplier int64 = 1
    switch {
    case strings.HasSuffix(sizeStr, "kb"):
        multiplier = 1024
        sizeStr = strings.TrimSuffix(sizeStr, "kb")
    case strings.HasSuffix(sizeStr, "mb"):
        multiplier = 1024 * 1024
        sizeStr = strings.TrimSuffix(sizeStr, "mb")
    case strings.HasSuffix(sizeStr, "gb"):
        multiplier = 1024 * 1024 * 1024
        sizeStr = strings.TrimSuffix(sizeStr, "gb")
    }
    size, err := strconv.ParseInt(sizeStr, 10, 64)
    if err != nil {
        return 0, err
    }
    return size * multiplier, nil
}

// Interface guards
var (
    _ caddyhttp.MiddlewareHandler = (*FolderLimit)(nil)
    _ caddyfile.Unmarshaler       = (*FolderLimit)(nil)
)
