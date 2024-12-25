package utils

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"log"
	"net/http"
)

func inicializarProcesoEnMemoria(tamaño int, pid int32) {
	var procesoRequest ProcessRequestMemory
	procesoRequest.Tamaño = tamaño
	procesoRequest.Pid = pid

	utils_general.PostRequest(procesoRequest, KernelConfig.Ip_memory, KernelConfig.Port_memory, "iniciarProceso")
}

func RespuestaInicializacionProceso(w http.ResponseWriter, r *http.Request) {
	var procesoRequest ResponseMemory
	utils_general.HandlePostRequest(w, r, &procesoRequest)
	w.WriteHeader(http.StatusOK)
	// El problema está acá:
	switch procesoRequest.Codigo {
	case 0: // Hay espacio en memoria de forma contigua para inicializar el proceso
		proceso := quitarDeNew()
		if len(colaNew) == 0 { // Quitar el proceso de la cola NEW
			creacionDeHilos(paramsInitMemoria[proceso.Pid].PrioridadTid0, paramsInitMemoria[proceso.Pid].Archivo, proceso.Pid)
		} // Crear el hilo principal del proceso
		if len(colaNew) > 0 {
			procesoSiguiente := colaNew[0]
			creacionDeHilosNoReplanificadora(paramsInitMemoria[proceso.Pid].PrioridadTid0, paramsInitMemoria[proceso.Pid].Archivo, proceso.Pid)
			reintentarInicializacion(paramsInitMemoria[procesoSiguiente.Pid].Tamaño, procesoSiguiente.Pid)
		}
	case 1: // No hay espacio continuo pero sí en huecos
		pausarPlanificacionCortoPlazo()
		solicitarCompactacion()
	case 2: // No hay espacio ni siquiera compactando la memoria
		proceso := colaNew[0]
		utils_general.LoggearMensaje(KernelConfig.Log_level, "Memoria insuficiente para proceso con PID %d", proceso.Pid)
		if hiloEnEjecucion != nil {
			if tengoElControl {
				go mandarACpuParaEjecutar(hiloEnEjecucion.Pid, hiloEnEjecucion.Tid)
			}
		} else {
			eventoReplanificacionOpcional()
		}
	default:
		log.Fatal("Código de respuesta inválido")
	}
}

func reintentarInicializacion(tamaño int, pid int32) {
	log.Printf("Reintento inicializar PID: %d, Tamaño: %d\n", pid, tamaño)
	inicializarProcesoEnMemoria(tamaño, pid)
}

func intentarInicializarProcesos() {
	if len(colaNew) > 0 {
		procesoInicial := colaNew[0]
		reintentarInicializacion(paramsInitMemoria[procesoInicial.Pid].Tamaño, procesoInicial.Pid)
	}
}

func solicitarCompactacion() {
	utils_general.PostRequest(CompactacionRequest{Estado: "Vacio"}, KernelConfig.Ip_memory, KernelConfig.Port_memory, "compactacion")
}

func RespuestaCompactacionFinalizada(w http.ResponseWriter, r *http.Request) {
	// Estructura para recibir la compactación
	var compactacion CompactacionRequest

	// Utilizamos la función genérica para manejar la solicitud
	utils_general.HandlePostRequest(w, r, &compactacion)

	// Verificamos si el campo fue decodificado correctamente
	fmt.Printf("Estado recibido en la compactación: %s\n", compactacion.Estado)

	if compactacion.Estado != "Finalizada" {
		fmt.Println("Estado inválido recibido en la compactación.")
		http.Error(w, "Estado inválido en la compactación", http.StatusBadRequest)
		return
	}

	// Continuar con la lógica después de recibir y deserializar correctamente
	w.WriteHeader(http.StatusOK) // Confirmar que todo está correcto

	proceso := colaNew[0]

	reintentarInicializacion(paramsInitMemoria[proceso.Pid].Tamaño, proceso.Pid)

	reanudarPlanificacionCortoPlazo()
}

func finalizarProcesoEnMemoria(pid int32) {
	procesoRequest := ProcessFinishMemory{
		Pid: pid,
	}
	utils_general.PostRequest(procesoRequest, KernelConfig.Ip_memory, KernelConfig.Port_memory, "finalizarProceso")
}

func finalizarEnMemoriaPorDump(pid int32) {
	procesoRequest := ProcessFinishMemory{
		Pid: pid,
	}
	utils_general.PostRequest(procesoRequest, KernelConfig.Ip_memory, KernelConfig.Port_memory, "finalizarProcesoDump")
}

