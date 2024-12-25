package utils

import (
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"net/http"
)

// Función que solicita el contexto a Memoria basado en el PID y TID
func solicitarContextoDeMemoria(pid int32, tid int32) {
	// Crear el cuerpo de la solicitud HTTP con el PID y TID
	solicitud := SolicitudMemoria{
		Pid: pid,
		Tid: tid,
	}

	utils_general.LoggearMensaje(CpuConfig.Log_level, "## TID: <%d> - Solicito Contexto Ejecución", tid) // LOG OBLIGATORIO

	utils_general.PostRequest(solicitud, CpuConfig.Ip_memory, CpuConfig.Port_memory, "contextoEjecucion")
}

func RecibirContextoEjecucion(w http.ResponseWriter, r *http.Request) {
	var Respuesta RespuestaContexto
	utils_general.HandlePostRequest(w, r, &Respuesta)
	w.WriteHeader(http.StatusOK)

	guardarContexto(Respuesta.Pid, Respuesta.Tid, Respuesta.Contexto)

	Fetch(Respuesta.Pid, Respuesta.Tid, Respuesta.Contexto.ContextoRegistros.PC)
}

func RecibirInstruccion(w http.ResponseWriter, r *http.Request) {
	var respuesta RespuestaFetch
	utils_general.HandlePostRequest(w, r, &respuesta)
	w.WriteHeader(http.StatusOK)

	Execute(respuesta.Pid, respuesta.Tid, respuesta.Instruccion)
}

func actualizarContextoEjecucion(pid int32, tid int32, contexto Contexto) error {
	// Crear la solicitud de actualización de contexto
	solicitud := RespuestaContexto{
		Pid:      pid,
		Tid:      tid,
		Contexto: contexto,
	}

	utils_general.LoggearMensaje(CpuConfig.Log_level, "## TID: <%d> - Actualizo Contexto Ejecución", tid) // LOG OBLIGATORIO

	// Enviar la solicitud utilizando PostRequest del paquete utils_general
	utils_general.PostRequest(solicitud, CpuConfig.Ip_memory, CpuConfig.Port_memory, "actualizarContexto")

	return nil
}
