# Verificar si se proporcionaron los argumentos necesarios
if ($args.Count -ne 1) {
    Write-Host "Uso: .\cambiar_softlink.ps1 <nuevo_destino>"
    exit 1
}

# Nombre del enlace simbólico
$SOFTLINK = "config.json"
# Nuevo destino proporcionado
$NEW_TARGET = $args[0]

# Verificar si el enlace simbólico existe
if (-not (Test-Path -Path $SOFTLINK -PathType Leaf)) {
    Write-Host "Error: '$SOFTLINK' no es un enlace simbólico existente."
    exit 1
}

# Eliminar el enlace simbólico existente
Remove-Item -Path $SOFTLINK

# Crear un nuevo enlace simbólico con el nuevo destino
New-Item -ItemType SymbolicLink -Path $SOFTLINK -Target $NEW_TARGET

# Verificar si el enlace fue actualizado correctamente
if ($?) {
    Write-Host "Enlace simbólico '$SOFTLINK' actualizado exitosamente para apuntar a '$NEW_TARGET'."
} else {
    Write-Host "Error al actualizar el enlace simbólico."
    exit 1
}