func RespuestaFinalizacionProceso(w http.ResponseWriter, r *http.Request) {
	var procesoRequest ResponseMemory
	utils_general.HandlePostRequest(w, r, &procesoRequest)
	w.WriteHeader(http.StatusOK)

	if procesoRequest.Codigo == 0 {
		finalizarHilosDeProceso(procesoRequest.Pid)
		liberarPcb(procesoRequest.Pid)
		moverTcbAExit(*hiloEnEjecucion)
		hiloEnEjecucion = nil
		chequearQuitarColas()
		evaluarPausaPlanificacion()

		utils_general.LoggearMensaje(KernelConfig.Log_level, "## (<%d>:<%d>) Finaliza el hilo", procesoRequest.Pid, 0) // LOG OBLIGATORIO
		utils_general.LoggearMensaje(KernelConfig.Log_level, "## Finaliza el proceso <%d>", procesoRequest.Pid)        // LOG OBLIGATORIO

		if len(colaNew) == 0 {
			evaluarInterrupcionReplanificacion(eventoReplanificacionObligatorio) // Siempre replanificar al finalizar un proceso
		} else {
			intentarInicializarProcesos()
		}

	} else {
		utils_general.LoggearMensaje(KernelConfig.Log_level, "Error al finalizar proceso con PID %d", procesoRequest.Pid)
	}
}

func RespuestaFinalizacionProcesoDump(w http.ResponseWriter, r *http.Request) {
	var procesoRequest ResponseMemory
	utils_general.HandlePostRequest(w, r, &procesoRequest)
	w.WriteHeader(http.StatusOK)

	if procesoRequest.Codigo == 0 {
		finalizarHilosDeProceso(procesoRequest.Pid)
		liberarPcb(procesoRequest.Pid)
		chequearQuitarColas()
		evaluarPausaPlanificacion()

		utils_general.LoggearMensaje(KernelConfig.Log_level, "## (<%d>:<%d>) Finaliza el hilo", procesoRequest.Pid, 0) // LOG OBLIGATORIO
		utils_general.LoggearMensaje(KernelConfig.Log_level, "## Finaliza el proceso <%d>", procesoRequest.Pid)        // LOG OBLIGATORIO

		if len(colaNew) == 0 {
			return // Si el dump no fue exitoso mato al proceso, de todos modos, el hilo que hizo el dump va a estar bloqueado así que ni replanifico
		} else {
			intentarInicializarProcesos()
		}

	} else {
		utils_general.LoggearMensaje(KernelConfig.Log_level, "Error al finalizar proceso con PID %d", procesoRequest.Pid)
	}
}

func inicializarHiloEnMemoria(pathArchivo string, pid int32, tid int32) { // por las dudas le paso el tid aunque no sé si es necesario
	var hiloRequest ThreadRequestMemory
	hiloRequest.Archivo = pathArchivo
	hiloRequest.Pid = pid
	hiloRequest.Tid = tid

	utils_general.PostRequest(hiloRequest, KernelConfig.Ip_memory, KernelConfig.Port_memory, "iniciarHilo")

	hiloSolicitado.L.Lock()
	hiloSolicitado.Signal()
	hiloSolicitado.L.Unlock()
}

func finalizarHiloEnMemoria(pid int32, tid int32) {
	var hiloRequest ThreadFinishMemory
	hiloRequest.Pid = pid
	hiloRequest.Tid = tid

	utils_general.PostRequest(hiloRequest, KernelConfig.Ip_memory, KernelConfig.Port_memory, "finalizarHilo")
}

func dumpearProceso(pid int32, tid int32) {
	var procesoRequest ProcessDumpRequestMemory
	procesoRequest.Pid = pid
	procesoRequest.Tid = tid

	utils_general.PostRequest(procesoRequest, KernelConfig.Ip_memory, KernelConfig.Port_memory, "dumpearProceso")
}

func RespuestaDump(w http.ResponseWriter, r *http.Request) {
	var procesoRequest ProcessDumpResponseMemory
	utils_general.HandlePostRequest(w, r, &procesoRequest)
	w.WriteHeader(http.StatusOK)
	tcbHilo := buscarTcb(procesoRequest.Pid, procesoRequest.Tid)

	if procesoRequest.Success {
		desbloquear(*tcbHilo)
	} else {
		finalizarPorDump(procesoRequest.Pid)
	}
}
