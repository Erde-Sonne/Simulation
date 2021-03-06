from utils.file_utils import *
from utils.log_utils import *
import json
from path_utils import get_prj_root
import os
from routing.instance import RoutingInput
from itertools import product
import random

traffic_dir=os.path.join(get_prj_root(),"cache/traffic")
files={
	"video":os.path.join(traffic_dir,"video.traffic.txt"),
	"iot":os.path.join(traffic_dir,"iot.traffic.txt"),
	"voip":os.path.join(traffic_dir,"voip.traffic.txt"),
	"ar":os.path.join(traffic_dir,"ar.traffic.txt")
}

traffic={
	"video":[],
	"iot":[],
	"voip":[],
	"ar":[]
}

res=[]

for traffic_type,fn in files.items():
	count=0
	with open(fn,"r") as fp:
		lines=fp.readlines()
		lines=[l[l.index("{"):] for l in lines]
		for l in lines:
			obj=json.loads(l)
			if int(obj["average1"])+int(obj["average2"])+int(obj["average3"])<1000:
				continue

			stats=obj["volumes"]
			# stat=[0 for _ in range(66*65)]
			for idx in range(66*65):
				stats[idx]+=stats[idx+66*65]
				stats[idx]+=stats[idx+66*65*2]
			traffic[traffic_type].append(stats[:66*65])
			count+=1
	debug("{} traffic {}".format(traffic_type,count))


debug("sorted start")
traffic["video"].sort(key=lambda x:len([y for y in x if y>10]),reverse=True)
traffic["iot"].sort(key=lambda x:len([y for y in x if y>10]),reverse=True)
traffic["voip"].sort(key=lambda x :len([y for y in x if y>10]),reverse=True)
traffic["ar"].sort(key=lambda x :len([y for y in x if y>10]),reverse=True)
debug("sorted done")

num=50

res=[
	RoutingInput(video=a, iot=b, voip=c, ar=d) for a in traffic["video"]
	for b in traffic["iot"][:25]
	for c in traffic["voip"][:50]
	for d in traffic["ar"][:100]
]

for _ in range(5):
	random.shuffle(res)


res_fn=os.path.join(traffic_dir,"ilp_inputs4.pkl")
debug("start to save")
save_pkl(res_fn,res)
debug("save done")









