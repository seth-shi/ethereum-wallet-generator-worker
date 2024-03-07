# ethereum-wallet-generator-nodes

## 分布式生成钱包

![运行图](assets/1.png)


## 运行
* 直接下载系统的二进制文件(或者自行构建)
  * [https://github.com/seth-shi/ethereum-wallet-generator-nodes/releases](https://github.com/seth-shi/ethereum-wallet-generator-nodes/releases)
* 服务端运行(必须有公网服务器), 会输出公网`$url`
  * `./ethereum-wallet-generator-nodes master --prefix=0x0000 --suffix=9999`
* 任意节点运行, 手机, 电脑, 台式机 (会统一从服务端拉取配置, 然后上报进度)
    * `./ethereum-wallet-generator-nodes node --server={$url}`