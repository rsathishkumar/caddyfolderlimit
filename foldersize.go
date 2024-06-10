package foldersize

import (
    "encoding/json"
    "net/http"
    "os"
    "path/filepath"

    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
    caddy.RegisterModule(FolderSize{})
}

// FolderSize implements an admin API endpoint that returns the size of a folder.
type FolderSize struct{}

// CaddyModule returns the Caddy module information.
func (FolderSize) CaddyModule() caddy.ModuleInfo {
    return caddy.ModuleInfo{
        ID:  "admin.api.foldersize",
        New: func() caddy.Module { return new(FolderSize) },
    }
}

// Provision sets up the module.
func (fs *FolderSize) Provision(ctx caddy.Context) error {
    return nil
}

// ServeHTTP handles the HTTP request.
func (fs FolderSize) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
    folder := r.URL.Query().Get("folder")
    if folder == "" {
        http.Error(w, "folder parameter is required", http.StatusBadRequest)
        return nil
    }

    size, err := getFolderSize(folder)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return nil
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "folder": folder,
        "size":   size,
    })

    return nil
}

func getFolderSize(path string) (int64, error) {
    var size int64
    err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            size += info.Size()
        }
        return nil
    })
    return size, err
}

// Interface guards
var (
    _ caddy.Provisioner = (*FolderSize)(nil)
    _ caddyhttp.MiddlewareHandler = (*FolderSize)(nil)
)
