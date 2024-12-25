package utils

import (
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"log"
	"net/http"
)

// Función que recibe el PID y TID desde el Kernel
func RecibirSolicitudKernel(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	var SK solicitudKernel
	utils_general.HandlePostRequest(w, r, &SK)
	utils_general.LoggearMensaje(CpuConfig.Log_level, "## Llega hilo para ejecutar - PID: <%d> - TID: <%d>", SK.Pid, SK.Tid)

	// Guardamos el pid y tid recibidos
	guardarPidTid(SK.Pid, SK.Tid) // Fuerzo el gourotineo

	solicitarContextoDeMemoria(SK.Pid, SK.Tid)
}

func PROCESS_CREATE(archivo string, tamaño int, prioridad int) {
	solicitud := ProcessCreateRequestCpu{
		Archivo:   archivo,
		Tamaño:    tamaño,
		Prioridad: prioridad,
	}

	utils_general.PostRequest(solicitud, CpuConfig.Ip_kernel, CpuConfig.Port_kernel, "crearProceso")
}

func PROCESS_EXIT() {

	utils_general.PostRequest(RequestVacia{Syscall: "PROCESS_EXIT"}, CpuConfig.Ip_kernel, CpuConfig.Port_kernel, "finalizarProceso")
}

func THREAD_CREATE(archivo string, prioridad int) {
	solicitud := ThreadCreateRequestCpu{
		Archivo:   archivo,
		Prioridad: prioridad,
	}

	utils_general.PostRequest(solicitud, CpuConfig.Ip_kernel, CpuConfig.Port_kernel, "crearHilo")
}

func THREAD_JOIN(tidJoin int) {
	solicitud := ThreadJoinRequestCpu{
		Tid: tidJoin,
	}

	utils_general.PostRequest(solicitud, CpuConfig.Ip_kernel, CpuConfig.Port_kernel, "bloquearHilo")
}

func THREAD_CANCEL(tidCancel int) {
	solicitud := ThreadCancelRequestCpu{
		Tid: tidCancel,
	}

	utils_general.PostRequest(solicitud, CpuConfig.Ip_kernel, CpuConfig.Port_kernel, "matarHilo")
}

func THREAD_EXIT() {
	utils_general.PostRequest(RequestVacia{Syscall: "THREAD_EXIT"}, CpuConfig.Ip_kernel, CpuConfig.Port_kernel, "suicidarHilo")
}

func MUTEX_CREATE(recurso string) {
	solicitud := MutexRequestCpu{
		Recurso: recurso,
	}

	utils_general.PostRequest(solicitud, CpuConfig.Ip_kernel, CpuConfig.Port_kernel, "crearMutex")
}

func MUTEX_LOCK(recurso string) {
	solicitud := MutexRequestCpu{
		Recurso: recurso,
	}

	utils_general.PostRequest(solicitud, CpuConfig.Ip_kernel, CpuConfig.Port_kernel, "bloquearMutex")
}

func MUTEX_UNLOCK(recurso string) {
	solicitud := MutexRequestCpu{
		Recurso: recurso,
	}

	utils_general.PostRequest(solicitud, CpuConfig.Ip_kernel, CpuConfig.Port_kernel, "desbloquearMutex")
}

func DUMP_MEMORY() {
	utils_general.PostRequest(RequestVacia{Syscall: "DUMP_MEMORY"}, CpuConfig.Ip_kernel, CpuConfig.Port_kernel, "dumpearProceso")
}

func IO(tiempo int) {
	solicitud := IORequestCpu{
		Tiempo: tiempo,
	}

	utils_general.PostRequest(solicitud, CpuConfig.Ip_kernel, CpuConfig.Port_kernel, "entradaSalida")
}

// notificarInterrupcionAlKernel envía una solicitud al Kernel sobre la interrupción
func notificarInterrupcionAlKernel(tid int32, motivo string) {
	log.Println("Notificando interrupción al Kernel")

	// Crear la solicitud
	respuesta := RespuestaKernel{
		Tid:    tid,
		Motivo: motivo,
	}

	utils_general.PostRequest(respuesta, CpuConfig.Ip_kernel, CpuConfig.Port_kernel, "retornoHilo")
}

// Función para manejar las interrupciones enviadas desde el Kernel
func ObtenerInterrupcionDelKernel(w http.ResponseWriter, r *http.Request) {
	utils_general.LoggearMensaje(CpuConfig.Log_level, "## Llega interrupción al puerto Interrupt") // LOG OBLIGATORIO

	var interrupcion Interrupcion

	// Utilizar HandlePostRequest para manejar la deserialización y la verificación del método
	utils_general.HandlePostRequest(w, r, &interrupcion)
	w.WriteHeader(http.StatusOK)

	var PidTidKey PidTid = PidTid(interrupcion)

	// Almacenar la interrupción en el mapa global de interrupciones
	muInterrupt.Lock()
	interrupciones[PidTidKey] = true
	muInterrupt.Unlock()
}
