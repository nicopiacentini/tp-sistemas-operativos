package utils

import (
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"log"
	"net/http"
)

func mandarACpuParaEjecutar(pid int32, tid int32) {
	hiloAEjecutar := ExecuteRequestCPU{
		Pid: pid,
		Tid: tid,
	}
	tengoElControl = false

	utils_general.LoggearMensaje(KernelConfig.Log_level, "Se manda a ejecutar el hilo (PID:<%d>,TID:<%d>)", hiloEnEjecucion.Pid, hiloEnEjecucion.Tid)
	if KernelConfig.Scheduler_algorithm == "CMN" {
		utils_general.PostRequestNoBloqueante(hiloAEjecutar, KernelConfig.Ip_cpu, KernelConfig.Port_cpu, "ejecutarHilo")
		go timerMultinivel(pid, tid)
	} else {
		utils_general.PostRequest(hiloAEjecutar, KernelConfig.Ip_cpu, KernelConfig.Port_cpu, "ejecutarHilo")
	}
}

func THREAD_RETURNED(w http.ResponseWriter, r *http.Request) { // CPU devuelve Tid con motivo de devolución
	var req ThreadReturnedRequestCpu
	utils_general.HandlePostRequest(w, r, &req)
	w.WriteHeader(http.StatusOK)
	tengoElControl = true

	if implicaReplanificar(req.Motivo) {
		desalojar()
		evaluarInterrupcionReplanificacion(eventoReplanificacionObligatorio)
	} else if req.Motivo == "Segmentation Fault" {
		planificarLargoPlazo("PROCESS_EXIT", Params{Hilo: *hiloEnEjecucion})
	}
}

func enviarInterrupcionACpu(pid int32, tid int32) {
	interrupcion := InterruptionRequestCpu{
		Pid: pid,
		Tid: tid,
	}
	tengoElControl = false
	utils_general.PostRequest(interrupcion, KernelConfig.Ip_cpu, KernelConfig.Port_cpu, "interrupcion")
}

// SYSCALLS //
func PROCESS_CREATE(w http.ResponseWriter, r *http.Request) {
	logSyscall()
	var req ProcessCreateRequestCpu
	utils_general.HandlePostRequest(w, r, &req)
	w.WriteHeader(http.StatusOK)

	crearProceso(req.Archivo, req.Tamaño, req.Prioridad)
}

func PROCESS_EXIT(w http.ResponseWriter, r *http.Request) {
	logSyscall()
	var req RequestVacia
	utils_general.HandlePostRequest(w, r, &req)
	w.WriteHeader(http.StatusOK)

	if hiloEnEjecucion.Tid == 0 {
		planificarLargoPlazo("PROCESS_EXIT", Params{Pid: hiloEnEjecucion.Pid})
		return
	} else {
		log.Println("No puede finalizar el proceso un hilo que no sea el principal (TID != 0)")
	}
}

func THREAD_CREATE(w http.ResponseWriter, r *http.Request) {
	logSyscall()
	var req ThreadCreateRequestCpu
	utils_general.HandlePostRequest(w, r, &req)

	go planificarLargoPlazo("THREAD_CREATE", Params{Archivo: req.Archivo, Prioridad: req.Prioridad, Pid: hiloEnEjecucion.Pid})

	w.WriteHeader(http.StatusOK)
}

func THREAD_JOIN(w http.ResponseWriter, r *http.Request) {
	logSyscall()
	var req ThreadJoinRequestCpu
	utils_general.HandlePostRequest(w, r, &req)
	w.WriteHeader(http.StatusOK)

	if hiloEnEjecucion.Tid != req.Tid {

		if existeHilo(hiloEnEjecucion.Pid, req.Tid) || finalizo(hiloEnEjecucion.Pid, req.Tid) { // buscar en la lista general de pcbs
			añadirABloqueadosPor(hiloEnEjecucion.Pid, req.Tid, *hiloEnEjecucion)
			bloquear(*hiloEnEjecucion, "THREAD_JOIN")
		} else {
			utils_general.LoggearMensaje(KernelConfig.Log_level, "Hilo con PID %d y TID %d no encontrado", hiloEnEjecucion.Pid, req.Tid)
		}
	} else {
		utils_general.LoggearMensaje(KernelConfig.Log_level, "No se puede hacer join a sí mismo")
	}
}

