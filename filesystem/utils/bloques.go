package utils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"os"
	"time"
)

func hayEspacioDisponible(tamaño int) bool {
	cantidadBloquesNecesarios := 1 + cantidadBloquesDatosNecesarios(tamaño)
	return cantidadBloquesNecesarios <= cantidadBloquesLibres()
}

func cantidadBloquesDatosNecesarios(tamaño int) int {
	return dividirYRedondearSuperiormente(tamaño, FilesystemConfig.Block_size)
}

func cantidadBloquesLibres() int {
	contador := 0
	mutexBitmap.Lock()
	for _, bit := range bitmap {
		if bit == 0 {
			contador++
		}
	}
	mutexBitmap.Unlock()
	return contador
}

func reservarBloques(nombre string, tamañoDatos int) (int, []int) {
	bloquesAReservar := tamañoDatos
	var indices []int
	indiceBloqueIndice, err := reservarBloqueIndice(nombre)
	if err != nil {
		panic(err)
	}

	for bloquesAReservar > 0 {
		mutexBitmap.Lock()
		for i, bit := range bitmap {
			if bit == 0 {
				bitmap[i] = 1 // Reservo el bloque
				mutexBitmap.Unlock()
				utils_general.LoggearMensaje(FilesystemConfig.Log_level, "## Bloque asignado: <%d> - Archivo: <%s> - Bloques Libres: <%d>", i, nombre, cantidadBloquesLibres()) // LOG OBLIGATORIO
				indices = append(indices, i)                                                                                                                                    // Guardo el número de bloque en índices
				bloquesAReservar -= 1
				break // Salir del bucle interno y continuar con el siguiente bloque
			}
		}
	}

	escribirEnBitmap() // Actualizo archivo bitmap
	return indiceBloqueIndice, indices
}

func reservarBloqueIndice(nombre string) (int, error) {
	mutexBitmap.Lock()
	for i, bit := range bitmap {
		if bit == 0 {
			bitmap[i] = 1 // Reservo el bloque
			mutexBitmap.Unlock()
			utils_general.LoggearMensaje(FilesystemConfig.Log_level, "## Bloque asignado: <%d> - Archivo: <%s> - Bloques Libres: <%d>", i, nombre, cantidadBloquesLibres()) // LOG OBLIGATORIO
			return i, nil
		}
	}
	mutexBitmap.Unlock()
	return -1, errors.New("no hay bloques de índices disponibles")
}

func fragmentarContenido(contenido []byte) [][]byte {
	var fragmentos [][]byte
	// Recorrer el contenido en fragmentos de tamaño blockSize
	for i := 0; i < len(contenido); i += FilesystemConfig.Block_size {
		end := i + FilesystemConfig.Block_size
		if end > len(contenido) {
			end = len(contenido) // Ajusta si el fragmento final es menor que el tamaño del bloque
		}
		fragmentos = append(fragmentos, contenido[i:end])
	}

	return fragmentos
}

func grabarPunterosReservados(bloqueIndice int, indices []int) {
	escribirEnBloque(bloqueIndice, intSliceToByteSlice(indices))
}

func escribirEnBloque(indice int, datos []byte) error {
	// Abrir el archivo en modo lectura/escritura
	mutexBloques.Lock()
	defer mutexBloques.Unlock()
	rutaBloques := crearRutaArchivo(false, "bloques.dat")
	tamañoBloque := FilesystemConfig.Block_size
	f, err := os.OpenFile(rutaBloques, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el archivo: %v", err)
	}
	defer f.Close()

	// Calcular la posición de inicio del bloque en el archivo
	offset := int64(indice * tamañoBloque)

	// Verificar que los datos no excedan el tamaño del bloque
	if len(datos) > tamañoBloque {
		return fmt.Errorf("los datos exceden el tamaño del bloque de %d bytes", tamañoBloque)
	}

	// Escribir los datos en la posición del bloque usando WriteAt
	_, err = f.WriteAt(datos, offset)
	aplicarRetardo()
	if err != nil {
		return fmt.Errorf("error al escribir en el bloque %d: %v", indice, err)
	}

	return nil
}

func escribirContenidoEnBloques(nombre string, bloqueIndice int, contenido []byte) {
	contenidoFragmentado := fragmentarContenido(contenido)
	indices, err := leerBloqueIndice(bloqueIndice, nombre)
	if err != nil {
		panic(err)
	}

	for i, indice := range indices {
		escribirEnBloque(indice, contenidoFragmentado[i])
		utils_general.LoggearMensaje(FilesystemConfig.Log_level, "## Acceso Bloque - Archivo: <%s> - Tipo Bloque: <DATOS> - Bloque File System <%d>", nombre, indice) // LOG OBLIGATORIO
	}
}

func leerBloqueIndice(indice int, nombre string) ([]int, error) {
	rutaBloques := crearRutaArchivo(false, "bloques.dat")
	tamañoBloque := FilesystemConfig.Block_size
	f, err := os.Open(rutaBloques)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el archivo: %v", err)
	}
	defer f.Close()

	offset := int64(indice * tamañoBloque)
	bloqueIndice := make([]byte, tamañoBloque)

	_, err = f.ReadAt(bloqueIndice, offset)
	if err != nil {
		return nil, fmt.Errorf("error al leer el bloque de índice: %v", err)
	}

	utils_general.LoggearMensaje(FilesystemConfig.Log_level, "## Acceso Bloque - Archivo: <%s> - Tipo Bloque: <INDICE> - Bloque File System <%d>", nombre, indice) // LOG OBLIGATORIO
	aplicarRetardo()

	var indices []int
	for i := 0; i < tamañoBloque; i += 4 { // 4 bytes por puntero
		if i+4 > len(bloqueIndice) {
			break
		}
		indice := int(binary.BigEndian.Uint32(bloqueIndice[i : i+4])) // Convierte los 4 bytes en un int
		if indice == 0 {
			break // Suponemos que 0 es un índice no válido y marca el fin
		}
		indices = append(indices, indice)
	}

	return indices, nil
}

func aplicarRetardo() {
	time.Sleep(time.Duration(FilesystemConfig.Block_access_delay) * time.Millisecond)
}
