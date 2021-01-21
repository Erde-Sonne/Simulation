from utils.file_utils import save_pkl, save_json
from utils.log_utils import info, debug
from common.graph import NetworkTopo

node_labels = {"level1": list(range(12))}

# 生成topo
topo = [[[-1, -1, -1, -1] for _ in range(12)] for _ in range(12)]
topo1 = [[ 0 for _ in range(12)] for _ in range(12)]
link = [100, 0, 0, 0]


def connect(i, j):
	global topo
	topo[i][j] = link
	topo[j][i] = link
	topo1[i][j] = 1
	topo1[j][i] = 1


# level1 内部连接
for idx, node in enumerate(node_labels["level1"]):
	n = len(node_labels["level1"])
	connect(idx, (idx + 1) % n)
	connect(idx, (idx + 11) % n)

import networkx as nx

g = nx.havel_hakimi_graph([3 for _ in range(12)])

for i in range(12):
	for j in range(12):
		if i == j: continue
		if j in g[i]:
			connect(i, j)

net = NetworkTopo(topo)

debug(net.g.number_of_edges())


from path_utils import get_prj_root
import os

static_dir = os.path.join(get_prj_root(), "static")
save_json(os.path.join(static_dir,"topo.json"),{"topo":topo1})

res = []
for _ in range(44):
	res.append({
		"topo": topo,
		"duration": 0
	})

save_pkl(os.path.join(static_dir, "militaryAC.pkl"), res)
