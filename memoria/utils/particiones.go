package utils

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"log"
	"net/http"
	"sort"
)

func puedeCompactar(tamanio int) bool {
	noOcupadas := filtrarParticionesDinamicas(estaLibre)
	espacioLibre := sumarLimites(noOcupadas)

	return espacioLibre >= tamanio
}

func filtrarParticionesDinamicas(condicionOcupada func(Particion) bool) []Particion {
	var particionesLibres []Particion

	for _, particion := range particiones {
		if condicionOcupada(particion) { // si no esta ocupada la meto en mi lista de particiones
			particionesLibres = append(particionesLibres, particion)
		}
	}

	return particionesLibres
}

func estaLibre(particion Particion) bool {
	return !particion.ocupado
}

func sumarLimites(particiones []Particion) int {
	suma := 0
	for _, particion := range particiones {
		suma += particion.limite
	}
	return suma
}

func Compactar(w http.ResponseWriter, r *http.Request) {
	memoriaVieja := memoria.MemoriaUsuario

	// Ordenar las particiones ocupadas primero
	sort.Slice(particiones, func(i, j int) bool {
		return particiones[i].ocupado && !particiones[j].ocupado
	})

	fmt.Println("Particiones después de ordenar ocupadas primero:")
	for _, p := range particiones {
		fmt.Printf("PID: %d, Base: %d, Límite: %d, Ocupado: %t\n", p.pid, p.base, p.limite, p.ocupado)
	}

	base := 0 // Base inicial de la memoria compactada
	nuevasParticiones := []Particion{}

	// Compactar particiones ocupadas
	for _, particion := range particiones {
		if particion.ocupado {
			// Validar tamaño de memoria antes de copiar
			if base+particion.limite > len(memoria.MemoriaUsuario) ||
				particion.base+particion.limite > len(memoriaVieja) {
				fmt.Println("Error: los límites de las particiones exceden la memoria disponible.")
				return
			}

			fmt.Printf("Compactando partición -> PID: %d, Base: %d, Límite: %d, Nuevo Base: %d\n", particion.pid, particion.base, particion.limite, base)

			// Copiar datos y actualizar base
			copy(memoria.MemoriaUsuario[base:base+particion.limite], memoriaVieja[particion.base:particion.base+particion.limite])
			particion.base = base
			base += particion.limite

			// Actualizar contexto
			if contexto, existe := memoria.MemoriaSistema[particion.pid]; existe {
				contexto.Base = uint32(particion.base)
				contexto.Limite = uint32(particion.limite)
				memoria.MemoriaSistema[particion.pid] = contexto
			}

			nuevasParticiones = append(nuevasParticiones, particion)
		}
	}

	// Consolidar espacio libre
	espacioLibre := MemoryConfig.Memory_Size - base
	if espacioLibre > 0 {
		nuevasParticiones = append(nuevasParticiones, Particion{
			base:    base,
			limite:  espacioLibre,
			ocupado: false,
			pid:     -1,
		})
	}

	particiones = nuevasParticiones

	fmt.Println("Particiones después de compactar:")
	for _, p := range particiones {
		fmt.Printf("PID: %d, Base: %d, Límite: %d, Ocupado: %t\n", p.pid, p.base, p.limite, p.ocupado)
	}

	// Notificar al kernel
	compactacion := CompactacionRequest{
		Estado: "Finalizada",
	}
	fmt.Println("Enviando solicitud de compactación finalizada al kernel")
	utils_general.PostRequest(compactacion, MemoryConfig.Ip_Kernel, MemoryConfig.Port_Kernel, "compactacionFinalizada")
}

func liberarParticion(pid int32) {

	if MemoryConfig.Scheme == "FIJAS" {
		// En esquema de particiones fijas, solo se marca como libre
		for i, particion := range particiones {
			if particion.pid == pid {
				particiones[i].ocupado = false
				particiones[i].pid = -1 // Limpia el PID para la partición
				break
			}
		}
	} else {
		// En esquema de particiones dinámicas, liberamos y consolidamos
		for i, particion := range particiones {
			if particion.pid == pid {
				particiones[i].ocupado = false
				particiones[i].pid = -1 // Limpia el PID para la partición
				consolidarParticionesLibres(i)
				break
			}
		}
	}
}

// Función para consolidar particiones libres adyacentes en particiones dinámicas
func consolidarParticionesLibres(index int) {
	// Si la partición liberada tiene adyacentes libres, consolídalas en una sola
	if index > 0 && !particiones[index-1].ocupado { // Verifica si la partición anterior es libre
		particiones[index-1].limite += particiones[index].limite            // Extiende el límite
		particiones = append(particiones[:index], particiones[index+1:]...) // Elimina la partición actual
		index--                                                             // Ajusta el índice para volver a revisar
	}
	if index < len(particiones)-1 && !particiones[index+1].ocupado { // Verifica si la siguiente es libre
		particiones[index].limite += particiones[index+1].limite              // Extiende el límite
		particiones = append(particiones[:index+1], particiones[index+2:]...) // Elimina la partición siguiente
	}
}

