package utils

import (
	"encoding/binary"
	"fmt"
	"log"
)

// inicializa la list de particiones dependiendo de esquema elegido
func iniciarParticiones() []Particion {
	listaParticiones := MemoryConfig.Partitions //
	if MemoryConfig.Scheme == "FIJAS" {         //trabajando con particiones fijas
		var nuevasParticiones []Particion = make([]Particion, len(listaParticiones))
		var base int = 0
		for i := 0; i < len(listaParticiones); i++ {
			nuevasParticiones[i].base = base                  //la base de la particion
			nuevasParticiones[i].limite = listaParticiones[i] //el limite de la particion
			base += listaParticiones[i]                       //incremento la base general para poder usarle como base de la siguiente particion
			nuevasParticiones[i].pid = -1                     //no existe el pid -1 por eso las inicializo con eso
		}
		return nuevasParticiones
	}
	if MemoryConfig.Scheme == "DINAMICAS" { //trabajando con particiones dinamica
		var nuevasParticiones []Particion
		nuevasParticiones = append(nuevasParticiones, Particion{base: 0, limite: MemoryConfig.Memory_Size, ocupado: false, pid: -1})
		return nuevasParticiones
	}
	log.Fatal("Configuracion de particiones no definida")
	return nil
}

// Función auxiliar para leer los 4 bytes de la memoria de usuario
func buscarEnMemoriaUsuario(direccionFisica uint32) uint32 {
	// Leer los 4 bytes a partir de la dirección física en la memoria de usuario
	if int(direccionFisica)+4 > len(memoria.MemoriaUsuario) {
		fmt.Printf("Error, se esta leyendo fuera de limites en direccion fisica %d", direccionFisica) // Error si se intenta leer fuera de los límites
	}
	return binary.LittleEndian.Uint32(memoria.MemoriaUsuario[direccionFisica : direccionFisica+4])
}
