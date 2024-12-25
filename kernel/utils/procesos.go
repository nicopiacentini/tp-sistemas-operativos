package utils

import (
	"github.com/sisoputnfrba/tp-golang/utils_general"
)

// FUNCIONES RELATIVAS A PROCESOS //

func crearProceso(pathArchivo string, tamañoArchivo int, prioridadTid0 int32) {

	pcb := crearPcb() // Crear el PCB

	mutexParamsInitMemoria.Lock()
	paramsInitMemoria[pcb.Pid] = ParamsInitMemoria{
		Archivo:       pathArchivo,
		Tamaño:        tamañoArchivo,
		PrioridadTid0: prioridadTid0,
	}
	mutexParamsInitMemoria.Unlock()

	agregarANew(pcb) // Agregar el PCB a la cola NEW

	utils_general.LoggearMensaje(KernelConfig.Log_level, "## (<%d>:0) Se crea el proceso - Estado: NEW", pcb.Pid) // LOG OBLIGATORIO

	planificarLargoPlazo("PROCESS_CREATE", Params{}) // Mandar a planificar por creación de proceso
}

func crearPcb() Pcb {
	// Inicializar el struct pcb con el pid global(para el proceso inicial será 0 y luego irá incrementando)
	pcb := Pcb{
		Pid:     global_pid,
		Tids:    []int32{},
		Mutexes: []*Mutex{},
	}

	// Incrementar global_pid
	mutexGlobalPid.Lock()
	global_pid++
	mutexGlobalPid.Unlock()

	// Agregar el PCB al array global
	mutexPcbs.Lock()
	pcbs = append(pcbs, pcb)
	mutexPcbs.Unlock()

	return pcb
}

func agregarANew(pcb Pcb) {
	mutexColaNew.Lock()    // Bloquear la cola NEW
	encolar(pcb, &colaNew) // Mando proceso a New
	mutexColaNew.Unlock()  // Desbloquear la cola NEW
}

func quitarDeNew() Pcb {
	mutexColaNew.Lock()   // Bloquear la cola NEW
	pcb := pop(&colaNew)  // Quito el proceso de NEW
	mutexColaNew.Unlock() // Desbloquear la cola NEW
	return pcb            // Devuelvo el proceso
}

func liberarPcb(pid int32) {
	mutexPcbs.Lock()
	defer mutexPcbs.Unlock()

	for i, pcb := range pcbs {
		if pcb.Pid == pid {
			// Eliminar el PCB del array
			pcbs = append(pcbs[:i], pcbs[i+1:]...)
			break
		}
	}
}
