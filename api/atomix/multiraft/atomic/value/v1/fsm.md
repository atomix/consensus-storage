# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/atomic/value/v1/fsm.proto](#atomix_multiraft_atomic_value_v1_fsm-proto)
    - [AtomicValueInput](#atomix-multiraft-atomic-value-v1-AtomicValueInput)
    - [AtomicValueOutput](#atomix-multiraft-atomic-value-v1-AtomicValueOutput)
    - [AtomicValueState](#atomix-multiraft-atomic-value-v1-AtomicValueState)
    - [DeleteInput](#atomix-multiraft-atomic-value-v1-DeleteInput)
    - [DeleteOutput](#atomix-multiraft-atomic-value-v1-DeleteOutput)
    - [Event](#atomix-multiraft-atomic-value-v1-Event)
    - [Event.Created](#atomix-multiraft-atomic-value-v1-Event-Created)
    - [Event.Deleted](#atomix-multiraft-atomic-value-v1-Event-Deleted)
    - [Event.Updated](#atomix-multiraft-atomic-value-v1-Event-Updated)
    - [EventsInput](#atomix-multiraft-atomic-value-v1-EventsInput)
    - [EventsOutput](#atomix-multiraft-atomic-value-v1-EventsOutput)
    - [GetInput](#atomix-multiraft-atomic-value-v1-GetInput)
    - [GetOutput](#atomix-multiraft-atomic-value-v1-GetOutput)
    - [SetInput](#atomix-multiraft-atomic-value-v1-SetInput)
    - [SetOutput](#atomix-multiraft-atomic-value-v1-SetOutput)
    - [UpdateInput](#atomix-multiraft-atomic-value-v1-UpdateInput)
    - [UpdateOutput](#atomix-multiraft-atomic-value-v1-UpdateOutput)
    - [Value](#atomix-multiraft-atomic-value-v1-Value)
    - [WatchInput](#atomix-multiraft-atomic-value-v1-WatchInput)
    - [WatchOutput](#atomix-multiraft-atomic-value-v1-WatchOutput)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_atomic_value_v1_fsm-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/atomic/value/v1/fsm.proto



<a name="atomix-multiraft-atomic-value-v1-AtomicValueInput"></a>

### AtomicValueInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| get | [GetInput](#atomix-multiraft-atomic-value-v1-GetInput) |  |  |
| set | [SetInput](#atomix-multiraft-atomic-value-v1-SetInput) |  |  |
| update | [UpdateInput](#atomix-multiraft-atomic-value-v1-UpdateInput) |  |  |
| delete | [DeleteInput](#atomix-multiraft-atomic-value-v1-DeleteInput) |  |  |
| watch | [WatchInput](#atomix-multiraft-atomic-value-v1-WatchInput) |  |  |
| events | [EventsInput](#atomix-multiraft-atomic-value-v1-EventsInput) |  |  |






<a name="atomix-multiraft-atomic-value-v1-AtomicValueOutput"></a>

### AtomicValueOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| get | [GetOutput](#atomix-multiraft-atomic-value-v1-GetOutput) |  |  |
| set | [SetOutput](#atomix-multiraft-atomic-value-v1-SetOutput) |  |  |
| update | [UpdateOutput](#atomix-multiraft-atomic-value-v1-UpdateOutput) |  |  |
| delete | [DeleteOutput](#atomix-multiraft-atomic-value-v1-DeleteOutput) |  |  |
| watch | [WatchOutput](#atomix-multiraft-atomic-value-v1-WatchOutput) |  |  |
| events | [EventsOutput](#atomix-multiraft-atomic-value-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-atomic-value-v1-AtomicValueState"></a>

### AtomicValueState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-atomic-value-v1-Value) |  |  |
| expire | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="atomix-multiraft-atomic-value-v1-DeleteInput"></a>

### DeleteInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| prev_index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-atomic-value-v1-DeleteOutput"></a>

### DeleteOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-atomic-value-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-value-v1-Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| created | [Event.Created](#atomix-multiraft-atomic-value-v1-Event-Created) |  |  |
| updated | [Event.Updated](#atomix-multiraft-atomic-value-v1-Event-Updated) |  |  |
| deleted | [Event.Deleted](#atomix-multiraft-atomic-value-v1-Event-Deleted) |  |  |






<a name="atomix-multiraft-atomic-value-v1-Event-Created"></a>

### Event.Created



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-atomic-value-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-value-v1-Event-Deleted"></a>

### Event.Deleted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-atomic-value-v1-Value) |  |  |
| expired | [bool](#bool) |  |  |






<a name="atomix-multiraft-atomic-value-v1-Event-Updated"></a>

### Event.Updated



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-atomic-value-v1-Value) |  |  |
| prev_value | [Value](#atomix-multiraft-atomic-value-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-value-v1-EventsInput"></a>

### EventsInput







<a name="atomix-multiraft-atomic-value-v1-EventsOutput"></a>

### EventsOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#atomix-multiraft-atomic-value-v1-Event) |  |  |






<a name="atomix-multiraft-atomic-value-v1-GetInput"></a>

### GetInput







<a name="atomix-multiraft-atomic-value-v1-GetOutput"></a>

### GetOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-atomic-value-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-value-v1-SetInput"></a>

### SetInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="atomix-multiraft-atomic-value-v1-SetOutput"></a>

### SetOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-atomic-value-v1-UpdateInput"></a>

### UpdateInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| prev_index | [uint64](#uint64) |  |  |
| ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="atomix-multiraft-atomic-value-v1-UpdateOutput"></a>

### UpdateOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| prev_value | [Value](#atomix-multiraft-atomic-value-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-value-v1-Value"></a>

### Value



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-atomic-value-v1-WatchInput"></a>

### WatchInput







<a name="atomix-multiraft-atomic-value-v1-WatchOutput"></a>

### WatchOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-atomic-value-v1-Value) |  |  |





 

 

 

 



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

