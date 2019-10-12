# nsqflux

## Payload Formats
```json
{"_measurement":  "my-sensors", "_timestamp":  123456789, "tag":  "tag-value", "value":  555}
```
The field `_measurement` overrides the configured default influxdb measurment name.
If no `_timestamp` field (unix time in milliseconds) is presentm, the server assigns the timestamp on reception.

If an array of objects is provided, each element is inserted as a single datapoint. 
```json
[
  {"_measurement":  "my-sensors", "_timestamp":  123456, "tag":  "tag", "value":  123.2, "value2": 12},
  {"_measurement":  "my-sensors", "_timestamp":  684658, "tag":  "tag1", "value3":  555}
]
```
