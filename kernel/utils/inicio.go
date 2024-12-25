package utils

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"os"
	"strconv"
)

// Inicio de Kernel //
func init() {
	// Abro el logger
	utils_general.ConfigurarLogger("kernel.log")

	if len(os.Args) < 2 { // si no se pasan los argumentos necesarios, se termina el programa
		os.Exit(1)
	}

	archivo := os.Args[1]
	tamaño, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("Error al convertir el tamaño: %v\n", err)
		os.Exit(1)
	}

	crearProceso(archivo, tamaño, 0) //al iniciar se crea un proceso inicial con un archivo de pseudocódigo, un tamaño y prioridad del tid 0 = 0
}
