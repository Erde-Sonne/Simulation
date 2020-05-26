import os
dir_path = os.path.dirname(os.path.realpath(__file__))
import argparse
from argparse import ArgumentParser
from pathlib import Path
import json
from topo.distributed.topobuilder import TopoManager
from path_utils  import get_prj_root


from typing import List
import loguru
logger=loguru.logger


def load_json(filename):
    if Path(filename).is_file():
        with open(filename) as f:
            return json.load(f)
    raise FileNotFoundError

def add(idx1,idx2,topo:List):
	topo[idx1][idx2]=(1000,20,0,0)
	topo[idx2][idx1]=(1000,20,0,0)


if __name__ == '__main__':
	topo=[[(-1,-1,-1,-1) for _ in range(6)] for _ in range(6)]
	config_fn=os.path.join(get_prj_root(),"topo/distributed/mock_config.json")
	config=load_json(config_fn)
	add(0,1,topo)
	add(1,2,topo)
	add(0,3,topo)
	add(1,4,topo)
	add(2,5,topo)

	parser=ArgumentParser()
	parser.add_argument("--id",required=True,type=int)
	parser.add_argument("--intf",required=True,type=str,default="ens33")
	args=parser.parse_args()
	id_=int(args.id)
	intf=args.intf

	manager=None
	manager=TopoManager(config,id_,intf)

	manager.diff_topo(topo)
	manager.tear_down()


	


