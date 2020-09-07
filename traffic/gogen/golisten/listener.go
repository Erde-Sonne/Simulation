package main

import (
	"chandler.com/gogen/common"
	"chandler.com/gogen/utils"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)



var (
	snapshotLen    int32 =1024
	promiscuous          = false
	timeout              =30*time.Second
	delayBaseDir   string
	pktLossBaseDir string
	enablePktLossStats bool
	lid int=0
)

type Listener struct {
	Intf          string
	NWorker       int
	EnableWorkers bool
	SrcPortLower  int
	SrcPortUpper  int
	DstPortLower  int
	DstPortUpper  int
	SrcSubnet     string
	DstSubnet     string
	SrcIPFile string

	DelayBaseDir string
	PktLossDir string
	workers []*worker

	//dispatcher---->worker的packet channel
	packetChannels []chan gopacket.Packet

	//dispatcher---->worker的channel size
	channelSize int64

	//delay sample size
	//延迟采样大小，暂时不用
	delaySampleSize int
}



func (l *Listener)getFilter() (filter string){

	srcIPFilter:=""
	if len(l.SrcIPFile)>0{
		srcIPFilter ="("
		log.Printf("Starting to parse source ip file %s\n",l.SrcIPFile)
		ips,err:=utils.ReadLines(l.SrcIPFile)
		if err!=nil{
			log.Fatalf("Error when parsing source ip file %s\n",l.SrcIPFile)
		}
		//https://serverfault.com/questions/280215/tcpdump-capturing-packets-on-multiple-ip-address-filter
		for _,ip:=range ips[:len(ips)-1]{
			srcIPFilter +=fmt.Sprintf("host %s or ",ip)
		}
		srcIPFilter +=fmt.Sprintf("host %s)",ips[len(ips)-1])
		log.Printf("Constructed src ip filter is %s\n", srcIPFilter)
	}else{
		srcIPFilter=fmt.Sprintf("src net %s",l.SrcSubnet)
	}
	filter=fmt.Sprintf("inbound && %s && src portrange %d-%d && dst portrange %d-%d && ! ip broadcast && ! ether broadcast",
		srcIPFilter,
		l.SrcPortLower,
		l.SrcPortUpper,
		l.DstPortLower,
		l.DstPortUpper,
	)
	return filter
}

func (l *Listener)startDispatcher(stop chan struct{}, periodicFlushChan chan struct{})  {
	log.Println("Dispatcher start")

	handle,err:=pcap.OpenLive(l.Intf, snapshotLen,promiscuous,pcap.BlockForever)
	if err!=nil{
		log.Fatalf("Cannot open %s\n",l.Intf)
	}
	defer handle.Close()
	//set filter
	filter:=l.getFilter()
	log.Printf("Filter:%s\n",filter)

	err=handle.SetBPFFilter(filter)
	if err!=nil{
		log.Fatalf("cannot set bpf filter:%s\n",err)
	}

	packetSource:=gopacket.NewPacketSource(handle,handle.LinkType())
	stopRequested:=false
	for{
		select{
		case <-stop:
			log.Printf("Stop received,now shutting down receiver")
			//handle.Close()
			stopRequested=true
			for i:=0;i<l.NWorker;i++{
				close(l.packetChannels[i])
			}
			return
			//周期性的flush，防止收不到最后一个包导致内存爆掉
		case <-periodicFlushChan:
			//todo direct writeDelayStats to file
			//log.Println("periodically worker completeFlush")
			//?????
			continue
		case packet:=<-packetSource.Packets():
			meta := packet.Metadata()
			captureInfo := meta.CaptureInfo
			captureTime := captureInfo.Timestamp.UnixNano() / 1e6
			if net:=packet.NetworkLayer();net!=nil{
				l4Payload:=packet.TransportLayer().LayerPayload()
				if len(l4Payload)<9{
					continue
				}
				sendTime := utils.BytesToInt64(l4Payload[:8])
				utils.Copy(l4Payload,0,utils.Int64ToBytes(captureTime-sendTime),0,8)
				if !stopRequested{
					l.packetChannels[int(net.NetworkFlow().FastHash())&(l.NWorker-1)]<-packet
				}
			}
			continue
		}
	}
}

func (l *Listener)Start()  {
	ticker := time.NewTicker(3600 * time.Second)
	//register signal
	sigs:=make(chan os.Signal,1)
	flushChan:=make(chan struct{},10)
	signal.Notify(sigs,syscall.SIGINT,syscall.SIGTERM,syscall.SIGKILL)
	stopSig2Dispatcher :=make(chan struct{})
	stopSig2Ticker:=make(chan struct{})

	//start signal listener
	go func() {
		sig:=<-sigs
		log.Printf("received signal %s\n",sig)
		log.Printf("start to shutdown dispatcher\n")
		stopSig2Dispatcher <- struct{}{}
		stopSig2Ticker<- struct{}{}
	}()

	//start ticker
	go func() {
		for {
			select {
			case <-ticker.C:
				flushChan<- struct{}{}
			case <-stopSig2Ticker:
				ticker.Stop()
				return
			}
		}
	}()

	//ctx, cancelFunc := context.WithCancel(context.Background())
	log.Printf("Listener start")
	//waitGroup
	wg:=&sync.WaitGroup{}

	wg.Add(l.NWorker)
	for i:=0;i<l.NWorker;i++{
		go l.workers[i].start(l.packetChannels[i],wg)
	}
	go l.startDispatcher(stopSig2Dispatcher,flushChan)

	wg.Wait()
	//休息60s
	log.Println("Wait for 10 seconds to exit gracefully")
	time.Sleep(time.Duration(15)*time.Second)
	log.Println("All Work done,exiting")
}


func (l *Listener)Init()  {

	var roundUp = func(n int) int {
		n=n-1
		for;n&(n-1)!=0;{
			n=n&(n-1)
		}
		return n*2
	}

	l.delaySampleSize=5
	l.channelSize=327680
	//init channel
	if l.EnableWorkers{

		l.NWorker=roundUp(l.NWorker)
		log.Printf("Listener get workers enabled, #workers: %d\n",l.NWorker)
		l.workers=make([]*worker,0)
		l.packetChannels=make([]chan gopacket.Packet,l.NWorker)
		for i:=0;i<l.NWorker;i++{
			worker :=&worker{
				id:i,
				delaySampleSize:   l.delaySampleSize,
				flowDelay:         make(map[[5]string][]int64),
				flowDelayFinished: utils.NewSpecifierSet(),
				writerChannel:     make(chan *common.FlowDesc,102400),
			}

			//worker.flowWriter=NewDefaultWriter(i,l.DelayBaseDir, worker.writerChannel)
			delayBaseDir=l.DelayBaseDir
			pktLossBaseDir=l.PktLossDir
			worker.fiveTupleToFtype =make(map[[5]string]int)
			l.workers=append(l.workers, worker)
			l.packetChannels[i]=make(chan gopacket.Packet,l.channelSize)
			worker.Init()

		}

	}
}


