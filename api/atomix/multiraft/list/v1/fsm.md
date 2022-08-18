# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/list/v1/fsm.proto](#atomix_multiraft_list_v1_fsm-proto)
    - [AppendInput](#atomix-multiraft-list-v1-AppendInput)
    - [AppendOutput](#atomix-multiraft-list-v1-AppendOutput)
    - [ClearInput](#atomix-multiraft-list-v1-ClearInput)
    - [ClearOutput](#atomix-multiraft-list-v1-ClearOutput)
    - [ContainsInput](#atomix-multiraft-list-v1-ContainsInput)
    - [ContainsOutput](#atomix-multiraft-list-v1-ContainsOutput)
    - [Event](#atomix-multiraft-list-v1-Event)
    - [Event.Appended](#atomix-multiraft-list-v1-Event-Appended)
    - [Event.Inserted](#atomix-multiraft-list-v1-Event-Inserted)
    - [Event.Removed](#atomix-multiraft-list-v1-Event-Removed)
    - [Event.Updated](#atomix-multiraft-list-v1-Event-Updated)
    - [EventsInput](#atomix-multiraft-list-v1-EventsInput)
    - [EventsOutput](#atomix-multiraft-list-v1-EventsOutput)
    - [GetInput](#atomix-multiraft-list-v1-GetInput)
    - [GetOutput](#atomix-multiraft-list-v1-GetOutput)
    - [InsertInput](#atomix-multiraft-list-v1-InsertInput)
    - [InsertOutput](#atomix-multiraft-list-v1-InsertOutput)
    - [Item](#atomix-multiraft-list-v1-Item)
    - [ItemsInput](#atomix-multiraft-list-v1-ItemsInput)
    - [ItemsOutput](#atomix-multiraft-list-v1-ItemsOutput)
    - [ListInput](#atomix-multiraft-list-v1-ListInput)
    - [ListListener](#atomix-multiraft-list-v1-ListListener)
    - [ListOutput](#atomix-multiraft-list-v1-ListOutput)
    - [ListValue](#atomix-multiraft-list-v1-ListValue)
    - [RemoveInput](#atomix-multiraft-list-v1-RemoveInput)
    - [RemoveOutput](#atomix-multiraft-list-v1-RemoveOutput)
    - [SetInput](#atomix-multiraft-list-v1-SetInput)
    - [SetOutput](#atomix-multiraft-list-v1-SetOutput)
    - [SizeInput](#atomix-multiraft-list-v1-SizeInput)
    - [SizeOutput](#atomix-multiraft-list-v1-SizeOutput)
    - [Value](#atomix-multiraft-list-v1-Value)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_list_v1_fsm-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/list/v1/fsm.proto



<a name="atomix-multiraft-list-v1-AppendInput"></a>

### AppendInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-list-v1-Value) |  |  |






<a name="atomix-multiraft-list-v1-AppendOutput"></a>

### AppendOutput







<a name="atomix-multiraft-list-v1-ClearInput"></a>

### ClearInput







<a name="atomix-multiraft-list-v1-ClearOutput"></a>

### ClearOutput







<a name="atomix-multiraft-list-v1-ContainsInput"></a>

### ContainsInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-list-v1-Value) |  |  |






<a name="atomix-multiraft-list-v1-ContainsOutput"></a>

### ContainsOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| contains | [bool](#bool) |  |  |






<a name="atomix-multiraft-list-v1-Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint32](#uint32) |  |  |
| appended | [Event.Appended](#atomix-multiraft-list-v1-Event-Appended) |  |  |
| inserted | [Event.Inserted](#atomix-multiraft-list-v1-Event-Inserted) |  |  |
| updated | [Event.Updated](#atomix-multiraft-list-v1-Event-Updated) |  |  |
| removed | [Event.Removed](#atomix-multiraft-list-v1-Event-Removed) |  |  |






<a name="atomix-multiraft-list-v1-Event-Appended"></a>

### Event.Appended



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-list-v1-Value) |  |  |






<a name="atomix-multiraft-list-v1-Event-Inserted"></a>

### Event.Inserted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-list-v1-Value) |  |  |






<a name="atomix-multiraft-list-v1-Event-Removed"></a>

### Event.Removed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-list-v1-Value) |  |  |






<a name="atomix-multiraft-list-v1-Event-Updated"></a>

### Event.Updated



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-list-v1-Value) |  |  |
| prev_value | [Value](#atomix-multiraft-list-v1-Value) |  |  |






<a name="atomix-multiraft-list-v1-EventsInput"></a>

### EventsInput







<a name="atomix-multiraft-list-v1-EventsOutput"></a>

### EventsOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#atomix-multiraft-list-v1-Event) |  |  |






<a name="atomix-multiraft-list-v1-GetInput"></a>

### GetInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint32](#uint32) |  |  |






<a name="atomix-multiraft-list-v1-GetOutput"></a>

### GetOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| item | [Item](#atomix-multiraft-list-v1-Item) |  |  |






<a name="atomix-multiraft-list-v1-InsertInput"></a>

### InsertInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint32](#uint32) |  |  |
| value | [Value](#atomix-multiraft-list-v1-Value) |  |  |






<a name="atomix-multiraft-list-v1-InsertOutput"></a>

### InsertOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| item | [Item](#atomix-multiraft-list-v1-Item) |  |  |






<a name="atomix-multiraft-list-v1-Item"></a>

### Item



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint32](#uint32) |  |  |
| value | [Value](#atomix-multiraft-list-v1-Value) |  |  |






<a name="atomix-multiraft-list-v1-ItemsInput"></a>

### ItemsInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| watch | [bool](#bool) |  |  |






<a name="atomix-multiraft-list-v1-ItemsOutput"></a>

### ItemsOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| item | [Item](#atomix-multiraft-list-v1-Item) |  |  |






<a name="atomix-multiraft-list-v1-ListInput"></a>

### ListInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [SizeInput](#atomix-multiraft-list-v1-SizeInput) |  |  |
| append | [AppendInput](#atomix-multiraft-list-v1-AppendInput) |  |  |
| insert | [InsertInput](#atomix-multiraft-list-v1-InsertInput) |  |  |
| get | [GetInput](#atomix-multiraft-list-v1-GetInput) |  |  |
| set | [SetInput](#atomix-multiraft-list-v1-SetInput) |  |  |
| remove | [RemoveInput](#atomix-multiraft-list-v1-RemoveInput) |  |  |
| clear | [ClearInput](#atomix-multiraft-list-v1-ClearInput) |  |  |
| events | [EventsInput](#atomix-multiraft-list-v1-EventsInput) |  |  |
| items | [ItemsInput](#atomix-multiraft-list-v1-ItemsInput) |  |  |






<a name="atomix-multiraft-list-v1-ListListener"></a>

### ListListener



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-list-v1-ListOutput"></a>

### ListOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [SizeOutput](#atomix-multiraft-list-v1-SizeOutput) |  |  |
| append | [AppendOutput](#atomix-multiraft-list-v1-AppendOutput) |  |  |
| insert | [InsertOutput](#atomix-multiraft-list-v1-InsertOutput) |  |  |
| get | [GetOutput](#atomix-multiraft-list-v1-GetOutput) |  |  |
| set | [SetOutput](#atomix-multiraft-list-v1-SetOutput) |  |  |
| remove | [RemoveOutput](#atomix-multiraft-list-v1-RemoveOutput) |  |  |
| clear | [ClearOutput](#atomix-multiraft-list-v1-ClearOutput) |  |  |
| events | [EventsOutput](#atomix-multiraft-list-v1-EventsOutput) |  |  |
| items | [ItemsOutput](#atomix-multiraft-list-v1-ItemsOutput) |  |  |






<a name="atomix-multiraft-list-v1-ListValue"></a>

### ListValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [string](#string) |  |  |






<a name="atomix-multiraft-list-v1-RemoveInput"></a>

### RemoveInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint32](#uint32) |  |  |






<a name="atomix-multiraft-list-v1-RemoveOutput"></a>

### RemoveOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| item | [Item](#atomix-multiraft-list-v1-Item) |  |  |






<a name="atomix-multiraft-list-v1-SetInput"></a>

### SetInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint32](#uint32) |  |  |
| value | [Value](#atomix-multiraft-list-v1-Value) |  |  |






<a name="atomix-multiraft-list-v1-SetOutput"></a>

### SetOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| item | [Item](#atomix-multiraft-list-v1-Item) |  |  |






<a name="atomix-multiraft-list-v1-SizeInput"></a>

### SizeInput







<a name="atomix-multiraft-list-v1-SizeOutput"></a>

### SizeOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [uint32](#uint32) |  |  |






<a name="atomix-multiraft-list-v1-Value"></a>

### Value



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |





 

 

 

 



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

