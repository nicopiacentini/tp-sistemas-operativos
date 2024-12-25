package utils

import (
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"log"
	"net/http"
)

func AtenderSolicitudDump(w http.ResponseWriter, r *http.Request) {
	var solicitud DumpMemoryRequest
	utils_general.HandlePostRequest(w, r, &solicitud)
	w.WriteHeader(http.StatusOK)
	pid := obtenerPid(solicitud.Nombre)
	tid := obtenerTid(solicitud.Nombre)

	if !hayEspacioDisponible(solicitud.Tamaño) {
		respuesta := DumpMemoryResponse{
			Pid:     pid,
			Tid:     tid,
			Success: false,
		}
		log.Println("No se pudo realizar el dump, bloques insuficientes")
		responderSolicitudDump(respuesta)
	} else {
		tamañoDatos := cantidadBloquesDatosNecesarios(solicitud.Tamaño)
		bloqueIndice, indices := reservarBloques(solicitud.Nombre, tamañoDatos)
		crearArchivoMetadata(solicitud.Nombre, solicitud.Tamaño, bloqueIndice)
		grabarPunterosReservados(bloqueIndice, indices)
		utils_general.LoggearMensaje(FilesystemConfig.Log_level, "## Acceso Bloque - Archivo: <%s> - Tipo Bloque: <INDICE> - Bloque File System <%d>", solicitud.Nombre, bloqueIndice) // LOG OBLIGATORIO
		escribirContenidoEnBloques(solicitud.Nombre, bloqueIndice, solicitud.Contenido)

		respuesta := DumpMemoryResponse{
			Pid:     pid,
			Tid:     tid,
			Success: true,
		}

		responderSolicitudDump(respuesta)
	}
	utils_general.LoggearMensaje(FilesystemConfig.Log_level, "## Fin de solicitud - Archivo: <%s>", solicitud.Nombre) // LOG OBLIGATORIO
}

func responderSolicitudDump(respuesta DumpMemoryResponse) {
	utils_general.PostRequest(respuesta, FilesystemConfig.Ip_memory, FilesystemConfig.Port_memory, "respuestaDump")
}
