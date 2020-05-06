# Rule
Roadmap

模块化
每个模块使用dumb，先把系统运行起来，再说
socket+json,所有的编码均为ASCII编码

每个模块下面给出了输入例子（如果有）

# classify
这个模块用于判断属于哪种流
接受统计量，示例为receive.demo.json

返回结果  示例为resp.demo.json "0"代表大流量，"1"低时延

# deploy
存放系统启动脚本、编译脚本等
目前放的是ditg编译脚本，系统运行前首先要编译ditg

# traffic
这个模块主要用于主机生成流量

主机会上报流的统计信息，socket.demo.json
specifier 字段为五元祖，均为字符串，顺序为src_port，dst_port,src_ip,dst_ip,protocol,
stats为统计信息，均为float，顺序如示例

# topo
这个模块主要用于建立topo，目前是mininet，仅支持python2，因为python版本的原因，这个模块相对独立，不依赖于任何外部模块，外部模块使用python3


## 关于流量产生
工具为DITG，但是有bug，目前的patch是使用脚本，进程crash之后重新运行
ITGManager ips_file lambda duration controller_ip port

### DITG工作原理
DITG可以自定义流量特性，自定义包间隔、包长度，这两个量分别由两个文件控制idt、ps文件，这两个文件是由python分析pcap文件产生的
DITG有潜在的问题，比如流的时间无法精确控制，例如（流的时间如果是10s，那么DITG运行时间可能超过10s）
对源码进行了一些修改 放在traffic/ditg/下面

[ DITG手册 ](http://www.grid.unina.it/software/ITG/manual/)

DITG支持Daemon模式，项目采用ITGManager生成Possion到达的流，流根据idt和ps文件产生

# routing
这个模块主要用于决策路由,交互见json



# 运行测试
##  编译ditg
运行deploy/ditg.sh 编译ditg
## Mininet Topo
目前可以自定义
例子是topo/files/topo.json中给出的，三个节点，ABC，B---A---C,目前没有自定义链路QoS
如需自定义topo，修改topo.json
为邻接矩阵，矩阵中的每个元素代表一条链路，链路QoS为带宽、延迟、丢包率，以后的格式也这样
 
运行例子需要ryu controller，默认ip地址为localhost,监听默认端口，并且controller用于接受主机上报的socket端口为10000

## test 
文件夹test下面存放测试代码，目前主要测试socket通信



