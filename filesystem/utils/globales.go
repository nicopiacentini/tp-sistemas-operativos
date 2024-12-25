package utils

import (
	"github.com/sisoputnfrba/tp-golang/filesystem/globals"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"golang.org/x/exp/mmap"
	"sync"
)

type Metadata struct {
	IndexBlock int `json:"index_block"`
	Size       int `json:"size"`
}

type DumpMemoryRequest struct {
	Nombre    string `json:"nombre"`
	Tamaño    int    `json:"tamaño"`
	Contenido []byte `json:"contenido"`
}

type DumpMemoryResponse struct {
	Pid     int32 `json:"pid"`
	Tid     int32 `json:"tid"`
	Success bool  `json:"success"`
}

var (
	FilesystemConfig *globals.Config = utils_general.IniciarConfig("configs/config.json", &globals.Config{}).(*globals.Config)
	bitmap           []byte          = make([]byte, FilesystemConfig.Block_count)
	archivoBitmap    *mmap.ReaderAt
	mutexBitmap      sync.Mutex
	mutexBloques     sync.Mutex
)
