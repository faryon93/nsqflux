# nsqflux

## Payload Formats
```json
{"_measurement":  "my-sensors", "_timestamp":  123456789, "tag":  "tag-value", "value":  555}
```
The field `_measurement` overrides the configured default influxdb measurment name.
If no `_timestamp` field (unix time in milliseconds) is present, the server assigns the timestamp on reception.

Numerical values are stored as value fields and string values are stored as measurement tags. 
To override this behaviour, one can prefix the value name with the `$` character in order to treat strings as values.

If an array of objects is provided, each element is inserted as a single datapoint. 
```json
[
  {"_measurement":  "my-sensors", "_timestamp":  123456, "tag":  "tag", "value":  123.2, "value2": 12},
  {"_measurement":  "my-sensors", "_timestamp":  684658, "tag":  "tag1", "value3":  555}
]
```
