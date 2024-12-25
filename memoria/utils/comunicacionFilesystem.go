package utils

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"net/http"
	"time"
)

func DumpearProceso(w http.ResponseWriter, r *http.Request) {
	var solicitudDump ProcessExitRequestCpu
	utils_general.HandlePostRequest(w, r, &solicitudDump)
	w.WriteHeader(http.StatusOK)

	timestamp := time.Now().Format("20060102-150405")

	contenido := buscarContenido(solicitudDump.Pid)

	nombreArchivo := fmt.Sprintf("%d-%d-%s.dmp", solicitudDump.Pid, solicitudDump.Tid, timestamp)

	solicitudFileS := FileRequest{
		Nombre:    nombreArchivo,
		Tamaño:    int(memoria.MemoriaSistema[solicitudDump.Pid].Limite),
		Contenido: contenido,
	}

	utils_general.LoggearMensaje(MemoryConfig.Log_level, "## Memory Dump solicitado - (PID:TID) - (<%d>:<%d>)", solicitudDump.Pid, solicitudDump.Tid) //Log obligatorio

	utils_general.PostRequest(solicitudFileS, MemoryConfig.Ip_Filesystem, MemoryConfig.Port_Filesystem, "solicitudDump")
}

func RespuestaFileSystem(w http.ResponseWriter, r *http.Request) {
	var respuestaDump DumpMemoryResponse
	utils_general.HandlePostRequest(w, r, &respuestaDump)
	w.WriteHeader(http.StatusOK)

	utils_general.PostRequest(respuestaDump, MemoryConfig.Ip_Kernel, MemoryConfig.Port_Kernel, "procesoDumpeado")
}

// devuelve un array de bytes  con todos los contenidos del proceso en memoria de usuario
func buscarContenido(Pid int32) []byte {
	base := memoria.MemoriaSistema[Pid].Base
	limite := memoria.MemoriaSistema[Pid].Limite
	return memoria.MemoriaUsuario[base : base+limite]
}
