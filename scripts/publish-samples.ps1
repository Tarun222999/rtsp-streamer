$ErrorActionPreference = 'Stop'

$projectRoot = Split-Path -Parent $PSScriptRoot
$ffmpegPath = (Get-Command ffmpeg -ErrorAction Stop).Source
$logDir = Join-Path $projectRoot 'dev/logs'
$publishers = @()

# Check every sample before starting any publishers.
foreach ($number in 1..4) {
    $sample = Join-Path $projectRoot "dev/media/camera$number.mp4"
    if (-not (Test-Path -LiteralPath $sample)) {
        throw "Missing sample: $sample"
    }
}

New-Item -ItemType Directory -Force -Path $logDir | Out-Null

try {
    foreach ($number in 1..4) {
        $sample = Join-Path $projectRoot "dev/media/camera$number.mp4"
        $url = "rtsp://127.0.0.1:8554/camera$number"

        $ffmpegArgs = @(
            '-hide_banner', '-loglevel', 'warning', '-nostats',
            '-nostdin', '-re', '-stream_loop', '-1',
            '-i', "`"$sample`"",
            '-an', '-c:v', 'copy',
            '-f', 'rtsp', '-rtsp_transport', 'tcp', $url
        )

        $process = Start-Process `
            -FilePath $ffmpegPath `
            -ArgumentList $ffmpegArgs `
            -WindowStyle Hidden `
            -RedirectStandardOutput (Join-Path $logDir "camera$number.stdout.log") `
            -RedirectStandardError (Join-Path $logDir "camera$number.stderr.log") `
            -PassThru

        $publishers += $process
        Write-Host "Started camera$number publisher (PID $($process.Id))"
    }

    Write-Host 'Press Ctrl+C to stop all four publishers.'

    while ($true) {
        Start-Sleep -Seconds 1

        foreach ($process in $publishers) {
            if ($process.HasExited) {
                throw "Publisher PID $($process.Id) exited. Check dev/logs."
            }
        }
    }
}
finally {
    foreach ($process in $publishers) {
        if (-not $process.HasExited) {
            Stop-Process -Id $process.Id -ErrorAction SilentlyContinue
        }
    }

    Write-Host 'Publishers stopped.'
}