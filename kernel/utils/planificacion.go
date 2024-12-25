package utils

import (
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"log"
	"time"
)

// PLANIFICACIÓN //

func planificarLargoPlazo(syscall string, params Params) {
	switch syscall {
	case "PROCESS_CREATE":
		creacionDeProcesos()
	case "PROCESS_EXIT":
		finalizacionDeProcesos(params.Pid)
	case "THREAD_CREATE":
		creacionDeHilos(params.Prioridad, params.Archivo, params.Pid)
	case "THREAD_EXIT":
		suicidioDeHilo()
	case "THREAD_CANCEL":
		matarHilo(params.Pid, params.Tid)
	}
}

func creacionDeProcesos() {
	if len(colaNew) == 0 {
		return
	}

	// Obtener el primer proceso de la cola NEW
	pcbProceso := colaNew[0]

	go inicializarProcesoEnMemoria(paramsInitMemoria[pcbProceso.Pid].Tamaño, pcbProceso.Pid)
}

func finalizacionDeProcesos(pid int32) {
	finalizarProcesoEnMemoria(pid)
}

func finalizarPorDump(pid int32) {
	finalizarEnMemoriaPorDump(pid)
}

func creacionDeHilos(prioridad int32, path string, pid int32) {
	hilo := crearTcb(prioridad, pid)
	go inicializarHiloEnMemoria(path, pid, hilo.Tid)

	hiloSolicitado.L.Lock()
	hiloSolicitado.Wait()
	hiloSolicitado.L.Unlock()

	utils_general.LoggearMensaje(KernelConfig.Log_level, "## (<%d>:<%d>) Se crea el Hilo - Estado: READY", hilo.Pid, hilo.Tid) // LOG OBLIGATORIO
	agregarAReady(hilo)
	if tengoElControl {
		if hiloEnEjecucion == nil {
			eventoReplanificacionObligatorio()
		} else if tieneDesalojo(KernelConfig.Scheduler_algorithm) && hilo.Prioridad < hiloEnEjecucion.Prioridad {
			eventoReplanificacionOpcional() // Si el algoritmo tiene desalojo, la creación de un hilo implica replanificar
		} else {
			mandarACpuParaEjecutar(hiloEnEjecucion.Pid, hiloEnEjecucion.Tid) // Retoma el curso normal
		}
	} else {
		enviarInterrupcionACpu(hiloEnEjecucion.Pid, hiloEnEjecucion.Tid)
	}
}

func creacionDeHilosNoReplanificadora(prioridad int32, archivo string, pid int32) {
	hilo := crearTcb(prioridad, pid)
	go inicializarHiloEnMemoria(archivo, pid, hilo.Tid)

	hiloSolicitado.L.Lock()
	hiloSolicitado.Wait()
	hiloSolicitado.L.Unlock()

	utils_general.LoggearMensaje(KernelConfig.Log_level, "## (<%d>:<%d>) Se crea el Hilo - Estado: READY", hilo.Pid, hilo.Tid) // LOG OBLIGATORIO
	agregarAReady(hilo)
}

func suicidioDeHilo() {
	finalizarHilo(hiloEnEjecucion.Pid, hiloEnEjecucion.Tid)

	mutexHiloEnEjecucion.Lock()
	hiloEnEjecucion = nil
	mutexHiloEnEjecucion.Unlock()

	evaluarPausaPlanificacion()
	eventoReplanificacionObligatorio() // Cuando muere el hilo en ejecución se debe replanificar
}

func matarHilo(pid int32, tid int32) {
	if hiloEnEjecucion.Pid == pid && hiloEnEjecucion.Tid == tid {
		suicidioDeHilo()
	} else {
		finalizarHilo(pid, tid)

		mutexHiloEnEjecucion.Lock()
		hiloEnEjecucion = nil
		mutexHiloEnEjecucion.Unlock()
	}
}

func finalizarHilo(pid int32, tid int32) {
	finalizarHiloEnMemoria(pid, tid)
	liberarTodosLosMutex(pid, tid)
	tcb := desbloquearHilosBloqueados(pid, tid)
	moverTcbAExit(*tcb)
	utils_general.LoggearMensaje(KernelConfig.Log_level, "## (<%d>:<%d>) Finaliza el hilo", pid, tid) // LOG OBLIGATORIO
}

