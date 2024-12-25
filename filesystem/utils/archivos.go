package utils

import (
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"os"
)

func validarExistenciaArchivos() {
	verificarYCrearArchivo("bitmap.dat")
	verificarYCrearArchivo("bloques.dat")
}

func verificarYCrearArchivo(nombreArchivo string) {
	if !existeArchivo(nombreArchivo) {
		crearArchivo(nombreArchivo, false, 0, 0)
	}
}

func existeArchivo(nombreArchivo string) bool {
	rutaArchivo := crearRutaArchivo(false, nombreArchivo)
	_, err := os.Stat(rutaArchivo)
	return !os.IsNotExist(err)
}

func crearArchivo(nombreArchivo string, esMetadata bool, bloqueIndice int, tamañoMetadata int) {
	rutaArchivo := crearRutaArchivo(esMetadata, nombreArchivo)
	archivo, err := os.Create(rutaArchivo)
	if err != nil {
		panic(err)
	}
	defer archivo.Close()

	tamaño := calcularTamaño(nombreArchivo, tamañoMetadata)

	asignarTamaño(archivo, nombreArchivo, tamaño)

	inicializarArchivo(nombreArchivo, archivo, bloqueIndice, tamañoMetadata)

	utils_general.LoggearMensaje(FilesystemConfig.Log_level, "## Archivo Creado: <%s> - Tamaño: <%d>", nombreArchivo, tamaño) // LOG OBLIGATORIO
}

func asignarTamaño(archivo *os.File, nombreArchivo string, tamaño int) {
	if tamaño <= 0 {
		panic(fmt.Sprintf("Tamaño inválido calculado para el archivo %s: %d", nombreArchivo, tamaño))
	}

	err := archivo.Truncate(int64(tamaño))
	if err != nil {
		panic(err)
	}
}

func inicializarArchivo(nombreArchivo string, archivo *os.File, bloqueIndice int, tamañoMetadata int) {
	switch nombreArchivo {
	case "bitmap.dat":
		inicializarBitmap(archivo)

	case "bloques.dat":
		inicializarBloques(archivo)

	default:
		inicializarDump(bloqueIndice, tamañoMetadata, archivo)
	}
}

func inicializarBitmap(bitmap *os.File) {
	// Creo un slice de bytes todos en 0
	cantidadCeros := FilesystemConfig.Block_count
	bitmapBytes := make([]byte, cantidadCeros)

	_, err := bitmap.Write(bitmapBytes)
	if err != nil {
		panic(err)
	}
}

func inicializarBloques(bloques *os.File) {
	// Creo un slice de bytes todos en 0
	cantidadCeros := FilesystemConfig.Block_count
	bloqBytes := make([]byte, cantidadCeros)

	_, err := bloques.Write(bloqBytes)
	if err != nil {
		panic(err)
	}
}

func inicializarDump(index_block int, size int, dump *os.File) {
	// Crear la estructura con los datos
	metadata := Metadata{
		IndexBlock: index_block,
		Size:       size,
	}

	// Serializar la estructura en JSON
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		panic(err)
	}

	// Escribir el JSON en el archivo
	_, err = dump.Write(jsonData)
	if err != nil {
		panic(err)
	}
}

func crearArchivoMetadata(nombreArchivo string, tamaño int, bloqueIndice int) {
	crearArchivo(nombreArchivo, true, bloqueIndice, tamaño)
}

func crearRutaArchivo(esMetadata bool, nombreArchivo string) string {
	return rutaSegun(esMetadata) + nombreArchivo
}

func calcularTamaño(nombreArchivo string, tamañoMetadata int) int {
	switch nombreArchivo {
	case "bitmap.dat":
		return dividirYRedondearSuperiormente(FilesystemConfig.Block_count, 8)
	case "bloques.dat":
		return FilesystemConfig.Block_count * FilesystemConfig.Block_size
	default:
		return tamañoMetadata
	}
}

func rutaSegun(esMetadata bool) string {
	if esMetadata {
		return FilesystemConfig.Mount_dir + "files/"
	} else {
		return FilesystemConfig.Mount_dir
	}
}

func cargarBitmap() {
	// Leer el contenido del archivo mapeado en un slice de bytes
	mutexBitmap.Lock()
	bitmap = make([]byte, archivoBitmap.Len())
	_, err := archivoBitmap.ReadAt(bitmap, 0)
	mutexBitmap.Unlock()

	if err != nil {
		panic(err)
	}
}

func escribirEnBitmap() {
	// Abrir el archivo con permisos de escritura
	rutaArchivo := crearRutaArchivo(false, "bitmap.dat")
	archivo, err := os.OpenFile(rutaArchivo, os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer archivo.Close()

	// Escribir el contenido del bitmap en el archivo
	mutexBitmap.Lock()
	_, err = archivo.WriteAt(bitmap, 0)
	mutexBitmap.Unlock()
	if err != nil {
		panic(err)
	}
}
