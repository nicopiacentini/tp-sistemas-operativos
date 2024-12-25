package main

import (
	"github.com/sisoputnfrba/tp-golang/memoria/utils"
	"net/http"
	"strconv"
)

func main() {
	puerto := strconv.Itoa(utils.MemoryConfig.Port)

	http.HandleFunc("/iniciarProceso", utils.CrearProceso)              
	http.HandleFunc("/iniciarHilo", utils.IniciarHilo)                    
	http.HandleFunc("/finalizarProceso", utils.FinalizarProceso)     
	http.HandleFunc("/finalizarProcesoDump", utils.FinalizarProcesoDump)
	http.HandleFunc("/finalizarHilo", utils.FinalizarHilo)                
	http.HandleFunc("/dumpearProceso", utils.DumpearProceso)              
	http.HandleFunc("/contextoEjecucion", utils.RecibirSolicitudContexto) 
	http.HandleFunc("/actualizarContexto", utils.RecibirSolicitudActualizacionContexto)
	http.HandleFunc("/instruccion", utils.RecibirSolicitudFetch)
	http.HandleFunc("/read", utils.RecibirSolicitudLecturaMemoria)
	http.HandleFunc("/write", utils.RecibirSolicitudEscrituraMemoria)
	http.HandleFunc("/compactacion",utils.Compactar)
	http.HandleFunc("/respuestaDump", utils.RespuestaFileSystem)

	err := http.ListenAndServe(":"+puerto, nil)
	if err != nil {
		panic(err)
	}
}