func asignarParticion(tamanio int) int { //devuele la posicion en la list de particones de la particion asignada
	switch MemoryConfig.Search_Algorithm {
	case "FIRST":
		return algFirstFit(tamanio)
	case "BEST":
		return algBestFit(tamanio)
	case "WORST":
		return algWorstFit(tamanio)
	default:
		log.Fatalf("no se detecto un algoritmo valido")
	}
	return -1
}

func algFirstFit(tamanio int) int { //devuelve la primera particion que encuentra que satisfaga el tamanio

	for index, particion := range particiones {
		if particion.limite >= tamanio && !particion.ocupado {
			if MemoryConfig.Scheme == "DINAMICAS" && particion.limite > tamanio {
				dividirParticionDinamica(index, tamanio, particion.limite-tamanio)
				return index
			} else if MemoryConfig.Scheme == "DINAMICAS" && particion.limite == tamanio {
				ocuparParticionDinamica(index)
				return index
			}
			return index
		}
	}
	return -1
}

func algBestFit(tamanio int) int { //devuelve la particion mas chica disponible que satisfaga el tamanio
	encontrado := false
	indexDeMenorTamanioActual := -1
	for index, particion := range particiones {
		if particion.limite >= tamanio && !particion.ocupado && !encontrado {
			encontrado = true
			indexDeMenorTamanioActual = index
		}
		if particion.limite >= tamanio && !particion.ocupado && particion.limite < particiones[indexDeMenorTamanioActual].limite {
			indexDeMenorTamanioActual = index
		}
		if tamanio == particion.limite && !particion.ocupado {
			if MemoryConfig.Scheme == "DINAMICAS" {
				ocuparParticionDinamica(index)
				return indexDeMenorTamanioActual
			}
			particiones[index].ocupado = true
			return indexDeMenorTamanioActual
		}
	}
	if encontrado {
		if MemoryConfig.Scheme == "DINAMICAS" {
			dividirParticionDinamica(indexDeMenorTamanioActual, tamanio, particiones[indexDeMenorTamanioActual].limite-tamanio)
			return indexDeMenorTamanioActual
		}
		return indexDeMenorTamanioActual
	}
	return -1
}

func algWorstFit(tamanio int) int { //devuelve la particon mas grande disponible que satisfaga el tamanio
	sort.Slice(particiones, func(i, j int) bool { //funcion que ordena de mayor a menor segun el tamaño de las particiones
		return particiones[i].limite > particiones[j].limite
	})
	for index, particion := range particiones {
		if !particion.ocupado {
			if particion.limite < tamanio {
				return -1
			} else if MemoryConfig.Scheme == "FIJAS" {
				particiones[index].ocupado = true
				return index
			}
			if MemoryConfig.Scheme == "DINAMICAS" && particion.limite >= tamanio {
				if particion.limite == tamanio {
					ocuparParticionDinamica(index)
				} else {
					dividirParticionDinamica(index, tamanio, particiones[index].limite-tamanio)
				}
				return index
			}
		}
	}
	return -1
}

func dividirParticionDinamica(index int, tamanioProceso int, particionLibre int) { // ejemplo si el proceso pesa 3 unidades de memoria
	//crear nueva particion ADELANTE DE LA DEL INDEX: {5,7,3,6,8} -> {3,2,7,3,6,8}
	nuevaParticion := Particion{
		pid:     -1, // hardcodeado para que luego se cambie
		base:    particiones[index].base,
		limite:  tamanioProceso,
		ocupado: true,
	}
	particiones[index].base = particiones[index].base + tamanioProceso
	particiones[index].ocupado = false
	particiones[index].limite = particionLibre
	particiones = append(particiones[:index], append([]Particion{nuevaParticion}, particiones[index:]...)...)
	//pongo como ocupada la nueva particion
	// modifico el tamanio de la particion vieja
}

func ocuparParticionDinamica(index int) {
	particiones[index].ocupado = true
}

// mostrar particiones para debuggear
func mostrarParticiones(particiones []Particion) {
	// Indica que la goroutine ha terminado
	for _, p := range particiones {
		if p.ocupado {
			log.Printf("Partición ocupada -> Base: %d, Límite: %d, PID: %d\n", p.base, p.limite, p.pid)
		} else {
			log.Printf("Partición libre -> Base: %d, Límite: %d\n", p.base, p.limite)
		}
	}
}
