# 设置变量
$binName = "ethereum-wallet-generator-worker.exe"
$zipName = "ethereum-wallet-generator-worker-v9.9.9-windows-amd64.zip"
$downloadUrl = "https://github.com/seth-shi/ethereum-wallet-generator-worker/releases/download/v9.9.9/ethereum-wallet-generator-worker-v9.9.9-windows-amd64.zip"

# 下载文件
if (Test-Path $zipName) {
    Remove-Item -Path $zipName -Force
    Write-Host "文件 $zipName 已被删除。"
}
Invoke-WebRequest -Uri $downloadUrl -OutFile $zipName

# 解压文件
if (Test-Path $binName) {
    Remove-Item -Path $binName -Force
    Write-Host "文件 $binName 已被删除。"
}
Expand-Archive -Force -Path $zipName -DestinationPath .

# 校验文件是否正确
Write-Host "远程文件 md5"
Invoke-WebRequest -Uri "$downloadUrl.md5" -OutFile "$zipName.md5"
Get-FileHash -Algorithm MD5 -Path "$zipName.md5" | Select-Object -ExpandProperty Hash
Write-Host "下载文件 md5"
Get-FileHash -Algorithm MD5 -Path $zipName | Select-Object -ExpandProperty Hash

# 删除压缩文件
Remove-Item -Path "$zipName.md5" -Force
Remove-Item -Path $zipName -Force

Write-Host "更新完成"
