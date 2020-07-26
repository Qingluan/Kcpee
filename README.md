# kcpee
a kcp-go + tcp relay

# 使用说明

## 普通代理

> 第一次使用的时候最好先测试线路的最快配置
>> Kcpee -Test -c /你的配置文件目录/

1. 使用json文件 / 参数和shadowsocks一样 除了 **加密函数支持 "tea/ aes/ xtea / 3des "**

```sh
~ kcpee -c some.json 
```

2. 使用ssuri 

```sh
~ kcpee -c ss://{base64}   [和shadowosocks 的uri一样]
```

3. 指定目录 ， 会自动扫描该目录下的所有json 文件，并全部解析配置，这样可以配合后续的强化文件同时使用多条线路

```sh
~ kcpee -c /xxx/xxx/
```

## 自动生成路由 （如果只有单挑线路这个忽略）｜（第一次使用前建议执行一次）

> 关掉你的chrome， 并且保证在命令行可执行 sqlite3

```sh
~ Kcpee -Test -c /配置文件目录/
```

## 反向代理

1. 第一步在被控端 (A)

```sh
~ Kcpee -c B.config 配置如上 -T # 默认转移流量到 127.0.0.1:22 / 可后续通过控制端动态更改 /或者加 ‘-connect 127.0.0.1:8080’  调整
 # ServerB <-> A 
 # 这样就将在服务器上开启一个tcp 监听，可以通过tcp 直接连接这个端口  
```
2. 连接被控端 （C)

```sh
~ kcpee -c B.config -C -connect "A 的来源ip" 
# A 的来源ip 会在
```
