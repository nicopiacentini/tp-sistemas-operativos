package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/utils"
	"net/http"
	"strconv"
)

func main() {
	// Obtengo el puerto del kernel
	puerto := strconv.Itoa(utils.KernelConfig.Port)

	http.HandleFunc("/crearProceso", utils.PROCESS_CREATE)
	http.HandleFunc("/finalizarProceso", utils.PROCESS_EXIT)
	http.HandleFunc("/crearHilo", utils.THREAD_CREATE)
	http.HandleFunc("/bloquearHilo", utils.THREAD_JOIN)
	http.HandleFunc("/matarHilo", utils.THREAD_CANCEL)
	http.HandleFunc("/suicidarHilo", utils.THREAD_EXIT)
	http.HandleFunc("/crearMutex", utils.MUTEX_CREATE)
	http.HandleFunc("/bloquearMutex", utils.MUTEX_LOCK)
	http.HandleFunc("/desbloquearMutex", utils.MUTEX_UNLOCK)
	http.HandleFunc("/entradaSalida", utils.IO)
	http.HandleFunc("/dumpearProceso", utils.DUMP_MEMORY)
	http.HandleFunc("/retornoHilo", utils.THREAD_RETURNED)
	http.HandleFunc("/procesoCreado", utils.RespuestaInicializacionProceso)
	http.HandleFunc("/procesoFinalizado", utils.RespuestaFinalizacionProceso)
	http.HandleFunc("/procesoFinalizadoDump", utils.RespuestaFinalizacionProcesoDump)
	http.HandleFunc("/procesoDumpeado", utils.RespuestaDump)
	http.HandleFunc("/compactacionFinalizada", utils.RespuestaCompactacionFinalizada)

	// Inicio el servidor
	err := http.ListenAndServe(":"+puerto, nil)
	if err != nil {
		panic(err)
	}
}
