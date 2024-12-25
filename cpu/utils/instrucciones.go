package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"net/http"
)

// Funciones auxiliares para manejar cada operación
func setRegistro(contexto *Contexto, registro string, valor uint32) {
	switch registro {
	case "AX":
		contexto.ContextoRegistros.AX = valor
	case "BX":
		contexto.ContextoRegistros.BX = valor
	case "CX":
		contexto.ContextoRegistros.CX = valor
	case "DX":
		contexto.ContextoRegistros.DX = valor
	case "EX":
		contexto.ContextoRegistros.EX = valor
	case "FX":
		contexto.ContextoRegistros.FX = valor
	case "GX":
		contexto.ContextoRegistros.GX = valor
	case "HX":
		contexto.ContextoRegistros.HX = valor
	case "PC":
		contexto.ContextoRegistros.PC = valor
	}
}

func readMem(contexto *Contexto, registroDatos string, registroDireccion string, pid int32, tid int32) {
	// Obtener la dirección lógica desde el registro
	direccionLogica := getValorRegistro(contexto, registroDireccion)

	// Validar la dirección lógica y traducirla a dirección física, asegurando 4 bytes
	direccionFisica, err := MMU(contexto, direccionLogica, 4, pid, tid)
	if err != nil {
		fmt.Printf("Error al traducir la dirección lógica a física: %v", err)
		return
	}
	solicitudLectura := SolicitudLectura{Pid: pid, Tid: tid, DireccionFisica: direccionFisica}

	// Crear la solicitud para leer memoria
	body, err := json.Marshal(solicitudLectura)
	if err != nil {
		fmt.Printf("Error al serializar la solicitud de lectura de memoria: %v", err)
		return
	}

	memoriaURL := fmt.Sprintf("http://%s:%d/read", CpuConfig.Ip_memory, CpuConfig.Port_memory)
	req, err := http.NewRequest("POST", memoriaURL, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error al crear la solicitud HTTP para lectura de memoria: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Enviar la solicitud y manejar la respuesta
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error al enviar la solicitud HTTP para lectura de memoria: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error en la respuesta de memoria al leer: código %d", resp.StatusCode)
		return
	}

	// Decodificar la respuesta y guardar el valor en el registro
	var contenidoMemoria uint32
	if err := json.NewDecoder(resp.Body).Decode(&contenidoMemoria); err != nil {
		fmt.Printf("Error al decodificar la respuesta de lectura de memoria: %v", err)
		return
	}

	setRegistro(contexto, registroDatos, contenidoMemoria)

	utils_general.LoggearMensaje(CpuConfig.Log_level, "## TID: <%d> - Acción: <LEER> - Dirección Física: <%d>", tid, direccionFisica) // LOG OBLIGATORIO
}

func writeMem(contexto *Contexto, registroDireccion string, registroDatos string, pid int32, tid int32) {
	// Obtener la dirección lógica desde el registro
	direccionLogica := getValorRegistro(contexto, registroDireccion)

	// Traducir la dirección lógica a física usando la MMU, validando 4 bytes
	direccionFisica, err := MMU(contexto, direccionLogica, 4, pid, tid)
	if err != nil {
		fmt.Printf("Error al traducir la dirección lógica a física: %v", err)
		return
	}

	// Obtener el valor del registro correspondiente
	valor := getValorRegistro(contexto, registroDatos)

	body := map[string]interface{}{
		"direccionFisica": direccionFisica,
		"valor":           valor,
		"pid":             pid,
		"tid":             tid,
	}

	// Enviar la solicitud a la memoria
	utils_general.PostRequest(body, CpuConfig.Ip_memory, CpuConfig.Port_memory, "write")

	utils_general.LoggearMensaje(CpuConfig.Log_level, "## TID: <%d> - Acción: <ESCRIBIR> - Dirección Física: <%d>", tid, direccionFisica) // LOG OBLIGATORIO
}

func sumRegistros(contexto *Contexto, registroDestino string, registroOrigen string) {
	// Obtiene los valores de ambos registros
	valorDestino := getValorRegistro(contexto, registroDestino)
	valorOrigen := getValorRegistro(contexto, registroOrigen)
	// Suma los valores y almacena el resultado en el registro destino
	resultado := valorDestino + valorOrigen
	setRegistro(contexto, registroDestino, resultado)
}

func subRegistros(contexto *Contexto, registroDestino string, registroOrigen string) {
	// Obtiene los valores de ambos registros
	valorDestino := getValorRegistro(contexto, registroDestino)
	valorOrigen := getValorRegistro(contexto, registroOrigen)
	// Resta los valores y almacena el resultado en el registro destino
	resultado := valorDestino - valorOrigen
	setRegistro(contexto, registroDestino, resultado)
}

func jnz(contexto *Contexto, registro string, nuevaInstruccion uint32) {
	// Verifica si el valor del registro es diferente de cero
	if getValorRegistro(contexto, registro) != 0 {
		contexto.ContextoRegistros.PC = nuevaInstruccion
	}
}

func logRegistro(contexto *Contexto, registro string) {
	// Obtiene el valor del registro
	valor := getValorRegistro(contexto, registro)

	// Escribe el valor en el log
	utils_general.LoggearMensaje(CpuConfig.Log_level, "Valor del registro %s: %d", registro, valor)
}

func getValorRegistro(contexto *Contexto, registro string) uint32 {
	switch registro {
	case "AX":
		return contexto.ContextoRegistros.AX
	case "BX":
		return contexto.ContextoRegistros.BX
	case "CX":
		return contexto.ContextoRegistros.CX
	case "DX":
		return contexto.ContextoRegistros.DX
	case "EX":
		return contexto.ContextoRegistros.EX
	case "FX":
		return contexto.ContextoRegistros.FX
	case "GX":
		return contexto.ContextoRegistros.GX
	case "HX":
		return contexto.ContextoRegistros.HX
	case "PC":
		return contexto.ContextoRegistros.PC
	}
	return 0
}

// verifica si un string es un registro
func esRegistro(cadena string) bool {
	switch cadena {
	case "AX":
		return true
	case "BX":
		return true
	case "CX":
		return true
	case "DX":
		return true
	case "EX":
		return true
	case "FX":
		return true
	case "GX":
		return true
	case "HX":
		return true
	case "PC":
		return true
	}
	return false
}
