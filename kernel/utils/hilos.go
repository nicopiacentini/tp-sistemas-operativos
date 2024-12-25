package utils

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"sync"
	"time"
)

// FUNCIONES RELATIVAS A HILOS //

func crearTcb(prioridad int32, pid int32) Tcb {
	var pcb *Pcb

	mutexPcbs.Lock()
	for i := range pcbs {
		if pcbs[i].Pid == pid {
			pcb = &pcbs[i]
			break
		}
	}
	mutexPcbs.Unlock()

	if pcb == nil {
		panic(fmt.Sprintf("PCB no encontrado para PID %d", pid))
	}

	tid := int32(len(pcb.Tids))      // Utilizar el tamaño de la lista de TIDs para asignar el TID
	pcb.Tids = append(pcb.Tids, tid) // Agregar el TID a la lista de TIDs del PCB

	tcb := Tcb{
		Tid:        tid,
		Prioridad:  prioridad,
		Pid:        pid,
		Bloqueados: []*Tcb{},
	}

	return tcb
}

func finalizarHilosDeProceso(pid int32) {
	// Buscar y eliminar TIDs en colaReady
	if KernelConfig.Scheduler_algorithm == "CMN" {
		mutexColasMultinivel.Lock()
		buscarYEliminarTcbsDeColasMultinivel(pid)
		mutexColasMultinivel.Unlock()
	} else {
		buscarYEliminarTcbsDeCola(&colaReady, pid, &mutexColaReady)
	}
	// Buscar y eliminar TIDs en colaBlocked
	tcbs := buscarYEliminarTcbsDeCola(&colaBlocked, pid, &mutexColaBlocked)

	for _, tcb := range tcbs {
		moverTcbAExit(tcb)
	}
}

func buscarYEliminarTcbsDeColasMultinivel(pid int32) {
	for index, cola := range colasMultinivel {

		for j, tcb := range cola {
			if tcb.Pid == pid {
				colasMultinivel[index] = append((cola)[:j], (cola)[j+1:]...)

			}
		}
	}
	chequearQuitarColas()
}

func buscarYEliminarTcbsDeCola(cola *[]Tcb, pid int32, mutex *sync.Mutex) []Tcb {
	mutex.Lock()
	defer mutex.Unlock()

	var tcbs []Tcb
	var indices []int

	// Buscar los TCBs y sus índices
	for i, tcb := range *cola {
		if tcb.Pid == pid {
			tcbs = append(tcbs, tcb)
			indices = append(indices, i)
		}
	}

	// Eliminar los TCBs de la cola
	for i := len(indices) - 1; i >= 0; i-- {
		*cola = append((*cola)[:indices[i]], (*cola)[indices[i]+1:]...)
	}

	return tcbs
}

func moverTcbAExit(tcb Tcb) {
	quitarDeDondeEsta(tcb)

	mutexColaExit.Lock()
	colaExit = append(colaExit, tcb)
	mutexColaExit.Unlock()
}

func quitarDeDondeEsta(tcb Tcb) {
	if KernelConfig.Scheduler_algorithm == "CMN" {
		buscarYEliminarTcbDeColasMultinivel(tcb)
	} else {
		quitarDeCola(&colaReady, tcb, &mutexColaReady)
	}
	quitarDeCola(&colaBlocked, tcb, &mutexColaBlocked)
}

func buscarYEliminarTcbDeColasMultinivel(tcb Tcb) {
	mutexColasMultinivel.Lock()
	defer mutexColasMultinivel.Unlock() // Liberamos el mutex al final de la función

	for nivelIndex, Tcbs := range colasMultinivel {
		for otroIndex, Tcb := range Tcbs {
			if Tcb.Tid == tcb.Tid && Tcb.Pid == tcb.Pid {
				// Modificamos directamente el slice en colasMultinivel
				colasMultinivel[nivelIndex] = append(Tcbs[:otroIndex], Tcbs[otroIndex+1:]...)
				chequearQuitarColas()
				return
			}
		}
	}
}

func quitarDeCola(cola *[]Tcb, tcb Tcb, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()

	for i := 0; i < len(*cola); i++ {
		if (*cola)[i].Tid == tcb.Tid && (*cola)[i].Pid == tcb.Pid {
			// Eliminar el TCB de la cola
			*cola = append((*cola)[:i], (*cola)[i+1:]...)
			return
		}
	}
}

func bloquear(tcb Tcb, motivo string) {
	mutexColaBlocked.Lock()
	colaBlocked = append(colaBlocked, tcb)
	mutexColaBlocked.Unlock()

	mutexHiloEnEjecucion.Lock()
	hiloEnEjecucion = nil
	mutexHiloEnEjecucion.Unlock()

	utils_general.LoggearMensaje(KernelConfig.Log_level, "## (<%d>:<%d>) - Bloqueado por: <%s>", tcb.Pid, tcb.Tid, motivo) // LOG OBLIGATORIO

	evaluarPausaPlanificacion()
	evaluarInterrupcionReplanificacion(eventoReplanificacionObligatorio) // Siempre replanificar al bloquear un hilo
}

