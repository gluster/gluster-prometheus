 
#### Gluster prometheus end point schema: 

#### Cluster Metrics:
| Gluster Prometheus Schema | example | comments |
|--------|----------|--|
|gluster_volume_total_count | gluster_volume_total_count{} 20 |
|gluster_volume_started_count | gluster_volume_started_count {} 10 | 
| gluster_volume_stopped_count | gluster_volume_stopped_count{} 10 |

#### Volume Metrics : 

|Gluster Promethus schema| example| comments |
|--------|-----------|---|
|gluster_peer_status |	gluster_peer_status{peerid="xxxxx"} 1 |	1 - connected, 0 - disconnected	
|gluster_volume_state |	gluster_volume_status{volume="xxx"} 1	|  1 - up, 0 - down, 2 - partially up, 3 - degraded
|gluster_volume_fop_max_latency	| gluster_volume_max_latency{volume=""xxx"", brick_path=""xxxx"",fop=""xxx"", host="xxx"}  800970.00| 	
|gluster_volume_fop_min_latency	|gluster_volume_fop_min_latency{volume="xxx", brick_path="xxx", fop="xxx", host="xxx"} 21.00	|
|gluster_volume_fop_avg_latency|	gluster_volume_fop_avg_latency{volume="xxx", brick_path="xxx", fop="xxx", host="xxx"} 1132.00|	
|gluster_volume_fop_calls |gluster_volume_fop_calls{volume="xxx". brick_path="xxx", fop="xxx", host="xxx"} 13	|
|gluster_volume_fop_percentage_latency| gluster_volume_fop_percentage_latency(volume="xxx", brick_path="xxx", fop="xxx", host="xxx"} 13	|
|gluster_volume_total_bytes| gluster_volume_total_bytes{volume="xxx"}  104857600	|
|gluster_volume_used_bytes| gluster_volume_used_bytes{volume="xxx"}  4857600	|
|gluster_volume_available_bytes|	gluster_volume_available_bytes{volume="xxx"} 9856780	|
|gluster_vol_num_bricks |  gluster_vol_num_bricks{vol="xxx"} 23 | number of bricks per volume

#### Brick Metrics:
|Gluster Promethus schema| example
|--------|-----------|
| gluster_brick_inodes_used |	gluster_brick_inodes_used {volume="xxx", brick_path="/mnt/xxx", peerid="xxxxx", host="xxxx"} 100	|
gluster_brick_inodes_free | gluster_brick_inodes_free {volume="xxx", brick_path="/mnt/xxx", peerid="xxxxx", host="xxxx"} 100	|
| gluster_brick_inodes_total |	gluster_brick_inodes_total{volume="xxx", brick_path="/mnt/xxx", peerid="xxxxx", host="xxx"} 200 |
| gluster_brick_inodes_available | gluster_brick_inodes_available{volume="xxx", brick_path="/mnt/xxx", peerid="xxxxx", host="xxx"} 100 |
| gluster_brick_capacity_total | gluster_brick_capacity_total{volume="xxx", brick_path="xxxx", peerid="xxx", host="xxx"} 1048576	|
| gluster_brick_capacity_used|	gluster_brick_capacity_used{volume="xxx", brick_path="xxxx", peerid="xxx", host="xxx"} 1048	|
| gluster_brick_capacity_free |	gluster_brick_capacity_free{volume="xxx", brick_path="xxx", peerid="xxx", host="xxx"} 1047576	|
 | gluster_disk_read_ios | gluster_disk_read_ios{brick_path="xxx",device="xxx",disk="vda",host="xxx",id="xxx",instance="xxx:8080",subvolume="xxx",volume="xxx"} 0
 |   gluster_disk_read_merges | gluster_disk_read_merges{brick_path="xxx",device="xxx",disk="vda",host="xxx",id="xxx",instance="xxx:8080",subvolume="xxx",volume="xxx"} 0
 |   gluster_disk_read_sectors | gluster_disk_read_sectors{brick_path="xxx",device="xxx",disk="vda",host="xxx",id="xxx",instance="xxx:8080",subvolume="xxx",volume="xxx"} 0
 |   gluster_disk_read_ticks | gluster_disk_read_ticks{brick_path="xxx",device="xxx",disk="vda",host="xxx",id="xxx",instance="xxx:8080",subvolume="xxx",volume="xxx"} 0
 |   gluster_disk_write_ios | gluster_disk_write_ios{brick_path="xxx",device="xxx",disk="vda",host="xxx",id="xxx",instance="xxx:8080",subvolume="xxx",volume="xxx"} 0
 |   gluster_disk_write_merges | gluster_disk_write_merges{brick_path="xxx",device="xxx",disk="vda",host="xxx",id="xxx",instance="xxx:8080",subvolume="xxx",volume="xxx"} 0 |
 |   gluster_disk_write_sectors | gluster_disk_write_sectors {brick_path="xxx",device="xxx",disk="vda",host="xxx",id="xxx",instance="xxx:8080",subvolume="xxx",volume="xxx"} 0 |
 |   gluster_disk_write_ticks | gluster_disk_write_ticks {brick_path="xxx",device="xxx",disk="vda",host="xxx",id="xxx",instance="xxx:8080",subvolume="xxx",volume="xxx"} 0 |
 |   gluster_disk_inflight | gluster_disk_inflight{brick_path="xxx",device="xxx",disk="vda",host="xxx",id="xxx",instance="xxx:8080",subvolume="xxx",volume="xxx"} 0 |
 |   gluster_disk_total_ticks | gluster_disk_total_ticks{brick_path="xxx", device="xxx",disk="vda",host="xxx",id="xxx",instance="xxx:8080",subvolume="xxx",volume="xxx"} 0 |
  |  gluster_disk_time_inqueue | gluster_disk_time_inqueue{brick_path="xxx",device="xxx",disk="vda",host="xxx",id="xxx",instance="xxx:8080",subvolume="xxx",volume="xxx"} 0 |

#### Gluster process Metrics: 

|Gluster Prometheus schema |example| comments |
|--------|-----------|------|
| gluster_virtual_memory | gluster_virtual_memory{name="glusterfsd",nodeid="xxx",brick_path="/bricks/b1",volume="gv1", host="xxx"} 912260
| gluster_resident_memory | gluster_resident_memory{name="glusterfsd",nodeid="xxx",brick_path="/bricks/b1",volume="gv1", host="xxx"} 15392
| gluster_memory_percentage | gluster_memory_percentage{name="glusterfsd",peerid="xxx",brick_path="/bricks/b1",volume="gv1", host="xxx"} 0.7
| gluster_cpu_percentage | gluster_cpu_percentage{name="glusterfsd",peerid="xxx",brick_path="/bricks/b2",volume="gv1", host="xxx"} 0
|gluster_elapsed_time_seconds | gluster_elapsed_time_seconds{name="glusterfsd",peerid="xxx",brick_path="/bricks/b1",volume="gv1", host="xxx"} 2969 | for name="glusterd" brick_path label will not be there |
