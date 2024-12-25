package utils_general

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
)

type ConfigInterface interface{}

func IniciarConfig(filePath string, config ConfigInterface) ConfigInterface {
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(config)
	if err != nil {
		log.Fatal("Error al decodificar el archivo de configuración:", err)
	}

	// Verificar si la configuración se cargó correctamente
	if reflect.ValueOf(config).IsNil() {
		log.Fatalln("No se pudo cargar la configuración")
	}
	log.Printf("Configuración establecida %+v\n", config)

	return config
}
