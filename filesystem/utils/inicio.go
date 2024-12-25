package utils

import (
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"golang.org/x/exp/mmap"
)

func init() {
	utils_general.ConfigurarLogger("filesystem.log")

	validarExistenciaArchivos()

	var err error
	archivoBitmap, err = mmap.Open(FilesystemConfig.Mount_dir + "bitmap.dat")
	if err != nil {
		panic(err)
	}
	defer archivoBitmap.Close()

	cargarBitmap()
}
