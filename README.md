## 以太坊靓号生成器
* github地址 [https://github.com/seth-shi/ethereum-wallet-generator-nodes](https://github.com/seth-shi/ethereum-wallet-generator-nodes)

![运行图](assets/1.png)


## 注意
* 生成之后的助记词是被加密的, 需要这样子解密
* `wallet.csv` 文件中包含如下内容
```text
rmkcmvvaevdzupda," "看仓库readme说明解密"
0x000Bb65F2e742F10492546bDaFD74F9BCEA427BD,nxI3NdxHKsIezLVu7UFBdkDM9pLLnEvgP/ZxlHtzq1ICtTDhhkKNCj8T1bAbyVKarh3wnYEng3hx1dPrxOEmYla9JO+OqIFTWc9VXdq8pay20DL+Kixy6mS3o4K0r11/Z3E5YQ==
```
* 解密
```text
ethereum-wallet-generator-nodes decrypt --key=rmkcmvvaevdzupda --data=Z8qTxdkhRlTwbZEO+dppDCY7F3TWcsP/4B9q4V456stR2Ss7oFdHl5nWhMBNnrAP7J3H3xkk3jEzUqnAdW9V2nfgJKEsKh1QlCKy7lnfzKbVlKl020x5NAdfvzgeZ8+LQj5GgT9VbHI/jeO1XWc= --limit=12
# 助记词 0-12 开始
# glory helmet alter river mystery method fuel number acoustic address parade flavor
# 助记词 0-12 结束
# 你也可以分尅几次解密 (--limit=6 --offset=0)
# 解密后之后保存好, 删除所有记录
# 解密后之后保存好, 删除所有记录
# 解密后之后保存好, 删除所有记录
```

## 快开开始
### 一键安装

| 系统        | 命令                                                                                                                                                                                 |
|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Linux**    | `wget -O upgrade.sh https://raw.githubusercontent.com/seth-shi/ethereum-wallet-generator-nodes/master/linux.sh  && sh upgrade.sh`                                                  |
| **Windows** | 下载 [windows.ps1](https://raw.githubusercontent.com/seth-shi/ethereum-wallet-generator-nodes/master/windows.ps1) <br>然后右键`windows.ps1`使用`Powershell`运行 <br> 或者在命令行运行`./windows.ps1` |
| **macOS** | 请查看手动安装部分                                                                                                                                                                          |

#### 手动安装
<details> <summary>点击查看详情(一键安装成功的, 忽略此项)</summary>

* 下载二进制文件(或者自行构建), 现在基本都是`amd`架构, 如果不能运行,自行把链接中的`amd`换成`arm`(比如安卓中)

| 系统        | 命令                                                                                                                                                                                                                                                                                                 |
|-----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Linux**    | `wget https://github.com/seth-shi/ethereum-wallet-generator-nodes/releases/download/v9.9.9/ethereum-wallet-generator-nodes-v9.9.9-linux-amd64.tar.gz`                                                                                                                                              |
| **Windows** | 点击下载[https://github.com/seth-shi/ethereum-wallet-generator-nodes/releases/download/v9.9.9/ethereum-wallet-generator-nodes-v9.9.9-windows-amd64.zip](https://github.com/seth-shi/ethereum-wallet-generator-nodes/releases/download/v9.9.9/ethereum-wallet-generator-nodes-v9.9.9-windows-amd64.zip) |
| **macOS** | 点击下载[https://github.com/seth-shi/ethereum-wallet-generator-nodes/releases/download/v9.9.9/ethereum-wallet-generator-nodes-v9.9.9-darwin-amd64.tar.gz](https://github.com/seth-shi/ethereum-wallet-generator-nodes/releases/download/v9.9.9/ethereum-wallet-generator-nodes-v9.9.9-darwin-amd64.tar.gz)]                                                                                                                                           |

* 解压文件

| 系统        | 命令                                                                  |
|-----------|---------------------------------------------------------------------|
| **Linux**    | `tar xvf ethereum-wallet-generator-nodes-v9.9.9-linux-amd64.tar.gz` |
| **Windows** | 双击解压                                                                |
| **macOS** | 双击解压                                                                |
</details>


### `master`运行 (必须有公网服务器), 会输出公网`$url`

| 系统        | 命令                                                                         |
|-----------|----------------------------------------------------------------------------|
| **Linux**    | `./ethereum-wallet-generator-nodes master --prefix=0x0000 --suffix=9999`   |
| **Windows** | `ethereum-wallet-generator-nodes.exe master --prefix=0x0000 --suffix=9999` |
| **macOS** | `./ethereum-wallet-generator-nodes master --prefix=0x0000 --suffix=9999`   |

### 节点运行, 手机, 电脑, 台式机 (会统一从服务端拉取配置, 然后上报进度)

| 系统        | 命令                                                                                                   |
|-----------|------------------------------------------------------------------------------------------------------|
| **Linux**    | `./ethereum-wallet-generator-nodes  node --server="{$url}"`                                          |
| **Windows** | `.\ethereum-wallet-generator-nodes node --server="{$url}"`                                           |
| **macOS** | `./ethereum-wallet-generator-nodes master  node --server="{$url}"`                                   |

## 已生成成功
* `0x0000f31fE398080ef162841Cae9111735C919999`

![运行图](assets/2.png)
![运行图](assets/3.png)