func evaluarInterrupcionReplanificacion(eventoReplanificacion func()) {
	if KernelConfig.Scheduler_algorithm == "CMN" {
		if len(colasMultinivel) == 0 {
			return
		}
	} else if len(colaReady) == 0 {
		return
	}

	hiloAEjecutar := proximoHiloAEjecutarSimulado()
	if hiloAEjecutar == nil {
		return
	}
	if hiloAEjecutar != hiloEnEjecucion {
		eventoReplanificacion()
	}
}

func planificarCortoPlazo() {
	select {
	case pauseChan <- struct{}{}: // Esperar a que el planificador no esté en pausa
		chequearQuitarColas()
		if len(colaReady) == 0 && KernelConfig.Scheduler_algorithm != "CMN" {
			hiloEnEjecucion = nil
		} else {

			proximoHilo := proximoHiloAEjecutar() // precondición, es un hilo distinto al que está ejecutando
			if proximoHilo == nil {
				<-pauseChan
				return
			}

			mutexHiloEnEjecucion.Lock()
			hiloEnEjecucion = proximoHilo
			mutexHiloEnEjecucion.Unlock()
			go mandarACpuParaEjecutar(hiloEnEjecucion.Pid, hiloEnEjecucion.Tid)
			<-pauseChan // Liberar canal de planificación
		}
	default:
		return
	}
}

func planificarFIFO() *Tcb {
	mutexColaReady.Lock()
	proximoHilo := pop(&colaReady)
	mutexColaReady.Unlock()

	return &proximoHilo
}

func planificarPrioridades() *Tcb {
	proximoHilo := buscarHiloMayorPrioridad()
	if hiloEnEjecucion == nil { // Puede ser que el hilo haya muerto, entonces no hace falta ni comparar con el hilo en ejecución
		quitarDeCola(&colaReady, proximoHilo, &mutexColaReady)
		return &proximoHilo
	}

	if proximoHilo.Prioridad < hiloEnEjecucion.Prioridad {
		enviarInterrupcionACpu(hiloEnEjecucion.Pid, hiloEnEjecucion.Tid)
		return nil
	}

	return hiloEnEjecucion
}

func planificarMultinivel() *Tcb {
	mutexColasMultinivel.Lock()
	defer mutexColasMultinivel.Unlock()

	if hiloEnEjecucion == nil {
		// no hacer nada
	} else if !finalizo(hiloEnEjecucion.Pid, hiloEnEjecucion.Tid) && !hiloBloqueado(hiloEnEjecucion.Pid, hiloEnEjecucion.Tid) { // saco de ejecucion al hilo en ejecucion
		agregarAlPrincipioDeColasMultinivel(*hiloEnEjecucion)
	}

	chequearQuitarColas()

	if len(colasMultinivel) == 0 {
		return nil
	}

	prioridad := proximoHiloMultinivel()
	hiloAEjecutar := colasMultinivel[prioridad][0]

	colasMultinivel[prioridad] = colasMultinivel[prioridad][1:]
	return &hiloAEjecutar
}

func chequearQuitarColas() {
	for clave, slice := range colasMultinivel {
		if len(slice) == 0 {
			delete(colasMultinivel, clave)
		}
	}
}

func proximoHiloMultinivel() int {
	chequearQuitarColas()
	nivelMayorPrioridad := -1
	for _, cola := range colasMultinivel {
		if nivelMayorPrioridad == -1 || int(cola[0].Prioridad) < nivelMayorPrioridad {
			nivelMayorPrioridad = int(cola[0].Prioridad)
		}
	}
	return nivelMayorPrioridad
}

func timerMultinivel(pid int32, tid int32) {
	time.Sleep(time.Duration(KernelConfig.Quantum) * time.Millisecond)
	if hiloEnEjecucion != nil {
		enviarInterrupcionACpu(pid, tid)
		utils_general.LoggearMensaje(KernelConfig.Log_level, "## (<%d>:<%d>) - Desalojado por fin de Quantum", pid, tid) // LOG OBLIGATORIO
	}
}

