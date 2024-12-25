#!/bin/bash

# Directorio hardcodeado
DIRECTORIO="../filesystem/mount_dir/files/"

# Verifica si el directorio existe y es un directorio
if [ ! -d "$DIRECTORIO" ]; then
    echo "Error: '$DIRECTORIO' no es un directorio válido."
    exit 1
fi

# Elimina todos los archivos, incluyendo los ocultos, pero no las carpetas
find "$DIRECTORIO" -type f -exec rm -f {} +

# Mensaje final
echo "Todos los archivos en '$DIRECTORIO' han sido eliminados exitosamente."
