# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/map/v1/fsm.proto](#atomix_multiraft_map_v1_fsm-proto)
    - [ClearInput](#atomix-multiraft-map-v1-ClearInput)
    - [ClearOutput](#atomix-multiraft-map-v1-ClearOutput)
    - [EntriesInput](#atomix-multiraft-map-v1-EntriesInput)
    - [EntriesOutput](#atomix-multiraft-map-v1-EntriesOutput)
    - [Entry](#atomix-multiraft-map-v1-Entry)
    - [Event](#atomix-multiraft-map-v1-Event)
    - [EventsInput](#atomix-multiraft-map-v1-EventsInput)
    - [EventsOutput](#atomix-multiraft-map-v1-EventsOutput)
    - [GetInput](#atomix-multiraft-map-v1-GetInput)
    - [GetOutput](#atomix-multiraft-map-v1-GetOutput)
    - [InsertInput](#atomix-multiraft-map-v1-InsertInput)
    - [InsertOutput](#atomix-multiraft-map-v1-InsertOutput)
    - [MapEntry](#atomix-multiraft-map-v1-MapEntry)
    - [MapKey](#atomix-multiraft-map-v1-MapKey)
    - [MapListener](#atomix-multiraft-map-v1-MapListener)
    - [MapSnapshot](#atomix-multiraft-map-v1-MapSnapshot)
    - [MapValue](#atomix-multiraft-map-v1-MapValue)
    - [PutInput](#atomix-multiraft-map-v1-PutInput)
    - [PutOutput](#atomix-multiraft-map-v1-PutOutput)
    - [RemoveInput](#atomix-multiraft-map-v1-RemoveInput)
    - [RemoveOutput](#atomix-multiraft-map-v1-RemoveOutput)
    - [SizeInput](#atomix-multiraft-map-v1-SizeInput)
    - [SizeOutput](#atomix-multiraft-map-v1-SizeOutput)
    - [UpdateInput](#atomix-multiraft-map-v1-UpdateInput)
    - [UpdateOutput](#atomix-multiraft-map-v1-UpdateOutput)
    - [Value](#atomix-multiraft-map-v1-Value)
  
    - [Event.Type](#atomix-multiraft-map-v1-Event-Type)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_map_v1_fsm-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/map/v1/fsm.proto



<a name="atomix-multiraft-map-v1-ClearInput"></a>

### ClearInput







<a name="atomix-multiraft-map-v1-ClearOutput"></a>

### ClearOutput







<a name="atomix-multiraft-map-v1-EntriesInput"></a>

### EntriesInput







<a name="atomix-multiraft-map-v1-EntriesOutput"></a>

### EntriesOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-map-v1-Entry) |  |  |






<a name="atomix-multiraft-map-v1-Entry"></a>

### Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Value](#atomix-multiraft-map-v1-Value) |  |  |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-map-v1-Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [Event.Type](#atomix-multiraft-map-v1-Event-Type) |  |  |
| entry | [Entry](#atomix-multiraft-map-v1-Entry) |  |  |






<a name="atomix-multiraft-map-v1-EventsInput"></a>

### EventsInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| replay | [bool](#bool) |  |  |






<a name="atomix-multiraft-map-v1-EventsOutput"></a>

### EventsOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#atomix-multiraft-map-v1-Event) |  |  |






<a name="atomix-multiraft-map-v1-GetInput"></a>

### GetInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |






<a name="atomix-multiraft-map-v1-GetOutput"></a>

### GetOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-map-v1-Entry) |  |  |






<a name="atomix-multiraft-map-v1-InsertInput"></a>

### InsertInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Value](#atomix-multiraft-map-v1-Value) |  |  |






<a name="atomix-multiraft-map-v1-InsertOutput"></a>

### InsertOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-map-v1-Entry) |  |  |






<a name="atomix-multiraft-map-v1-MapEntry"></a>

### MapEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [MapKey](#atomix-multiraft-map-v1-MapKey) |  |  |
| value | [MapValue](#atomix-multiraft-map-v1-MapValue) |  |  |






<a name="atomix-multiraft-map-v1-MapKey"></a>

### MapKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| key | [string](#string) |  |  |






<a name="atomix-multiraft-map-v1-MapListener"></a>

### MapListener



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| key | [string](#string) |  |  |






<a name="atomix-multiraft-map-v1-MapSnapshot"></a>

### MapSnapshot



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| listeners | [MapListener](#atomix-multiraft-map-v1-MapListener) | repeated |  |
| entries | [MapEntry](#atomix-multiraft-map-v1-MapEntry) | repeated |  |






<a name="atomix-multiraft-map-v1-MapValue"></a>

### MapValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| expire | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="atomix-multiraft-map-v1-PutInput"></a>

### PutInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Value](#atomix-multiraft-map-v1-Value) |  |  |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-map-v1-PutOutput"></a>

### PutOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-map-v1-Entry) |  |  |






<a name="atomix-multiraft-map-v1-RemoveInput"></a>

### RemoveInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-map-v1-RemoveOutput"></a>

### RemoveOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-map-v1-Entry) |  |  |






<a name="atomix-multiraft-map-v1-SizeInput"></a>

### SizeInput







<a name="atomix-multiraft-map-v1-SizeOutput"></a>

### SizeOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [uint32](#uint32) |  |  |






<a name="atomix-multiraft-map-v1-UpdateInput"></a>

### UpdateInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Value](#atomix-multiraft-map-v1-Value) |  |  |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-map-v1-UpdateOutput"></a>

### UpdateOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-map-v1-Entry) |  |  |






<a name="atomix-multiraft-map-v1-Value"></a>

### Value



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |





 


<a name="atomix-multiraft-map-v1-Event-Type"></a>

### Event.Type


| Name | Number | Description |
| ---- | ------ | ----------- |
| NONE | 0 |  |
| INSERT | 1 |  |
| UPDATE | 2 |  |
| REMOVE | 3 |  |
| REPLAY | 4 |  |


 

 

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

