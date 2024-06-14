# openstack_exporter
Prometheus exporter to retrieve OpenStack metrics about limits and status

## Usage

```
usage: main [<flags>]


Flags:
  -h, --[no-]help          Show context-sensitive help (also try --help-long and --help-man).
      --volume.limit=-1    Max number of volumes when on OTC
      --log.level=info     Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt  Output format of log messages. One of: [logfmt, json]
```

The `--volume.limit` is only used when running the exporter on OTC, because we currently have no way of getting the limits via the API.

### Authentication

You should authenticate by using environment variables.
For OTC, make sure following environment variables are set:

    * OS_AUTH_URL
    * OS_DOMAIN_NAME
    * OS_PASSWORD
    * OS_USERNAME
    * OS_PROJECT_ID
    * OS_ACCESS_KEY
    * OS_SECRET_KEY

For CloudFerro, following environment variables are needed:

    * OS_AUTH_URL
    * OS_USERNAME
    * OS_REGION_NAME
    * OS_PROJECT_ID
    * OS_PASSWORD
    * OS_DOMAIN_ID

### Running the exporter via Podman

```bash
podman build -t registry.example.com/openstack_exporter .
podman run -p 9595:9595 <ALL_NECESSARY_ENV_VARS> registry.example.com/openstack_exporter:latest <args>

curl http://localhost:9595/metrics
```

## Exposed metrics

| Metric                               | Description                                                         |
|--------------------------------------|---------------------------------------------------------------------|
| openstack_collect_duration_seconds   | The time it took to collect the metrics in seconds                  |
| openstack_container_bytes_used       | The total of bytes stored in the container                          |
| openstack_max_total_cores            | The limit of cores that can be assigned to instances in the project |
| openstack_max_total_instances        | The limit of total instances in the project                         |
| openstack_max_total_volumes          | The limit of total volumes in the project                           |
| openstack_max_total_volume_gigabytes | The limit of total volume size in the project                       |
| openstack_max_total_ram_size         | The limit of RAM that can be assigned to instances in the project   |
| openstack_max_total_volumes          | The limit of total volumes in the project                           |
| openstack_per_flavor_instance_count  | Number of instances per flavor                                      |
| openstack_per_status_instance_count  | Number of instances per status                                      |
| openstack_per_status_volume_count    | Number of volumes per status                                        |
| openstack_total_cores_used           | The current number of cores used                                    |
| openstack_total_instances_used       | The current number of instances                                     |
| openstack_total_ram_used             | The current number RAM used                                         |
| openstack_total_volumes_used         | The current number of volumes                                       |
