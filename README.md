mqlux
=====

mqlux forwards messages from MQTT into InfluxDB. It can be used to archive and visualize sensor data in combination with Grafana.

Some buzzwords:

- Flexible configuration
- Powerful regex-based topic matching
- Support for custom payload parsers (JavaScript)


Configuration
=============

mqlux can subscribe to one or more MQTT topics. It inserts each MQTT message as a record into the configured InfluxDB database. You can configure the measurement name and tags (optional) for each topic. The value is always stored inside the `value` field. [Refer to the InfluxDB documentation about measurement and tags and field concepts][1].

[1]: https://docs.influxdata.com/influxdb/v1.4/concepts/key_concepts/

Please read `mqlux.tml` for more *"documentation"* of the configuration format.

Simple float values
-------------------

```
[[subscription]]
topic = "/sensors/kitchen/temperature"
measurement = "temperature"
[subscription.tags]
sensor = "dht22"
location = "kitchen"
```

Regexp topic
------------

You can use regular expressions in topics to match multiple topics.
Named capturing groups are automatically converted to InfluxDB tags.
This can be used to simplify the configuration if you have multiple, similar sensors.

For example, if you want to collect statistics from each port of multiple switches:

```
[[subscription]]
topic = "/net/switch/(?P<location>[a-z_-]+)/port-(?P<port>\\d+)/stats/tx-bytes"
measurement = "tx_bytes"
[subscription.tags]
device = "switch"
```

All matching MQTT messages are recorded as a `tx_bytes` measurement.
The tag `location` and `port` is extracted from the topic. For example, messages to `/net/switch/rack-a/port-42/stats/tx-bytes` will have the `location=rack-a`, `port=42` and `device=switch` tags set.

[Refer to the Go documentation for the supported regxep syntax][1].
Note that the TOML configuration format requires that backslashes are escaped.

[1]: https://golang.org/pkg/regexp/syntax/


JavaScript based message parser
-------------------------------

mqlux embeds a JavaScript interpreter that can be used write your own parser.
You can use this when the payload of your MQTT message does not contain simple strings of float numbers.

Your script needs to define a `parse` function that takes the topic and payload. The function can return:

- a simple value (float/integer, boolean, string)
- an object with `value`, `measurement` and `tags`
- an array of multiple objects with `value`, `measurement` and `tags`

mqlux will use the `measurement` name if your function provides one, otherwise the `measurement` name from the `subscription` configuration is used.

`tags` from the `subscription` configuration and the retured `tags` are combined. Returned `tags` take precedence. 

Your `parse` function is called for each MQTT message. Each subscription gets its own independent JavaScript VM that is reused for each message, so that you can use global variables to keep a state (e.g. count the number of messages).

The following parses JSON payload and returns three independent measurements:

```
[[subscription]]
topic = "/net/devices"
script = """
function parse(topic, payload) { 
    var data = JSON.parse(payload);
    return [
        {"measurement": "people", "value": data["peopleCount"]},
        {"measurement": "devices_total", "value": data["deviceCount"]},
        {"measurement": "devices_unknown", "value": data["unknownDevicesCount"]}
    ];
}
"""
```