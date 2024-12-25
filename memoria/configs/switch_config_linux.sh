#!/bin/bash

# Nombre fijo del enlace simbólico
SOFTLINK="config.json"

# Verificar que se pase el nuevo destino como argumento
if [ -z "$1" ]; then
    echo "Uso: $0 <nuevo destino>"
    exit 1
fi

# Establecer el nuevo destino
NUEVO_DESTINO="$1"

# Verificar si el archivo de destino existe
if [ ! -f "$NUEVO_DESTINO" ]; then
    echo "Error: El archivo de destino '$NUEVO_DESTINO' no existe."
    exit 2
fi

# Verificar si el enlace simbólico ya existe
if [ -L "$SOFTLINK" ]; then
    echo "El enlace simbólico '$SOFTLINK' ya existe. Eliminándolo..."
    rm "$SOFTLINK"
else
    echo "El enlace simbólico '$SOFTLINK' no existe. Creando uno nuevo..."
fi

# Crear el nuevo enlace simbólico
ln -s "$NUEVO_DESTINO" "$SOFTLINK"
if [ $? -eq 0 ]; then
    echo "Enlace simbólico '$SOFTLINK' ahora apunta a '$NUEVO_DESTINO'."
else
    echo "Error al crear el enlace simbólico."
    exit 3
fi

