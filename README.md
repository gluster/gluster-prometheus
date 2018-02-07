# Prometheus exporter for Gluster Metrics

These exporters will be run on all Gluster nodes, So it makes sense to
collect only local metrics and aggregate in Prometheus server when
required.

Run `gluster_exporter` with Gluster Node ID(Current limitation, this
will be detected in future releases)

    ./gluster_exporter -nodeid <gluster-node-id>

For example,

    ./gluster_exporter -nodeid 019042a8-fc13-4abe-88b4-f070905bf78b

Other available options,

    -metrics-path string
          Metrics API Path (default "/metrics")
    -nodeid string
          Gluster Node ID
    -port int
          Exporter Port (default 8080)

In this initial version, following metrics are available

- CPU usage Percentage of Gluster Processes
- Elapsed Time/Uptime of Gluster processes
- Memory usage percentage of Gluster Processes
- Resident Memory of Gluster processes
- Virtual Memory of Gluster processes

Example output:

    # HELP gluster_cpu_percentage CPU Percentage used by Gluster processes
    # TYPE gluster_cpu_percentage gauge
    gluster_cpu_percentage{name="glusterfsd",nodeid="019042a8-fc13-4abe-88b4-f070905bf78b",brick_path="/bricks/b1",volume="gv1"} 0
    gluster_cpu_percentage{name="glusterfsd",nodeid="019042a8-fc13-4abe-88b4-f070905bf78b",brick_path="/bricks/b2",volume="gv1"} 0
    # HELP gluster_elapsed_time_seconds Elapsed Time of Gluster processes
    # TYPE gluster_elapsed_time_seconds gauge
    gluster_elapsed_time_seconds{name="glusterfsd",nodeid="019042a8-fc13-4abe-88b4-f070905bf78b",brick_path="/bricks/b1",volume="gv1"} 2969
    gluster_elapsed_time_seconds{name="glusterfsd",nodeid="019042a8-fc13-4abe-88b4-f070905bf78b",brick_path="/bricks/b2",volume="gv1"} 2969
    # HELP gluster_memory_percentage Memory Percentage used by Gluster processes
    # TYPE gluster_memory_percentage gauge
    gluster_memory_percentage{name="glusterfsd",nodeid="019042a8-fc13-4abe-88b4-f070905bf78b",brick_path="/bricks/b1",volume="gv1"} 0.7
    gluster_memory_percentage{name="glusterfsd",nodeid="019042a8-fc13-4abe-88b4-f070905bf78b",brick_path="/bricks/b2",volume="gv1"} 0.7
    # HELP gluster_resident_memory Resident Memory of Gluster processes
    # TYPE gluster_resident_memory gauge
    gluster_resident_memory{name="glusterfsd",nodeid="019042a8-fc13-4abe-88b4-f070905bf78b",brick_path="/bricks/b1",volume="gv1"} 15392
    gluster_resident_memory{name="glusterfsd",nodeid="019042a8-fc13-4abe-88b4-f070905bf78b",brick_path="/bricks/b2",volume="gv1"} 14760
    # HELP gluster_virtual_memory Virtual Memory of Gluster processes
    # TYPE gluster_virtual_memory gauge
    gluster_virtual_memory{name="glusterfsd",nodeid="019042a8-fc13-4abe-88b4-f070905bf78b",brick_path="/bricks/b1",volume="gv1"} 912260
    gluster_virtual_memory{name="glusterfsd",nodeid="019042a8-fc13-4abe-88b4-f070905bf78b",brick_path="/bricks/b2",volume="gv1"} 912520

Exporter will automatically detect the running processes which are
related to Gluster, and extracts the meta informations like Volume
name, Brick path etc from the `/proc/<pid>/cmdline`. These meta
information will be available as labels in Prometheus. Currently it is
detecting `glusterd`, `glusterfsd` and `glusterd2` processes only.
Support will be added to more processes like Geo-replication, Self
Heal etc.

## More Metrics
We are working on the following metrics,

- Utilization
- Profile Information
- Cluster/Volume/Bricks Health information
- Disk stats
- Many more

As mentioned before, exporter will collect only local metrics and to
be aggregated in Prometheus server. For example, aggregation rule for
utilization of a Distributed Replicated Gluster volume is

    # b1 and b2 are one replica pair, b3 and b4 are another replica pair
    Bricks = b1 b2 b3 b4

    total_free = min(free_b1, free_b2) + min(free_b3, free_b4)
    total = min(total_b1, total_b2) + min(total_b3, total_b4)

Writing rules will become very complex as number of bricks/nodes grows
and this formula differs for each Gluster Volume types. To address
this issue, `gluster_exporter` will also exports the rules file. Fetch
the rules file by reaching any one of the Gluster node and reload
Prometheus to get latest rules.

    # Get latest rules and send signal to running Prometheus to pickup latest rules
    wget <gluster-node-ip>:8080/rules -O ~/prometheus/gluster-rules.yml

Note: This is not yet implemented, but will be available soon

## Adding New metrics

- Add new file under `gluster_exporter` directory.
- Define Metrics depending on the type of
  Metric(https://prometheus.io/docs/concepts/metric_types/)
  For example, "Gauge" Metrics type

        glusterCPUPercentage = prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace: "gluster",
                Name:      "cpu_percentage",
                Help:      "CPU Percentage used by Gluster processes",
            },
            []string{"volume", "nodeid", "brick_path"},
        )

- Implement the function to gather data, and register to gather data
  in required interval

        prometheus.MustRegister(glusterCPUPercentage)

        // Register to update this every 2 seconds
        // Name, Callback Func, Interval Seconds
        registerMetric("gluster_ps", psfunc, 2)

- Thats it! Exporter will run these registered metrics.

## TODO
- Build and Install Makefile
- More metrics need to be implemented
- Logging
- Handling failures
- Rules generation
- Consuming Glusterd2 REST APIs
- Detect Gluster Node ID automatically(Now accepts as Commandline
  parameter)
- Tutorial for setup, integration with Grafana etc
