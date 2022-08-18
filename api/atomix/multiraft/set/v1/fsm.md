# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/set/v1/fsm.proto](#atomix_multiraft_set_v1_fsm-proto)
    - [AddInput](#atomix-multiraft-set-v1-AddInput)
    - [AddOutput](#atomix-multiraft-set-v1-AddOutput)
    - [ClearInput](#atomix-multiraft-set-v1-ClearInput)
    - [ClearOutput](#atomix-multiraft-set-v1-ClearOutput)
    - [ContainsInput](#atomix-multiraft-set-v1-ContainsInput)
    - [ContainsOutput](#atomix-multiraft-set-v1-ContainsOutput)
    - [Element](#atomix-multiraft-set-v1-Element)
    - [ElementsInput](#atomix-multiraft-set-v1-ElementsInput)
    - [ElementsOutput](#atomix-multiraft-set-v1-ElementsOutput)
    - [Event](#atomix-multiraft-set-v1-Event)
    - [Event.Added](#atomix-multiraft-set-v1-Event-Added)
    - [Event.Removed](#atomix-multiraft-set-v1-Event-Removed)
    - [EventsInput](#atomix-multiraft-set-v1-EventsInput)
    - [EventsOutput](#atomix-multiraft-set-v1-EventsOutput)
    - [RemoveInput](#atomix-multiraft-set-v1-RemoveInput)
    - [RemoveOutput](#atomix-multiraft-set-v1-RemoveOutput)
    - [SetElement](#atomix-multiraft-set-v1-SetElement)
    - [SetInput](#atomix-multiraft-set-v1-SetInput)
    - [SetOutput](#atomix-multiraft-set-v1-SetOutput)
    - [SizeInput](#atomix-multiraft-set-v1-SizeInput)
    - [SizeOutput](#atomix-multiraft-set-v1-SizeOutput)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_set_v1_fsm-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/set/v1/fsm.proto



<a name="atomix-multiraft-set-v1-AddInput"></a>

### AddInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| element | [Element](#atomix-multiraft-set-v1-Element) |  |  |
| ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="atomix-multiraft-set-v1-AddOutput"></a>

### AddOutput







<a name="atomix-multiraft-set-v1-ClearInput"></a>

### ClearInput







<a name="atomix-multiraft-set-v1-ClearOutput"></a>

### ClearOutput







<a name="atomix-multiraft-set-v1-ContainsInput"></a>

### ContainsInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| element | [Element](#atomix-multiraft-set-v1-Element) |  |  |






<a name="atomix-multiraft-set-v1-ContainsOutput"></a>

### ContainsOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| contains | [bool](#bool) |  |  |






<a name="atomix-multiraft-set-v1-Element"></a>

### Element



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [string](#string) |  |  |






<a name="atomix-multiraft-set-v1-ElementsInput"></a>

### ElementsInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| watch | [bool](#bool) |  |  |






<a name="atomix-multiraft-set-v1-ElementsOutput"></a>

### ElementsOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| element | [Element](#atomix-multiraft-set-v1-Element) |  |  |






<a name="atomix-multiraft-set-v1-Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| added | [Event.Added](#atomix-multiraft-set-v1-Event-Added) |  |  |
| removed | [Event.Removed](#atomix-multiraft-set-v1-Event-Removed) |  |  |






<a name="atomix-multiraft-set-v1-Event-Added"></a>

### Event.Added



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| element | [Element](#atomix-multiraft-set-v1-Element) |  |  |






<a name="atomix-multiraft-set-v1-Event-Removed"></a>

### Event.Removed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| element | [Element](#atomix-multiraft-set-v1-Element) |  |  |
| expired | [bool](#bool) |  |  |






<a name="atomix-multiraft-set-v1-EventsInput"></a>

### EventsInput







<a name="atomix-multiraft-set-v1-EventsOutput"></a>

### EventsOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#atomix-multiraft-set-v1-Event) |  |  |






<a name="atomix-multiraft-set-v1-RemoveInput"></a>

### RemoveInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| element | [Element](#atomix-multiraft-set-v1-Element) |  |  |






<a name="atomix-multiraft-set-v1-RemoveOutput"></a>

### RemoveOutput







<a name="atomix-multiraft-set-v1-SetElement"></a>

### SetElement



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| expire | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="atomix-multiraft-set-v1-SetInput"></a>

### SetInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [SizeInput](#atomix-multiraft-set-v1-SizeInput) |  |  |
| contains | [ContainsInput](#atomix-multiraft-set-v1-ContainsInput) |  |  |
| add | [AddInput](#atomix-multiraft-set-v1-AddInput) |  |  |
| remove | [RemoveInput](#atomix-multiraft-set-v1-RemoveInput) |  |  |
| clear | [ClearInput](#atomix-multiraft-set-v1-ClearInput) |  |  |
| elements | [ElementsInput](#atomix-multiraft-set-v1-ElementsInput) |  |  |
| events | [EventsInput](#atomix-multiraft-set-v1-EventsInput) |  |  |






<a name="atomix-multiraft-set-v1-SetOutput"></a>

### SetOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [SizeOutput](#atomix-multiraft-set-v1-SizeOutput) |  |  |
| contains | [ContainsOutput](#atomix-multiraft-set-v1-ContainsOutput) |  |  |
| add | [AddOutput](#atomix-multiraft-set-v1-AddOutput) |  |  |
| remove | [RemoveOutput](#atomix-multiraft-set-v1-RemoveOutput) |  |  |
| clear | [ClearOutput](#atomix-multiraft-set-v1-ClearOutput) |  |  |
| elements | [ElementsOutput](#atomix-multiraft-set-v1-ElementsOutput) |  |  |
| events | [EventsOutput](#atomix-multiraft-set-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-set-v1-SizeInput"></a>

### SizeInput







<a name="atomix-multiraft-set-v1-SizeOutput"></a>

### SizeOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [uint32](#uint32) |  |  |





 

 

 

 



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

