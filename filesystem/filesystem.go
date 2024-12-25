package main

import (
	"github.com/sisoputnfrba/tp-golang/filesystem/utils"
	"net/http"
	"strconv"
)

func main() {
	// Obtengo el puerto del filesystem
	puerto := strconv.Itoa(utils.FilesystemConfig.Port)

	http.HandleFunc("/solicitudDump", utils.AtenderSolicitudDump)

	// Inicio el servidor
	err := http.ListenAndServe(":"+puerto, nil)
	if err != nil {
		panic(err)
	}
}
