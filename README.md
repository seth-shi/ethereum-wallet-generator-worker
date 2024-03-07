# ethereum-wallet-generator-nodes

## 分布式生成钱包

![运行图](assets/1.png)


## 运行
```shell
// 编译二进制文件
go build -o eth
// 服务端运行(必须有公网服务器), 会输出公网 $url
./eth master --prefix=0x0000 --prefix=9999 --port=9090
// 任意节点运行, 手机, 电脑, 台式机 (会统一从服务端拉取配置, 然后上报进度)
./eth node --server=$url
```