func proximoHiloAEjecutar() *Tcb {
	var proximoHilo *Tcb
	algoritmo := KernelConfig.Scheduler_algorithm
	switch algoritmo {
	case "FIFO":
		proximoHilo = planificarFIFO()
	case "PRIORIDADES":
		proximoHilo = planificarPrioridades()
	case "CMN":
		proximoHilo = planificarMultinivel()
	default:
		log.Fatalf("Algoritmo de planificación no reconocido: %s", algoritmo)
	}
	return proximoHilo
}

func buscarHiloMayorPrioridad() Tcb {
	mutexColaReady.Lock()
	proximoHilo := colaReady[0]
	mutexColaReady.Unlock()

	indiceProximoHilo := 0

	for i, tcb := range colaReady {
		if tcb.Prioridad < proximoHilo.Prioridad || (tcb.Prioridad == proximoHilo.Prioridad && i < indiceProximoHilo) {
			proximoHilo = tcb
			indiceProximoHilo = i
		}
	}

	return proximoHilo
}

func eventoReplanificacionObligatorio() {
	go planificarCortoPlazo()
}

func eventoReplanificacionOpcional() {
	if tieneDesalojo(KernelConfig.Scheduler_algorithm) {
		go planificarCortoPlazo()
	}
}

func desalojar() {
	agregarAReady(*hiloEnEjecucion)
	hiloEnEjecucion = nil
	evaluarPausaPlanificacion()
}

func tieneDesalojo(algoritmo string) bool {
	return algoritmo != "FIFO" // En el contexto del tp
}

func existePrioridad(prioridad int32) bool {
	if _, existe := colasMultinivel[int(prioridad)]; existe {
		return true
	}
	return false

}

func implicaReplanificar(motivoDevolucion string) bool {
	return motivoDevolucion == "Interrupción"
}

// Simulación
func proximoHiloAEjecutarSimulado() *Tcb {
	var proximoHilo *Tcb
	algoritmo := KernelConfig.Scheduler_algorithm
	switch algoritmo {
	case "FIFO":
		proximoHilo = planificarFIFOSimulado()
	case "PRIORIDADES":
		proximoHilo = planificarPrioridadesSimulado()
	case "CMN":
		proximoHilo = planificarMultinivelSimulado()
	default:
		log.Fatalf("Algoritmo de planificación no reconocido: %s", algoritmo)
	}
	return proximoHilo
}

func planificarFIFOSimulado() *Tcb {
	if len(colaReady) == 0 {
		return nil
	}
	proximoHilo := colaReady[0]
	return &proximoHilo
}

func planificarPrioridadesSimulado() *Tcb {
	proximoHilo := buscarHiloMayorPrioridad()
	return &proximoHilo
}

func planificarMultinivelSimulado() *Tcb {
	chequearQuitarColas()

	if len(colasMultinivel) == 0 {
		return nil
	}

	prioridad := proximoHiloMultinivel()

	hiloParaEjecutar := colasMultinivel[prioridad][0]
	return &hiloParaEjecutar
}

func pausarPlanificacionCortoPlazo() {
	pauseChan <- struct{}{}
	planificadorPausado = true
	log.Println("Pauso la planificación a corto plazo")
	// Esperar a que no haya un hilo en exec
	hiloEjecutando.L.Lock()
	go evaluarNadieEjecuta() // Caso borde donde no hay nadie más ejecutando
	hiloEjecutando.Wait()
	hiloEjecutando.L.Unlock()
}

func reanudarPlanificacionCortoPlazo() {
	planificadorPausado = false
	log.Println("Reanudo la planificación a corto plazo")
	<-pauseChan
	eventoReplanificacionObligatorio()
}

func evaluarPausaPlanificacion() {
	if planificadorPausado {
		hiloEjecutando.L.Lock()
		hiloEjecutando.Signal()
		hiloEjecutando.L.Unlock()
	}
}

func evaluarNadieEjecuta() {
	if hiloEnEjecucion == nil {
		hiloEjecutando.L.Lock()
		hiloEjecutando.Signal()
		hiloEjecutando.L.Unlock()
	}
}
