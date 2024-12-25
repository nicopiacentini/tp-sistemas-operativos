package utils

import (
	"encoding/binary"
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func RecibirSolicitudContexto(w http.ResponseWriter, r *http.Request) {
	var solicitud SolicitudContexto
	utils_general.HandlePostRequest(w, r, &solicitud)
	w.WriteHeader(http.StatusOK)

	utils_general.LoggearMensaje(MemoryConfig.Log_level, "## Contexto <Solicitado> - (PID:TID) - (<%d>:<%d>)", solicitud.Pid, solicitud.Tid) // LOG OBLIGATORIO

	// Obtener contexto de registros y base/limite del proceso solicitado
	contexto := buscarContexto(solicitud.Pid, solicitud.Tid)
	base := memoria.MemoriaSistema[solicitud.Pid].Base
	limite := memoria.MemoriaSistema[solicitud.Pid].Limite

	// Crear contexto de CPU que incluye base y limite
	var contextoDeCPU ContextoDeCPU = ContextoDeCPU{
		ContextoRegistros: contexto,
		Base:              base,
		Limite:            limite,
	}

	// Enviar el contexto a la CPU
	EnviarContextoACPU(solicitud.Pid, solicitud.Tid, contextoDeCPU)
}

func EnviarContextoACPU(pid int32, tid int32, contexto ContextoDeCPU) {
	respuesta := RespuestaContexto{
		Pid:      pid,
		Tid:      tid,
		Contexto: contexto,
	}

	//aplico retardo
	AplicarRetardo(MemoryConfig.Response_Delay)

	utils_general.PostRequest(respuesta, MemoryConfig.Ip_CPU, MemoryConfig.Port_CPU, "recibirContexto")
}

func RecibirSolicitudActualizacionContexto(w http.ResponseWriter, r *http.Request) {
	var contextoEjecucion RespuestaContexto

	utils_general.HandlePostRequest(w, r, &contextoEjecucion)
	w.WriteHeader(http.StatusOK)

	//aplico retardo
	AplicarRetardo(MemoryConfig.Response_Delay)

	pid := contextoEjecucion.Pid
	tid := contextoEjecucion.Tid
	nuevoContexto := contextoEjecucion.Contexto.ContextoRegistros
	actualizarContexto(pid, tid, nuevoContexto)
	utils_general.LoggearMensaje(MemoryConfig.Log_level, "## Contexto <Actualizado>  - (PID:TID) - (<%d>:<%d>)", pid, tid) // LOG OBLIGATORIO
}

// devuelve el contexto de ejecucion relacionado a un cierto PID y TID
func buscarContexto(PID int32, TID int32) Registros {
	contexto := memoria.MemoriaSistema[PID].tidsDePid[TID]
	return contexto.ContextoRegistros
}

// Actualiza el contexto de ejecucion relacionado a un cierto PID y TID
func actualizarContexto(PID int32, TID int32, nuevoContexto Registros) {
	contextoActualizado := memoria.MemoriaSistema[PID].tidsDePid[TID]
	contextoActualizado.ContextoRegistros = nuevoContexto
	memoria.MemoriaSistema[PID].tidsDePid[TID] = contextoActualizado
}

func RecibirSolicitudFetch(w http.ResponseWriter, r *http.Request) {
	var solicitud SolicitudDeFetch

	utils_general.HandlePostRequest(w, r, &solicitud)
	w.WriteHeader(http.StatusOK)

	instruccion := memoria.MemoriaSistema[solicitud.Pid].tidsDePid[solicitud.Tid].Instrucciones[solicitud.PC]
	respuesta := RespuestaFetch{
		Pid:         solicitud.Pid,
		Tid:         solicitud.Tid,
		Instruccion: instruccion,
	}

	utils_general.LoggearMensaje(MemoryConfig.Log_level, "## Obtener instrucción - (PID:TID) - (<%d>:<%d>) - Instrucción: <%s>", solicitud.Pid, solicitud.Tid, instruccion) // LOG OBLIGATORIO

	//aplico retardo antes de enviar
	AplicarRetardo(MemoryConfig.Response_Delay)

	utils_general.PostRequest(respuesta, MemoryConfig.Ip_CPU, MemoryConfig.Port_CPU, "recibirInstruccion")
}

// recibir solicitud de lectura de memoria de CPU
func RecibirSolicitudLecturaMemoria(w http.ResponseWriter, r *http.Request) {
	var solicitudLectura SolicitudLectura

	utils_general.HandlePostRequest(w, r, &solicitudLectura)
	w.WriteHeader(http.StatusOK)

	// Buscamos el contenido de la memoria usuario usando la dirección física
	contenidoMemoria := buscarEnMemoriaUsuario(solicitudLectura.DireccionFisica)

	// Serializar el contenido de la memoria y responder al CPU
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(contenidoMemoria); err != nil {
		http.Error(w, "Error al codificar el contenido de la memoria", http.StatusInternalServerError)
		return
	}

	utils_general.LoggearMensaje(MemoryConfig.Log_level, "## <Lectura> - (PID:TID) - (<%d>:<%d>) - Dir. Física: <%d> - Tamaño: <%d>", solicitudLectura.Pid, solicitudLectura.Tid, solicitudLectura.DireccionFisica, 4) // LOG OBLIGATORIO
	log.Printf("## Contenido Leido <%d>", contenidoMemoria)
}

// Recibir solicitud de escritura de memoria desde CPU
func RecibirSolicitudEscrituraMemoria(w http.ResponseWriter, r *http.Request) {
	var solicitudEscritura struct {
		DireccionFisica uint32 `json:"direccionFisica"`
		Valor           uint32 `json:"valor"`
		Pid             uint32 `json:"pid"`
		Tid             uint32 `json:"tid"`
	}

	// Manejar solicitud POST y decodificar en estructura
	utils_general.HandlePostRequest(w, r, &solicitudEscritura)

	// Aplicar retardo
	AplicarRetardo(MemoryConfig.Response_Delay)

	// Convertir el valor a un arreglo de 4 bytes
	bytesValor := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytesValor, solicitudEscritura.Valor)

	// Escribir los 4 bytes en la memoria de usuario
	success := escribirEnMemoriaUsuario(solicitudEscritura.DireccionFisica, bytesValor)

	utils_general.LoggearMensaje(MemoryConfig.Log_level, "## <Escritura> - (PID:TID) - (<%d>:<%d>) - Dir. Física: <%d> - Tamaño: <%d>", solicitudEscritura.Pid, solicitudEscritura.Tid, solicitudEscritura.DireccionFisica, len(bytesValor)) // LOG OBLIGATORIO
	log.Printf("## Contenido Escrito <%d>", solicitudEscritura.Valor)

	// Responder si la operación fue exitosa
	if success {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} else {
		http.Error(w, "Error al escribir en memoria", http.StatusInternalServerError)
	}
}

// Función auxiliar para escribir 4 bytes en la memoria de usuario
func escribirEnMemoriaUsuario(direccionFisica uint32, data []byte) bool {
	// Verificar que la dirección más los 4 bytes no exceda el tamaño de la memoria de usuario
	if int(direccionFisica)+4 > len(memoria.MemoriaUsuario) {
		return false // Error si se intenta escribir fuera de los límites
	}
	copy(memoria.MemoriaUsuario[direccionFisica:], data)
	return true
}

// Retardo en la respuesta de memoria
func AplicarRetardo(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// Función decodificador que lee las instrucciones desde un archivo y devuelve un slice con todad las instrucciones
func Decode(nombreArchivo string) ([]string, error) {
	// Leer el archivo
	caminoAlArchivo := MemoryConfig.Instruction_Path + nombreArchivo
	g, err := os.ReadFile(caminoAlArchivo)
	if err != nil {
		return nil, err
	}
	// Convertir el contenido del archivo a string
	lineasDeinstrucion := string(g)
	var instrucciones []string = strings.Split(lineasDeinstrucion, MemoryConfig.Fin_de_Linea)
	return instrucciones, nil
}
