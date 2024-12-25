package utils

import (
	"fmt"
)

// Función que guarda el pid y tid localmente
func guardarPidTid(pid int32, tid int32) {
	pidTidRecibidos[pid] = tid
}

func guardarContexto(pid int32, tid int32, contexto Contexto) {
	// Definir la clave única para identificar el contexto por PID y TID
	PidTidKey := PidTid{Pid: pid, Tid: tid}

	// Verificar si la entrada ya existe; si no, crear una nueva
	if _, exists := tablaContextos[PidTidKey]; !exists {
		tablaContextos[PidTidKey] = &Contexto{}
	}

	// Almacenar todo el contexto recibido, incluyendo Base y Limite
	tablaContextos[PidTidKey].ContextoRegistros = contexto.ContextoRegistros
	tablaContextos[PidTidKey].Base = contexto.Base
	tablaContextos[PidTidKey].Limite = contexto.Limite
}

func MMU(contexto *Contexto, direccionLogica uint32, tamano uint32, pid int32, tid int32) (uint32, error) {
	base := contexto.Base
	limite := contexto.Limite

	// Validar si la dirección lógica + tamaño está dentro de los límites
	if direccionLogica >= limite {
		// Manejar el fallo de segmentación
		handleSegmentationFault(contexto, pid, tid)
		return 0, fmt.Errorf("segmentation fault: dirección lógica %d fuera de límites para tamaño %d", direccionLogica, tamano)
	}

	// Calcular y devolver la dirección física
	direccionFisica := base + direccionLogica
	return direccionFisica, nil
}

// obtenerPCB busca el PCB en la tabla de procesos
func obtenerContexto(pid int32, tid int32) (*Contexto, error) {
	// Verificar si el PCB existe en la tabla
	pcb, existe := tablaContextos[PidTid{Pid: pid, Tid: tid}]
	if !existe {
		return nil, fmt.Errorf("Contexto no encontrado para PID: %d, Tid: %d", pid, tid)
	}
	return pcb, nil
}

// esSyscall verifica si la instrucción es una syscall
func esSyscall(instruccion string) bool {
	syscalls := []string{"PROCESS_CREATE", "PROCESS_EXIT", "THREAD_CREATE", "THREAD_JOIN", "THREAD_CANCEL", "THREAD_EXIT", "MUTEX_CREATE", "MUTEX_LOCK", "MUTEX_UNLOCK", "DUMP_MEMORY", "IO"}
	for _, syscall := range syscalls {
		if instruccion == syscall {
			return true
		}
	}
	return false
}

func tieneOperandos(instruccion []string) bool {
	return len(instruccion) > 1
}