func THREAD_CANCEL(w http.ResponseWriter, r *http.Request) {
	logSyscall()
	var req ThreadCancelRequestCpu
	utils_general.HandlePostRequest(w, r, &req)
	w.WriteHeader(http.StatusOK)

	planificarLargoPlazo("THREAD_CANCEL", Params{Tid: req.Tid, Pid: hiloEnEjecucion.Pid})
}

func THREAD_EXIT(w http.ResponseWriter, r *http.Request) {
	logSyscall()
	var req RequestVacia
	utils_general.HandlePostRequest(w, r, &req)
	w.WriteHeader(http.StatusOK)

	planificarLargoPlazo("THREAD_EXIT", Params{})
}

func MUTEX_CREATE(w http.ResponseWriter, r *http.Request) {
	logSyscall()
	var req MutexRequestCpu
	utils_general.HandlePostRequest(w, r, &req)
	w.WriteHeader(http.StatusOK)

	if !existeMutex(hiloEnEjecucion.Pid, req.Recurso) {
		go crearMutex(hiloEnEjecucion.Pid, req.Recurso)
	} else {
		utils_general.LoggearMensaje(KernelConfig.Log_level, "Mutex %s ya existe en proceso con PID %d", req.Recurso, hiloEnEjecucion.Pid)
	}
	mandarACpuParaEjecutar(hiloEnEjecucion.Pid, hiloEnEjecucion.Tid)
}

func MUTEX_LOCK(w http.ResponseWriter, r *http.Request) {
	logSyscall()
	var req MutexRequestCpu
	utils_general.HandlePostRequest(w, r, &req)
	w.WriteHeader(http.StatusOK)

	if existeMutex(hiloEnEjecucion.Pid, req.Recurso) {
		if mutexDisponible(hiloEnEjecucion.Pid, req.Recurso) {
			asignarMutex(*hiloEnEjecucion, req.Recurso)
			mandarACpuParaEjecutar(hiloEnEjecucion.Pid, hiloEnEjecucion.Tid)
		} else {
			bloquearPorMutex(*hiloEnEjecucion, req.Recurso)
		}
	} else {
		utils_general.LoggearMensaje(KernelConfig.Log_level, "Mutex %s no encontrado en proceso con PID %d", req.Recurso, hiloEnEjecucion.Pid)
	}
}

func MUTEX_UNLOCK(w http.ResponseWriter, r *http.Request) {
	logSyscall()
	var req MutexRequestCpu
	utils_general.HandlePostRequest(w, r, &req)
	w.WriteHeader(http.StatusOK)

	if existeMutex(hiloEnEjecucion.Pid, req.Recurso) && esDueño(*hiloEnEjecucion, req.Recurso) {
		liberarMutex(hiloEnEjecucion.Pid, req.Recurso)
	}
}

func DUMP_MEMORY(w http.ResponseWriter, r *http.Request) {
	logSyscall()
	var req RequestVacia
	utils_general.HandlePostRequest(w, r, &req)
	w.WriteHeader(http.StatusOK)
	pid := hiloEnEjecucion.Pid
	tid := hiloEnEjecucion.Tid

	bloquear(*hiloEnEjecucion, "DUMP")
	dumpearProceso(pid, tid)
}

func IO(w http.ResponseWriter, r *http.Request) {
	logSyscall()
	var req IORequestCpu
	utils_general.HandlePostRequest(w, r, &req)
	w.WriteHeader(http.StatusOK)

	hiloBloqueado := hiloEnEjecucion
	go bloquear(*hiloEnEjecucion, "IO")
	ejecutarIO(req.Tiempo, hiloBloqueado.Pid, hiloBloqueado.Tid)
	desbloquear(*hiloBloqueado)
}
