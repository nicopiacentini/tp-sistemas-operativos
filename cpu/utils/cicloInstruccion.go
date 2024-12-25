package utils

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"log"
	"strconv"
	"strings"
)

// Función que busca la próxima instrucción a ejecutar y actualiza el PC para hilos (TID)
func Fetch(pid int32, tid int32, pc uint32) {
	solicitudInstruccion := SolicitudFetch{
		Pid: pid,
		Tid: tid,
		Pc:  pc,
	}
	utils_general.LoggearMensaje(CpuConfig.Log_level, "## TID: <%d> - FETCH - Program Counter: <%d>", tid, pc) // LOG OBLIGATORIO

	utils_general.PostRequest(solicitudInstruccion, CpuConfig.Ip_memory, CpuConfig.Port_memory, "instruccion")
}

// Execute ejecuta la instrucción dada por un TID.
func Execute(pid int32, tid int32, instruccion string) error {
	contexto, err := obtenerContexto(pid, tid)
	if err != nil {
		return fmt.Errorf("error al obtener el contexto: %v", err)
	}

	pcInicial := contexto.ContextoRegistros.PC

	// Dividir la instrucción en función y parámetros
	partes := strings.Fields(instruccion) // Esto divide la instrucción en partes separadas por espacios
	if tieneOperandos(partes) {
		operandos := strings.Join(partes[1:], " ")
		utils_general.LoggearMensaje(CpuConfig.Log_level, "## TID: <%d> - Ejecutando: <%s> - <%s>", tid, partes[0], operandos) // LOG OBLIGATORIO
	} else {
		utils_general.LoggearMensaje(CpuConfig.Log_level, "## TID: <%d> - Ejecutando: <%s>", tid, partes[0]) // LOG OBLIGATORIO
	}

	if esSyscall(partes[0]) {
		contexto.ContextoRegistros.PC++ // Provisorio
		err := actualizarContextoEjecucion(pid, tid, *contexto)
		if err != nil {
			return fmt.Errorf("error al actualizar el contexto de ejecución en memoria: %v", err)
		}
	}

	// Procesar la instrucción según el comando
	switch partes[0] {
	case "SET":
		if esRegistro(partes[2]) {
			valorRegistro2 := getValorRegistro(contexto, partes[2])
			setRegistro(contexto, partes[1], valorRegistro2)
		} else {
			registro := partes[1]
			valor, err := strconv.Atoi(partes[2])
			if err != nil {
				return fmt.Errorf("error al convertir valor a entero en SET: %v", err)
			}
			setRegistro(contexto, registro, uint32(valor))
		}

	case "READ_MEM", "MEM_READ":
		registroDatos := partes[1]
		registroDireccion := partes[2]
		readMem(contexto, registroDatos, registroDireccion, pid, tid)

	case "WRITE_MEM", "MEM_WRITE":
		registroDireccion := partes[1]
		registroDatos := partes[2]
		writeMem(contexto, registroDireccion, registroDatos, pid, tid)

	case "SUM":
		registroDestino := partes[1]
		registroOrigen := partes[2]
		sumRegistros(contexto, registroDestino, registroOrigen)

	case "SUB":
		registroDestino := partes[1]
		registroOrigen := partes[2]
		subRegistros(contexto, registroDestino, registroOrigen)

	case "JNZ":
		registro := partes[1]
		nuevaInstruccion, _ := strconv.Atoi(partes[2])
		jnz(contexto, registro, uint32(nuevaInstruccion))

	case "LOG":
		registro := partes[1]
		logRegistro(contexto, registro)

	// syscalls
	case "DUMP_MEMORY":
		DUMP_MEMORY()
		return nil

	case "IO":
		tiempo, _ := strconv.Atoi(partes[1])
		IO(tiempo)
		return nil

	case "PROCESS_CREATE":
		archivo := partes[1]
		tamaño, _ := strconv.Atoi(partes[2])
		prioridad, _ := strconv.Atoi(partes[3])
		PROCESS_CREATE(archivo, tamaño, prioridad)
		return nil

	case "THREAD_CREATE":
		archivo := partes[1]
		prioridad, _ := strconv.Atoi(partes[2])
		THREAD_CREATE(archivo, prioridad)
		return nil

	case "THREAD_JOIN":
		tidJoin, _ := strconv.Atoi(partes[1])
		THREAD_JOIN(tidJoin)
		return nil

	case "THREAD_CANCEL":
		tidCancel, _ := strconv.Atoi(partes[1])
		THREAD_CANCEL(tidCancel)
		return nil

	case "MUTEX_CREATE":
		recurso := partes[1]
		MUTEX_CREATE(recurso)
		return nil

	case "MUTEX_LOCK":
		recurso := partes[1]
		MUTEX_LOCK(recurso)
		return nil

	case "MUTEX_UNLOCK":
		recurso := partes[1]
		MUTEX_UNLOCK(recurso)
		return nil

	case "THREAD_EXIT":
		THREAD_EXIT()
		return nil

	case "PROCESS_EXIT":
		PROCESS_EXIT()
		return nil

	default:
		return fmt.Errorf("instrucción no reconocida: %s", instruccion)
	}

	pcFinal := contexto.ContextoRegistros.PC

	if pcFinal == pcInicial {
		contexto.ContextoRegistros.PC++
	}
	// Paso 4: Check Interrupt - Verificar si hay interrupciones
	if CheckInterrupt(pid, tid) {
		err = actualizarContextoEjecucion(pid, tid, *contexto)
		if err != nil {
			return fmt.Errorf("error al enviar contexto a memoria: %v", err)
		}

		// Notificar al Kernel sobre la interrupción
		notificarInterrupcionAlKernel(tid, "Interrupción")

		return nil // Salir del ciclo si se maneja la interrupción
	}

	Fetch(pid, tid, contexto.ContextoRegistros.PC)
	return nil
}

// CheckInterrupt verifica si hay interrupciones para el TID actual
func CheckInterrupt(pid int32, tid int32) bool {
	PidTidKey := PidTid{Pid: pid, Tid: tid}
	// Verificar si hay una interrupción para el TID dado en el mapa
	log.Printf("Chequeando interrupciones para PID: %d, TID: %d", pid, tid)
	muInterrupt.Lock()
	existe := interrupciones[PidTidKey]
	muInterrupt.Unlock()

	if !existe { // No hay interrupción para este TID
		log.Printf("Interrupción no detectada para PID: %d ,TID %d", pid, tid)
		//limpiarInterrupciones()
		return false
	} else {
		log.Printf("Interrupción detectada para PID: %d ,TID %d", pid, tid)

		// Eliminar la interrupción del mapa una vez procesada
		interrupciones[PidTidKey] = false

		// Interrupción manejada exitosamente
		return true
	}
}

func handleSegmentationFault(contexto *Contexto, pid int32, tid int32) {
	log.Printf("Segmentation fault detectado para TID %d", tid)
	// Notificar al Kernel sobre el fallo
	notificarInterrupcionAlKernel(tid, "Segmentation Fault")
	// Enviar el contexto actual a memoria
	err := actualizarContextoEjecucion(pid, tid, *contexto)
	if err != nil {
		log.Printf("Error al enviar el contexto a memoria: %v", err)
		return
	}
}

/*
func limpiarInterrupciones() {
	for i := range interrupciones {
		interrupciones[i] = false
	}
}
*/
