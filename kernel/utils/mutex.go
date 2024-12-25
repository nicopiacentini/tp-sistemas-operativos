package utils

import (
	"github.com/sisoputnfrba/tp-golang/utils_general"
)

// FUNCIONES RELATIVAS A MUTEX //

func crearMutex(pid int32, recurso string) {
	for i := range pcbs {
		if pcbs[i].Pid == pid {
			mutexPcbs.Lock()
			pcbs[i].Mutexes = append(pcbs[i].Mutexes, &Mutex{Recurso: recurso, Asignado: false, Dueño: nil, HilosEnEspera: []*Tcb{}})
			mutexPcbs.Unlock()
			utils_general.LoggearMensaje(KernelConfig.Log_level, "Mutex %s añadido al proceso con PID %d", recurso, pid)
			return
		}
	}

	utils_general.LoggearMensaje(KernelConfig.Log_level, "Proceso con PID %d no encontrado", pid)
}

func existeMutex(pid int32, recurso string) bool {
	for _, pcb := range pcbs {
		if pcb.Pid == pid {
			for _, mutex := range pcb.Mutexes {
				if mutex.Recurso == recurso {
					return true
				}
			}
		}
	}
	return false
}

func mutexDisponible(pid int32, recurso string) bool {
	for _, pcb := range pcbs {
		if pcb.Pid == pid {
			for _, mutex := range pcb.Mutexes {
				if mutex.Recurso == recurso {
					return !mutex.Asignado
				}
			}
		}
	}
	return false
}

func asignarMutex(tcbHilo Tcb, recurso string) {
	for _, pcb := range pcbs {
		if pcb.Pid == tcbHilo.Pid {
			for _, mutex := range pcb.Mutexes {
				if mutex.Recurso == recurso {
					mutexPcbs.Lock()
					mutex.Asignado = true
					mutex.Dueño = &tcbHilo
					mutexPcbs.Unlock()
					utils_general.LoggearMensaje(KernelConfig.Log_level, "Mutex %s asignado al hilo con PID %d y TID %d", recurso, tcbHilo.Pid, tcbHilo.Tid)
				}
			}
		}
	}
}

func bloquearPorMutex(tcb Tcb, recurso string) {
	var mutex *Mutex

	// Buscar el PCB correspondiente y el mutex
	for i := range pcbs {
		if pcbs[i].Pid == tcb.Pid {
			for j := range pcbs[i].Mutexes {
				if pcbs[i].Mutexes[j].Recurso == recurso {
					mutexPcbs.Lock()
					mutex = pcbs[i].Mutexes[j]
					mutexPcbs.Unlock()

					mutexHilosEnEspera.Lock()
					mutex.HilosEnEspera = append(mutex.HilosEnEspera, &tcb) // Agregar el TCB a la lista de HilosEnEspera del mutex
					mutexHilosEnEspera.Unlock()
					break
				}
			}
			break
		}
	}

	bloquear(tcb, "MUTEX") // En definitiva, bloquear el hilo
}

func liberarMutex(pid int32, recurso string) {
	for i := range pcbs {
		if pcbs[i].Pid == pid {
			for j := range pcbs[i].Mutexes {
				if pcbs[i].Mutexes[j].Recurso == recurso {
					if len(pcbs[i].Mutexes[j].HilosEnEspera) > 0 {
						mutexPcbs.Lock()
						mutex := pcbs[i].Mutexes[j]
						mutexPcbs.Unlock()

						mutexHilosEnEspera.Lock()
						siguienteHilo := pop(&mutex.HilosEnEspera)
						mutexHilosEnEspera.Unlock()

						asignarMutex(*siguienteHilo, recurso)
						utils_general.LoggearMensaje(KernelConfig.Log_level, "Mutex %s liberado por proceso con PID %d, asignado a hilo con PID %d y TID %d", recurso, pid, siguienteHilo.Pid, siguienteHilo.Tid)
						desbloquear(*siguienteHilo)
						eventoReplanificacionOpcional()
						utils_general.LoggearMensaje(KernelConfig.Log_level, "## (<%d>:<%d>) Desbloqueado por: <MUTEX_UNLOCK>", siguienteHilo.Pid, siguienteHilo.Tid)
					} else {
						mutexPcbs.Lock()
						pcbs[i].Mutexes[j].Asignado = false
						pcbs[i].Mutexes[j].Dueño = nil
						mutexPcbs.Unlock()
						utils_general.LoggearMensaje(KernelConfig.Log_level, "Mutex %s liberado por proceso con PID %d", recurso, pid)
						mandarACpuParaEjecutar(hiloEnEjecucion.Pid, hiloEnEjecucion.Tid)
					}
					return
				}
			}
		}
	}
}

func liberarTodosLosMutex(pid int32, tid int32) {
	for i := range pcbs {
		if pcbs[i].Pid == pid {
			for j := range pcbs[i].Mutexes {
				if pcbs[i].Mutexes[j].Dueño == nil {
					// EL mutex ya fue liberado por mutex unlock previamente
					return
				} else if pcbs[i].Mutexes[j].Dueño.Pid == pid && pcbs[i].Mutexes[j].Dueño.Tid == tid {
					liberarMutex(pid, pcbs[i].Mutexes[j].Recurso)
				}
			}
			return
		}
	}
}

func esDueño(tcbHilo Tcb, recurso string) bool {
	for i := range pcbs {
		if pcbs[i].Pid == tcbHilo.Pid {
			for j := range pcbs[i].Mutexes {
				if pcbs[i].Mutexes[j].Recurso == recurso && pcbs[i].Mutexes[j].Dueño.Pid == tcbHilo.Pid && pcbs[i].Mutexes[j].Dueño.Tid == tcbHilo.Tid {
					return true
				}
			}
		}
	}
	return false
}
