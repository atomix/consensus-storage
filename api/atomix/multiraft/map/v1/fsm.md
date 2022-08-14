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
    - [Event.Inserted](#atomix-multiraft-map-v1-Event-Inserted)
    - [Event.Removed](#atomix-multiraft-map-v1-Event-Removed)
    - [Event.Updated](#atomix-multiraft-map-v1-Event-Updated)
    - [EventsInput](#atomix-multiraft-map-v1-EventsInput)
    - [EventsOutput](#atomix-multiraft-map-v1-EventsOutput)
    - [GetInput](#atomix-multiraft-map-v1-GetInput)
    - [GetOutput](#atomix-multiraft-map-v1-GetOutput)
    - [MapEntry](#atomix-multiraft-map-v1-MapEntry)
    - [MapInput](#atomix-multiraft-map-v1-MapInput)
    - [MapListener](#atomix-multiraft-map-v1-MapListener)
    - [MapOutput](#atomix-multiraft-map-v1-MapOutput)
    - [MapValue](#atomix-multiraft-map-v1-MapValue)
    - [PutInput](#atomix-multiraft-map-v1-PutInput)
    - [PutOutput](#atomix-multiraft-map-v1-PutOutput)
    - [RemoveInput](#atomix-multiraft-map-v1-RemoveInput)
    - [RemoveOutput](#atomix-multiraft-map-v1-RemoveOutput)
    - [SizeInput](#atomix-multiraft-map-v1-SizeInput)
    - [SizeOutput](#atomix-multiraft-map-v1-SizeOutput)
    - [Value](#atomix-multiraft-map-v1-Value)
  
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



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| watch | [bool](#bool) |  |  |






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






<a name="atomix-multiraft-map-v1-Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| inserted | [Event.Inserted](#atomix-multiraft-map-v1-Event-Inserted) |  |  |
| updated | [Event.Updated](#atomix-multiraft-map-v1-Event-Updated) |  |  |
| removed | [Event.Removed](#atomix-multiraft-map-v1-Event-Removed) |  |  |






<a name="atomix-multiraft-map-v1-Event-Inserted"></a>

### Event.Inserted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-map-v1-Value) |  |  |






<a name="atomix-multiraft-map-v1-Event-Removed"></a>

### Event.Removed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-map-v1-Value) |  |  |
| expired | [bool](#bool) |  |  |






<a name="atomix-multiraft-map-v1-Event-Updated"></a>

### Event.Updated



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| new_value | [Value](#atomix-multiraft-map-v1-Value) |  |  |
| prev_value | [Value](#atomix-multiraft-map-v1-Value) |  |  |






<a name="atomix-multiraft-map-v1-EventsInput"></a>

### EventsInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |






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
| value | [Value](#atomix-multiraft-map-v1-Value) |  |  |






<a name="atomix-multiraft-map-v1-MapEntry"></a>

### MapEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [MapValue](#atomix-multiraft-map-v1-MapValue) |  |  |






<a name="atomix-multiraft-map-v1-MapInput"></a>

### MapInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [SizeInput](#atomix-multiraft-map-v1-SizeInput) |  |  |
| put | [PutInput](#atomix-multiraft-map-v1-PutInput) |  |  |
| get | [GetInput](#atomix-multiraft-map-v1-GetInput) |  |  |
| remove | [RemoveInput](#atomix-multiraft-map-v1-RemoveInput) |  |  |
| clear | [ClearInput](#atomix-multiraft-map-v1-ClearInput) |  |  |
| entries | [EntriesInput](#atomix-multiraft-map-v1-EntriesInput) |  |  |
| events | [EventsInput](#atomix-multiraft-map-v1-EventsInput) |  |  |






<a name="atomix-multiraft-map-v1-MapListener"></a>

### MapListener



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| key | [string](#string) |  |  |






<a name="atomix-multiraft-map-v1-MapOutput"></a>

### MapOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [SizeOutput](#atomix-multiraft-map-v1-SizeOutput) |  |  |
| put | [PutOutput](#atomix-multiraft-map-v1-PutOutput) |  |  |
| get | [GetOutput](#atomix-multiraft-map-v1-GetOutput) |  |  |
| remove | [RemoveOutput](#atomix-multiraft-map-v1-RemoveOutput) |  |  |
| clear | [ClearOutput](#atomix-multiraft-map-v1-ClearOutput) |  |  |
| entries | [EntriesOutput](#atomix-multiraft-map-v1-EntriesOutput) |  |  |
| events | [EventsOutput](#atomix-multiraft-map-v1-EventsOutput) |  |  |






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
| entry | [Entry](#atomix-multiraft-map-v1-Entry) |  |  |






<a name="atomix-multiraft-map-v1-PutOutput"></a>

### PutOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| prev_value | [Value](#atomix-multiraft-map-v1-Value) |  |  |






<a name="atomix-multiraft-map-v1-RemoveInput"></a>

### RemoveInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |






<a name="atomix-multiraft-map-v1-RemoveOutput"></a>

### RemoveOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-map-v1-Value) |  |  |






<a name="atomix-multiraft-map-v1-SizeInput"></a>

### SizeInput







<a name="atomix-multiraft-map-v1-SizeOutput"></a>

### SizeOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [uint32](#uint32) |  |  |






<a name="atomix-multiraft-map-v1-Value"></a>

### Value



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |





 

 

 

 



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

