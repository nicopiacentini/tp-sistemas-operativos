package utils

import (
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"log"
)

func init() {
	utils_general.ConfigurarLogger("memoria.log")
	if MemoryConfig.Scheme == "FIJAS" || MemoryConfig.Scheme == "DINAMICAS" {
		particiones = iniciarParticiones()
	} else {
		log.Fatal("configuracion de memoria no identificada")
	}
}
