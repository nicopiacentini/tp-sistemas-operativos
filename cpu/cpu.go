package main

import (
	"github.com/sisoputnfrba/tp-golang/cpu/utils"
	"net/http"
	"strconv"
)

func main() {
	// Obtengo el puerto del cpu
	puerto := strconv.Itoa(utils.CpuConfig.Port)

	http.HandleFunc("/ejecutarHilo", utils.RecibirSolicitudKernel)      // Recibo PID y TID a ejecutar
	http.HandleFunc("/recibirContexto", utils.RecibirContextoEjecucion) // Recibo contexto de ejecucion
	http.HandleFunc("/recibirInstruccion", utils.RecibirInstruccion)
	http.HandleFunc("/interrupcion", utils.ObtenerInterrupcionDelKernel)

	// Inicio el servidor
	err := http.ListenAndServe(":"+puerto, nil)
	if err != nil {
		panic(err)
	}
}
