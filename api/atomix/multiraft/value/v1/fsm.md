# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/value/v1/fsm.proto](#atomix_multiraft_value_v1_fsm-proto)
    - [DeleteInput](#atomix-multiraft-value-v1-DeleteInput)
    - [DeleteOutput](#atomix-multiraft-value-v1-DeleteOutput)
    - [Event](#atomix-multiraft-value-v1-Event)
    - [Event.Created](#atomix-multiraft-value-v1-Event-Created)
    - [Event.Deleted](#atomix-multiraft-value-v1-Event-Deleted)
    - [Event.Updated](#atomix-multiraft-value-v1-Event-Updated)
    - [EventsInput](#atomix-multiraft-value-v1-EventsInput)
    - [EventsOutput](#atomix-multiraft-value-v1-EventsOutput)
    - [GetInput](#atomix-multiraft-value-v1-GetInput)
    - [GetOutput](#atomix-multiraft-value-v1-GetOutput)
    - [IndexedValue](#atomix-multiraft-value-v1-IndexedValue)
    - [InsertInput](#atomix-multiraft-value-v1-InsertInput)
    - [InsertOutput](#atomix-multiraft-value-v1-InsertOutput)
    - [SetInput](#atomix-multiraft-value-v1-SetInput)
    - [SetOutput](#atomix-multiraft-value-v1-SetOutput)
    - [UpdateInput](#atomix-multiraft-value-v1-UpdateInput)
    - [UpdateOutput](#atomix-multiraft-value-v1-UpdateOutput)
    - [ValueInput](#atomix-multiraft-value-v1-ValueInput)
    - [ValueOutput](#atomix-multiraft-value-v1-ValueOutput)
    - [ValueState](#atomix-multiraft-value-v1-ValueState)
    - [WatchInput](#atomix-multiraft-value-v1-WatchInput)
    - [WatchOutput](#atomix-multiraft-value-v1-WatchOutput)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_value_v1_fsm-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/value/v1/fsm.proto



<a name="atomix-multiraft-value-v1-DeleteInput"></a>

### DeleteInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| prev_index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-value-v1-DeleteOutput"></a>

### DeleteOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [IndexedValue](#atomix-multiraft-value-v1-IndexedValue) |  |  |






<a name="atomix-multiraft-value-v1-Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| created | [Event.Created](#atomix-multiraft-value-v1-Event-Created) |  |  |
| updated | [Event.Updated](#atomix-multiraft-value-v1-Event-Updated) |  |  |
| deleted | [Event.Deleted](#atomix-multiraft-value-v1-Event-Deleted) |  |  |






<a name="atomix-multiraft-value-v1-Event-Created"></a>

### Event.Created



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [IndexedValue](#atomix-multiraft-value-v1-IndexedValue) |  |  |






<a name="atomix-multiraft-value-v1-Event-Deleted"></a>

### Event.Deleted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [IndexedValue](#atomix-multiraft-value-v1-IndexedValue) |  |  |
| expired | [bool](#bool) |  |  |






<a name="atomix-multiraft-value-v1-Event-Updated"></a>

### Event.Updated



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [IndexedValue](#atomix-multiraft-value-v1-IndexedValue) |  |  |
| prev_value | [IndexedValue](#atomix-multiraft-value-v1-IndexedValue) |  |  |






<a name="atomix-multiraft-value-v1-EventsInput"></a>

### EventsInput







<a name="atomix-multiraft-value-v1-EventsOutput"></a>

### EventsOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#atomix-multiraft-value-v1-Event) |  |  |






<a name="atomix-multiraft-value-v1-GetInput"></a>

### GetInput







<a name="atomix-multiraft-value-v1-GetOutput"></a>

### GetOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [IndexedValue](#atomix-multiraft-value-v1-IndexedValue) |  |  |






<a name="atomix-multiraft-value-v1-IndexedValue"></a>

### IndexedValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-value-v1-InsertInput"></a>

### InsertInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="atomix-multiraft-value-v1-InsertOutput"></a>

### InsertOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-value-v1-SetInput"></a>

### SetInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="atomix-multiraft-value-v1-SetOutput"></a>

### SetOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| prev_value | [IndexedValue](#atomix-multiraft-value-v1-IndexedValue) |  |  |






<a name="atomix-multiraft-value-v1-UpdateInput"></a>

### UpdateInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| prev_index | [uint64](#uint64) |  |  |
| ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="atomix-multiraft-value-v1-UpdateOutput"></a>

### UpdateOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| prev_value | [IndexedValue](#atomix-multiraft-value-v1-IndexedValue) |  |  |






<a name="atomix-multiraft-value-v1-ValueInput"></a>

### ValueInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| get | [GetInput](#atomix-multiraft-value-v1-GetInput) |  |  |
| set | [SetInput](#atomix-multiraft-value-v1-SetInput) |  |  |
| insert | [InsertInput](#atomix-multiraft-value-v1-InsertInput) |  |  |
| update | [UpdateInput](#atomix-multiraft-value-v1-UpdateInput) |  |  |
| delete | [DeleteInput](#atomix-multiraft-value-v1-DeleteInput) |  |  |
| watch | [WatchInput](#atomix-multiraft-value-v1-WatchInput) |  |  |
| events | [EventsInput](#atomix-multiraft-value-v1-EventsInput) |  |  |






<a name="atomix-multiraft-value-v1-ValueOutput"></a>

### ValueOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| get | [GetOutput](#atomix-multiraft-value-v1-GetOutput) |  |  |
| set | [SetOutput](#atomix-multiraft-value-v1-SetOutput) |  |  |
| insert | [InsertOutput](#atomix-multiraft-value-v1-InsertOutput) |  |  |
| update | [UpdateOutput](#atomix-multiraft-value-v1-UpdateOutput) |  |  |
| delete | [DeleteOutput](#atomix-multiraft-value-v1-DeleteOutput) |  |  |
| watch | [WatchOutput](#atomix-multiraft-value-v1-WatchOutput) |  |  |
| events | [EventsOutput](#atomix-multiraft-value-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-value-v1-ValueState"></a>

### ValueState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [IndexedValue](#atomix-multiraft-value-v1-IndexedValue) |  |  |
| expire | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="atomix-multiraft-value-v1-WatchInput"></a>

### WatchInput







<a name="atomix-multiraft-value-v1-WatchOutput"></a>

### WatchOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [IndexedValue](#atomix-multiraft-value-v1-IndexedValue) |  |  |





 

 

 

 



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

