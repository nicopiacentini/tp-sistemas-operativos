package utils

import (
	"log"
	"math"
	"strconv"
	"strings"
)

func dividirYRedondearSuperiormente(x, y int) int {
	resultado := float64(x) / float64(y)
	return int(math.Ceil(resultado))
}

func intSliceToByteSlice(ints []int) []byte {
	bytes := make([]byte, len(ints)*4) // Cada int ocupa 4 bytes
	for i, v := range ints {
		bytes[i*4] = byte(v >> 24)
		bytes[i*4+1] = byte(v >> 16)
		bytes[i*4+2] = byte(v >> 8)
		bytes[i*4+3] = byte(v)
	}
	return bytes
}

func obtenerPid(nombre string) int32 {
	partes := strings.Split(nombre, "-")
	if len(partes) < 2 {
		log.Fatalf("Formato inválido: %s", nombre)
	}
	pid, err := strconv.Atoi(partes[0])
	if err != nil {
		log.Fatalf("Error al convertir PID a int: %v", err)
	}
	return int32(pid)
}

func obtenerTid(nombre string) int32 {
	partes := strings.Split(nombre, "-")
	if len(partes) < 2 {
		log.Fatalf("Formato inválido: %s", nombre)
	}
	tid, err := strconv.Atoi(partes[1])
	if err != nil {
		log.Fatalf("Error al convertir TID a int: %v", err)
	}
	return int32(tid)
}