func desbloquear(tcb Tcb) bool {
	mutexColaBlocked.Lock()
	defer mutexColaBlocked.Unlock()

	for i := 0; i < len(colaBlocked); i++ {
		if colaBlocked[i].Pid == tcb.Pid && colaBlocked[i].Tid == tcb.Tid {
			// Mover el TCB a la cola ready
			agregarAReady(colaBlocked[i])

			// Replanificar obligatoriamente si no hay hilo en ejecución
			if hiloEnEjecucion == nil {
				go evaluarInterrupcionReplanificacion(eventoReplanificacionObligatorio)
			} else if tcb.Prioridad > hiloEnEjecucion.Prioridad && tieneDesalojo(KernelConfig.Scheduler_algorithm) {
				go enviarInterrupcionACpu(hiloEnEjecucion.Pid, hiloEnEjecucion.Tid)
			}

			// Eliminar el TCB de la cola blocked
			colaBlocked = append(colaBlocked[:i], colaBlocked[i+1:]...)
			utils_general.LoggearMensaje(KernelConfig.Log_level, fmt.Sprintf("Se desbloquea PID: %d, TID: %d", tcb.Pid, tcb.Tid))
			return true
		}
	}
	return false
}

func añadirABloqueadosPor(pid int32, tid int32, tcbBloqueado Tcb) {
	tcbHilo := buscarTcb(pid, tid)
	tcbHilo.Bloqueados = append(tcbHilo.Bloqueados, &tcbBloqueado)
}

func desbloquearHilosBloqueados(pid int32, tid int32) *Tcb { // Desbloquea a aquellos hilos que se joinearon a este
	tcb := buscarTcb(pid, tid)
	for len(tcb.Bloqueados) > 0 {
		tcbBloqueado := pop(&tcb.Bloqueados)
		desbloquear(*tcbBloqueado)
	}

	return tcb
}

func buscarTcb(pid int32, tid int32) *Tcb {
	// Buscar en la cola de ready
	if KernelConfig.Scheduler_algorithm == "CMN" {
		for index, cola := range colasMultinivel {
			for otroIndex := range cola {
				if cola[otroIndex].Pid == pid && cola[otroIndex].Tid == tid {
					return &colasMultinivel[index][otroIndex]
				}
			}
		}
	} else {
		mutexColaReady.Lock()
		for i := 0; i < len(colaReady); i++ {
			if colaReady[i].Pid == pid && colaReady[i].Tid == tid {
				mutexColaReady.Unlock()
				return &colaReady[i]
			}
		}
		mutexColaReady.Unlock()
	}
	// Buscar en la cola de blocked
	mutexColaBlocked.Lock()
	for i := 0; i < len(colaBlocked); i++ {
		if colaBlocked[i].Pid == pid && colaBlocked[i].Tid == tid {
			mutexColaBlocked.Unlock()
			return &colaBlocked[i]
		}
	}
	mutexColaBlocked.Unlock()

	// Buscar en el hilo en ejecución
	mutexHiloEnEjecucion.Lock()
	if hiloEnEjecucion.Pid == pid && hiloEnEjecucion.Tid == tid {
		mutexHiloEnEjecucion.Unlock()
		return hiloEnEjecucion
	}
	mutexHiloEnEjecucion.Unlock()

	// Si no se encuentra, retornar nil
	return nil
}

func agregarAReady(tcb Tcb) {
	if KernelConfig.Scheduler_algorithm == "CMN" {
		mutexColasMultinivel.Lock()
		agregarAColasMultinivel(tcb)
		mutexColasMultinivel.Unlock()
	} else {
		mutexColaReady.Lock()
		encolar(tcb, &colaReady)
		mutexColaReady.Unlock()
	}
}

func agregarAColasMultinivel(tcb Tcb) {
	var indexPrioridad int = int(tcb.Prioridad)
	if len(colasMultinivel) == 0 {
		insertarHiloMultinivel(tcb)
	} else {
		if existePrioridad(tcb.Prioridad) {
			colasMultinivel[indexPrioridad] = append(colasMultinivel[indexPrioridad], tcb)
		} else {
			insertarHiloMultinivel(tcb)
		}
	}
}

func agregarAlPrincipioDeColasMultinivel(tcb Tcb) {
	prioridad := int(tcb.Prioridad)
	if _, existe := colasMultinivel[prioridad]; !existe {
		colasMultinivel[prioridad] = []Tcb{}
	}

	// Insertar el hilo al principio de la cola
	colasMultinivel[prioridad] = append([]Tcb{tcb}, colasMultinivel[prioridad]...)
}

func finalizo(pid int32, tid int32) bool {
	for _, tcb := range colaExit {
		if tcb.Pid == pid && tcb.Tid == tid {
			return true
		}
	}
	return false
}

func existeHilo(pid int32, tid int32) bool {
	mutexPcbs.Lock()
	defer mutexPcbs.Unlock()

	for _, pcb := range pcbs {
		if pcb.Pid == pid && pertenece(tid, pcb.Tids) {
			return true
		}
	}
	return false
}

func ejecutarIO(tiempo int, pid int32, tid int32) {
	canalIO <- struct{}{} //Encolar hilo en IO
	utils_general.LoggearMensaje(KernelConfig.Log_level, "(PID-TID) - (%d-%d): Comienza IO de %d ms", pid, tid, tiempo)
	mutexIO.Lock()
	time.Sleep(time.Duration(tiempo) * time.Millisecond)
	mutexIO.Unlock()
	utils_general.LoggearMensaje(KernelConfig.Log_level, "## (<%d>:<%d>) finalizó IO y pasa a READY", pid, tid) // LOG OBLIGATORIO
	<-canalIO                                                                                                   //Desencolar hilo en IO
}

func hiloBloqueado(pid int32, tid int32) bool {
	mutexColaBlocked.Lock()
	defer mutexColaBlocked.Unlock()

	for _, tcb := range colaBlocked {
		if tcb.Pid == pid && tcb.Tid == tid {
			return true
		}
	}
	return false
}
