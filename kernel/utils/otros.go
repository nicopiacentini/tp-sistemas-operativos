package utils

import (
	"github.com/sisoputnfrba/tp-golang/utils_general"
	"reflect"
	"runtime"
	"strings"
)

func encolar[T any](elemento T, lista *[]T) { // Función genérica para encolar un elemento en una lista
	*lista = append(*lista, elemento)
}

func pop[T any](lista *[]T) T { // Función para sacar el primer elemento de una lista
	if len(*lista) == 0 {
		var zero T
		return zero
	}
	elemento := (*lista)[0]
	*lista = (*lista)[1:]
	return elemento
}

func pertenece(elemento interface{}, lista interface{}) bool { // Función que verifica si un elemento pertenece a una lista genérica
	listaValor := reflect.ValueOf(lista)

	// Verificar que lista sea un slice o un array
	if listaValor.Kind() != reflect.Slice && listaValor.Kind() != reflect.Array {
		panic("lista debe ser un slice o un array")
	}

	for i := 0; i < listaValor.Len(); i++ {
		if reflect.DeepEqual(elemento, listaValor.Index(i).Interface()) {
			return true
		}
	}
	return false
}

func logSyscall() { // Loggea la syscall
	tengoElControl = true
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		utils_general.LoggearMensaje(KernelConfig.Log_level, "Error obteniendo el nombre de la función")
		return
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		utils_general.LoggearMensaje(KernelConfig.Log_level, "Error obteniendo el nombre de la función")
		return
	}

	funcName := fn.Name()
	parts := strings.Split(funcName, ".")
	if len(parts) > 1 {
		funcName = parts[len(parts)-1]
	}

	utils_general.LoggearMensaje(KernelConfig.Log_level, "## (<%d>:<%d>) - Solicitó syscall: <%s>", hiloEnEjecucion.Pid, hiloEnEjecucion.Tid, funcName) // LOG OBLIGATORIO
}

func insertarHiloMultinivel(tcb Tcb) {
	colaNueva := []Tcb{tcb}
	colasMultinivel[int(tcb.Prioridad)] = colaNueva
}
