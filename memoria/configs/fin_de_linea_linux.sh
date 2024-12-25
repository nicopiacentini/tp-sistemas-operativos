#!/bin/bash

# Directorio actual
DIRECTORIO="."  # Puedes cambiar esto si los archivos están en otra carpeta

# Valor nuevo para "fin_de_linea"
NUEVO_VALOR="\\\\n"

# Iterar sobre todos los archivos JSON en el directorio
for archivo in "$DIRECTORIO"/*.json; do
    if [ -f "$archivo" ]; then
        echo "Modificando $archivo..."
        
        # Reemplazar el contenido del atributo "fin_de_linea"
        sed -i "s/\"fin_de_linea\": \".*\"/\"fin_de_linea\": \"$NUEVO_VALOR\"/" "$archivo"
    fi
done

echo "Modificación completada."
