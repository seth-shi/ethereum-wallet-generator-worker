#!/bin/bash

bin_name="ethereum-wallet-generator-worker"
tar_name="ethereum-wallet-generator-worker.tar.gz"
download_url=

get_arch=$(arch 2> /dev/null)
echo "cpu 架构:$get_arch"

## 再次获取 cpu 架构
if [ ! "$get_arch" ]; then
    get_arch=$(uname -m 2> /dev/null)
fi

if [[ $get_arch =~ "x86_64" ]]; then
    download_url="https://github.com/seth-shi/ethereum-wallet-generator-worker/releases/download/v9.9.9/ethereum-wallet-generator-worker-v9.9.9-linux-amd64.tar.gz"
elif [[ $get_arch =~ "aarch64" ]]; then
    download_url="https://github.com/seth-shi/ethereum-wallet-generator-worker/releases/download/v9.9.9/ethereum-wallet-generator-worker-v9.9.9-linux-arm64.tar.gz"
else
    echo "不支持此 cpu 架构"
    exit 1
fi


delete_file_if_exists() {
    local file="$1"

    if [ -f "$file" ]; then
        rm -f "$file"
        echo "文件 $file 已被删除。"
    fi
}

## 下载文件
delete_file_if_exists "$tar_name"
wget -O "$tar_name" -c "$download_url"

## 解压文件
delete_file_if_exists "$bin_name"
tar -xvf "$tar_name"

## 校验文件是否正确
echo "远程文件 md5"
wget -q -O - "$download_url.md5"
echo "下载文件 md5"
md5sum "$tar_name"

## 删除压缩文件
rm "$tar_name"
echo "更新完成"
