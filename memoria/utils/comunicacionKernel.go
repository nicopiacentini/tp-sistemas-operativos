package utils

import (
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"log"
	"net/http"
)

func CrearProceso(w http.ResponseWriter, r *http.Request) {
	var nuevoProceso ProcessRequestMemory
	utils_general.HandlePostRequest(w, r, &nuevoProceso)
	w.WriteHeader(http.StatusOK)
	// agregar a memoria de sistema el nuevo proceso

	PID := nuevoProceso.Pid //guardo el pid
	muAsignarParticion.Lock()
	//base, limite := asignarParticion(nuevoProceso.Tamaño)
	particion := asignarParticion(nuevoProceso.Tamaño) //busco un particion libre para asignarle al pid y devuelvo el index de la particion
	// si no encuentro nada devuelvo -1
	muAsignarParticion.Unlock()
	//ASIGNAR LA PARTICION EN MEMORIA EN LA VIDA REAL
	if particion == -1 {

		if MemoryConfig.Scheme == "FIJAS" {
			responderAKernel(PID, 2, "procesoCreado")
			return
		}
		if MemoryConfig.Scheme == "DINAMICAS" {
			//AVISAR QUE NO TENGO ESPACIO CONTIGUO

			//EVALUAR COMPACTACION
			if puedeCompactar(nuevoProceso.Tamaño) {
				// avisar que no hay espacio contiguo pero si en huecos
				responderAKernel(PID, 1, "procesoCreado")
				// si alcanza el espaciolibre acumulandolo todo -> compacto
				return
			} else {
				responderAKernel(PID, 2, "procesoCreado")
				// sino le digo al kernel que no se pudo y retorno
				return
			}

		}
	}
	inicializarProcesoEnMemoria(PID, particion)
	responderAKernel(PID, 0, "procesoCreado")
	utils_general.LoggearMensaje(MemoryConfig.Log_level, "## Proceso <Creado> - PID: <%d> - Tamaño: <%d>", PID, nuevoProceso.Tamaño) // LOG OBLIGATORIO
	mostrarParticiones(particiones)
}

func IniciarHilo(w http.ResponseWriter, r *http.Request) {
	var nuevoHilo ThreadRequestMemory
	utils_general.HandlePostRequest(w, r, &nuevoHilo)
	inicializarHilo(nuevoHilo.Pid, nuevoHilo.Tid, nuevoHilo.Pathseudo)
	w.WriteHeader(http.StatusOK)
	utils_general.LoggearMensaje(MemoryConfig.Log_level, "## Hilo <Creado> - (PID:TID) - (<%d>:<%d>)", nuevoHilo.Pid, nuevoHilo.Tid) // LOG OBLIGATORIO
}

func FinalizarProceso(w http.ResponseWriter, r *http.Request) {
	// modificar particiones para liberar espacio que ocupaba ese proceso
	// liberar espacio de memoria asociada y mergear el espacio libre
	var processExit ProcessExitMemory
	utils_general.HandlePostRequest(w, r, &processExit)
	w.WriteHeader(http.StatusOK)

	tamaño := tamañoDeProcesoAFinalizar(processExit.Pid) //revisar

	liberarParticion(processExit.Pid)

	quitarPidDeMemoriaDeSistema(processExit.Pid)

	utils_general.LoggearMensaje(MemoryConfig.Log_level, "## Proceso <Destruido> - PID: <%d> - Tamaño: <%d>", processExit.Pid, tamaño) // LOG OBLIGATORIO

	responderAKernel(processExit.Pid, 0, "procesoFinalizado")
}

func FinalizarProcesoDump(w http.ResponseWriter, r *http.Request) {
	// modificar particiones para liberar espacio que ocupaba ese proceso
	// liberar espacio de memoria asociada y mergear el espacio libre
	var processExit ProcessExitMemory
	utils_general.HandlePostRequest(w, r, &processExit)
	w.WriteHeader(http.StatusOK)

	tamaño := tamañoDeProcesoAFinalizar(processExit.Pid) //revisar

	liberarParticion(processExit.Pid)

	quitarPidDeMemoriaDeSistema(processExit.Pid)

	utils_general.LoggearMensaje(MemoryConfig.Log_level, "## Proceso <Destruido> - PID: <%d> - Tamaño: <%d>", processExit.Pid, tamaño) // LOG OBLIGATORIO

	responderAKernel(processExit.Pid, 0, "procesoFinalizadoDump")
}

func tamañoDeProcesoAFinalizar(pid int32) int {
	proceso, existe := memoria.MemoriaSistema[pid]
	if !existe {
		return 0
	}
	return int(proceso.Limite)
}

func quitarPidDeMemoriaDeSistema(PID int32) {
	// Eliminar el proceso completo de la memoria del sistema
	if len(memoria.MemoriaSistema[PID].tidsDePid) != 0 {
		for TID := range memoria.MemoriaSistema[PID].tidsDePid {
			utils_general.LoggearMensaje(MemoryConfig.Log_level, "## Hilo <Destruido> - (PID:TID) - (<%d>:<%d>)", PID, TID) // LOG OBLIGATORIO
		}
	}
	delete(memoria.MemoriaSistema, PID)
}

func FinalizarHilo(w http.ResponseWriter, r *http.Request) {
	var threadExit ThreadExitMemory
	utils_general.HandlePostRequest(w, r, &threadExit)
	w.WriteHeader(http.StatusOK)
	quitarHiloDeMemoriaSistema(threadExit.Pid, threadExit.Tid)
	utils_general.LoggearMensaje(MemoryConfig.Log_level, "## Hilo <Destruido> - (PID:TID) - (<%d>:<%d>)", threadExit.Pid, threadExit.Tid) // LOG OBLIGATORIO
}

// borra el hilo de la memoria de sistema
func quitarHiloDeMemoriaSistema(PID int32, TID int32) {
	delete(memoria.MemoriaSistema[PID].tidsDePid, TID)
}

func responderAKernel(PID int32, respuesta int, endpoint string) {
	var respuestaKernel ResponseMemory
	respuestaKernel.Pid = PID
	respuestaKernel.Codigo = respuesta
	utils_general.PostRequest(respuestaKernel, MemoryConfig.Ip_Kernel, MemoryConfig.Port_Kernel, endpoint)
}

func inicializarProcesoEnMemoria(PID int32, Particion int) { //inicializa el proceso en memoria
	var contextoInicial contextoDePid
	contextoInicial.Estado = "NEW"
	contextoInicial.Base = uint32(particiones[Particion].base)
	contextoInicial.Limite = uint32(particiones[Particion].limite)
	contextoInicial.tidsDePid = make(map[int32]Contexto)
	memoria.MemoriaSistema[PID] = contextoInicial
	particiones[Particion].pid = PID
	particiones[Particion].ocupado = true
}

// inicializa el hilo en memroia
func inicializarHilo(pid int32, tid int32, nombreArchivo string) {
	registrosInit := Registros{AX: 0, BX: 0, CX: 0, DX: 0, EX: 0, FX: 0, GX: 0, HX: 0, PC: 0}
	instruccionesDecodificadas, err := Decode(nombreArchivo)
	if err != nil {
		log.Fatalf("No se pudo leer el archivo. ERROR: %d", err)
	}
	memoria.MemoriaSistema[pid].tidsDePid[tid] = Contexto{ContextoRegistros: registrosInit, Estado: "NEW", Instrucciones: instruccionesDecodificadas}
}
