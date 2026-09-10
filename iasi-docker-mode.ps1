<#
.SYNOPSIS
Configura la VM Hyper-V `iasi-docker` para distintos modos de uso.

.DESCRIPTION
Ajusta conjuntamente:
- Número de procesadores virtuales.
- Memoria dinámica.
- Memoria mínima, inicial y máxima.

Modos disponibles:

  tools
    CPU:       2
    Memoria:   4 / 4 / 8 GB
    Uso:       PlantUML y herramientas ligeras.

  ollama-light
    CPU:       4
    Memoria:   8 / 8 / 16 GB
    Uso:       Ollama con modelos pequeños.

  ollama
    CPU:       8
    Memoria:   16 / 16 / 32 GB
    Uso:       Configuración recomendada para Ollama.

  ollama-heavy
    CPU:       12
    Memoria:   32 / 32 / 64 GB
    Uso:       Modelos grandes o cargas más exigentes.

  ollama-beast
    CPU:       16
    Memoria:   64 / 64 / 128 GB
    Uso:       Modelos muy grandes o pruebas intensivas.

Con GPU-P asignada, Hyper-V puede exigir que la memoria mínima y
la memoria inicial tengan el mismo valor. Todos los perfiles respetan
esa restricción.

La VM debe estar apagada antes de ejecutar el script.

.USAGE
.\set-iasi-docker-mode.ps1 --mode tools
.\set-iasi-docker-mode.ps1 --mode ollama
.\set-iasi-docker-mode.ps1 --mode ollama-heavy

.REQUIREMENTS
- Ejecutar PowerShell como administrador.
- Hyper-V instalado.
- La VM `iasi-docker` debe existir y estar apagada.
#>

$VMName = "iasi-docker"

$Profiles = @{
    "tools" = @{
        Cpu     = 2
        Minimum = 4GB
        Startup = 4GB
        Maximum = 8GB
    }

    "ollama-light" = @{
        Cpu     = 4
        Minimum = 8GB
        Startup = 8GB
        Maximum = 16GB
    }

    "ollama" = @{
        Cpu     = 8
        Minimum = 16GB
        Startup = 16GB
        Maximum = 32GB
    }

    "ollama-heavy" = @{
        Cpu     = 12
        Minimum = 32GB
        Startup = 32GB
        Maximum = 64GB
    }

    "ollama-beast" = @{
        Cpu     = 16
        Minimum = 64GB
        Startup = 64GB
        Maximum = 128GB
    }
}

function Show-Usage {
    Write-Host "Usage:"
    Write-Host "  .\set-iasi-docker-mode.ps1 --mode <mode>"
    Write-Host ""
    Write-Host "Modes:"
    Write-Host "  tools"
    Write-Host "  ollama-light"
    Write-Host "  ollama"
    Write-Host "  ollama-heavy"
    Write-Host "  ollama-beast"
}

if ($args.Count -ne 2 -or $args[0] -ne "--mode") {
    Show-Usage
    exit 1
}

$Mode = $args[1]

if (-not $Profiles.ContainsKey($Mode)) {
    Write-Error "Unknown mode '$Mode'."
    Write-Host ""
    Show-Usage
    exit 1
}

$VM = Get-VM -Name $VMName -ErrorAction Stop

if ($VM.State -ne "Off") {
    throw "La VM '$VMName' debe estar apagada. Estado actual: $($VM.State)."
}

$Profile = $Profiles[$Mode]

Write-Host ""
Write-Host "Configurando '$VMName' en modo '$Mode'..."
Write-Host "CPU      : $($Profile.Cpu)"
Write-Host "Memoria  : $($Profile.Minimum / 1GB) / $($Profile.Startup / 1GB) / $($Profile.Maximum / 1GB) GB"
Write-Host ""

Set-VMProcessor -VMName $VMName -Count $Profile.Cpu

Set-VMMemory `
    -VMName $VMName `
    -DynamicMemoryEnabled $true `
    -MinimumBytes $Profile.Minimum `
    -StartupBytes $Profile.Startup `
    -MaximumBytes $Profile.Maximum

Write-Host "Configuracion aplicada."
Write-Host ""

Get-VMProcessor -VMName $VMName |
    Select-Object VMName, Count

Get-VMMemory -VMName $VMName |
    Select-Object VMName, DynamicMemoryEnabled, Minimum, Startup, Maximum
