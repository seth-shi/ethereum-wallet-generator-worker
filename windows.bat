@echo off
chcp 65001

set bin_name="ethereum-wallet-generator-nodes.exe"
set zip_name="ethereum-wallet-generator-nodes-v9.9.9-windows-amd64.zip"
set download_url="https://github.com/seth-shi/ethereum-wallet-generator-nodes/releases/download/v9.9.9/ethereum-wallet-generator-nodes-v9.9.9-windows-amd64.zip"

:: 下载文件
if exist "%zip_name%" (
    del /f "%zip_name%"
    echo 文件 "%zip_name%" 已被删除。
)
wget -O "%zip_name%" "%download_url%"

:: 解压文件
if exist "%bin_name%" (
    del /f "%bin_name%"
    echo 文件 "%bin_name%" 已被删除。
)
powershell -c " Expand-Archive %zip_name% ."

:: 校验文件是否正确
echo 远程文件 md5
@wget -O "%zip_name%.md5" "%download_url%.md5"
type "%zip_name%.md5"
echo 下载文件 md5
@certutil -hashfile "%zip_name%" MD5

:: 删除压缩文件
del "%zip_name%.md5"
@REM REM del "%zip_name%"
echo 更新完成
exit /b 0