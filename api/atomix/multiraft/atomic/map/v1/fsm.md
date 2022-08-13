# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/atomic/map/v1/fsm.proto](#atomix_multiraft_atomic_map_v1_fsm-proto)
    - [AtomicMapEntry](#atomix-multiraft-atomic-map-v1-AtomicMapEntry)
    - [AtomicMapInput](#atomix-multiraft-atomic-map-v1-AtomicMapInput)
    - [AtomicMapListener](#atomix-multiraft-atomic-map-v1-AtomicMapListener)
    - [AtomicMapOutput](#atomix-multiraft-atomic-map-v1-AtomicMapOutput)
    - [AtomicMapValue](#atomix-multiraft-atomic-map-v1-AtomicMapValue)
    - [ClearInput](#atomix-multiraft-atomic-map-v1-ClearInput)
    - [ClearOutput](#atomix-multiraft-atomic-map-v1-ClearOutput)
    - [EntriesInput](#atomix-multiraft-atomic-map-v1-EntriesInput)
    - [EntriesOutput](#atomix-multiraft-atomic-map-v1-EntriesOutput)
    - [Entry](#atomix-multiraft-atomic-map-v1-Entry)
    - [Event](#atomix-multiraft-atomic-map-v1-Event)
    - [Event.Inserted](#atomix-multiraft-atomic-map-v1-Event-Inserted)
    - [Event.Removed](#atomix-multiraft-atomic-map-v1-Event-Removed)
    - [Event.Updated](#atomix-multiraft-atomic-map-v1-Event-Updated)
    - [EventsInput](#atomix-multiraft-atomic-map-v1-EventsInput)
    - [EventsOutput](#atomix-multiraft-atomic-map-v1-EventsOutput)
    - [GetInput](#atomix-multiraft-atomic-map-v1-GetInput)
    - [GetOutput](#atomix-multiraft-atomic-map-v1-GetOutput)
    - [InsertInput](#atomix-multiraft-atomic-map-v1-InsertInput)
    - [InsertOutput](#atomix-multiraft-atomic-map-v1-InsertOutput)
    - [LockInput](#atomix-multiraft-atomic-map-v1-LockInput)
    - [LockOutput](#atomix-multiraft-atomic-map-v1-LockOutput)
    - [PutInput](#atomix-multiraft-atomic-map-v1-PutInput)
    - [PutOutput](#atomix-multiraft-atomic-map-v1-PutOutput)
    - [RemoveInput](#atomix-multiraft-atomic-map-v1-RemoveInput)
    - [RemoveOutput](#atomix-multiraft-atomic-map-v1-RemoveOutput)
    - [SizeInput](#atomix-multiraft-atomic-map-v1-SizeInput)
    - [SizeOutput](#atomix-multiraft-atomic-map-v1-SizeOutput)
    - [UnlockInput](#atomix-multiraft-atomic-map-v1-UnlockInput)
    - [UnlockOutput](#atomix-multiraft-atomic-map-v1-UnlockOutput)
    - [UpdateInput](#atomix-multiraft-atomic-map-v1-UpdateInput)
    - [UpdateOutput](#atomix-multiraft-atomic-map-v1-UpdateOutput)
    - [Value](#atomix-multiraft-atomic-map-v1-Value)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_atomic_map_v1_fsm-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/atomic/map/v1/fsm.proto



<a name="atomix-multiraft-atomic-map-v1-AtomicMapEntry"></a>

### AtomicMapEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [AtomicMapValue](#atomix-multiraft-atomic-map-v1-AtomicMapValue) |  |  |






<a name="atomix-multiraft-atomic-map-v1-AtomicMapInput"></a>

### AtomicMapInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [SizeInput](#atomix-multiraft-atomic-map-v1-SizeInput) |  |  |
| put | [PutInput](#atomix-multiraft-atomic-map-v1-PutInput) |  |  |
| insert | [InsertInput](#atomix-multiraft-atomic-map-v1-InsertInput) |  |  |
| update | [UpdateInput](#atomix-multiraft-atomic-map-v1-UpdateInput) |  |  |
| get | [GetInput](#atomix-multiraft-atomic-map-v1-GetInput) |  |  |
| remove | [RemoveInput](#atomix-multiraft-atomic-map-v1-RemoveInput) |  |  |
| clear | [ClearInput](#atomix-multiraft-atomic-map-v1-ClearInput) |  |  |
| lock | [LockInput](#atomix-multiraft-atomic-map-v1-LockInput) |  |  |
| unlock | [UnlockInput](#atomix-multiraft-atomic-map-v1-UnlockInput) |  |  |
| entries | [EntriesInput](#atomix-multiraft-atomic-map-v1-EntriesInput) |  |  |
| events | [EventsInput](#atomix-multiraft-atomic-map-v1-EventsInput) |  |  |






<a name="atomix-multiraft-atomic-map-v1-AtomicMapListener"></a>

### AtomicMapListener



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| key | [string](#string) |  |  |






<a name="atomix-multiraft-atomic-map-v1-AtomicMapOutput"></a>

### AtomicMapOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [SizeOutput](#atomix-multiraft-atomic-map-v1-SizeOutput) |  |  |
| put | [PutOutput](#atomix-multiraft-atomic-map-v1-PutOutput) |  |  |
| insert | [InsertOutput](#atomix-multiraft-atomic-map-v1-InsertOutput) |  |  |
| update | [UpdateOutput](#atomix-multiraft-atomic-map-v1-UpdateOutput) |  |  |
| get | [GetOutput](#atomix-multiraft-atomic-map-v1-GetOutput) |  |  |
| remove | [RemoveOutput](#atomix-multiraft-atomic-map-v1-RemoveOutput) |  |  |
| clear | [ClearOutput](#atomix-multiraft-atomic-map-v1-ClearOutput) |  |  |
| lock | [LockOutput](#atomix-multiraft-atomic-map-v1-LockOutput) |  |  |
| unlock | [UnlockOutput](#atomix-multiraft-atomic-map-v1-UnlockOutput) |  |  |
| entries | [EntriesOutput](#atomix-multiraft-atomic-map-v1-EntriesOutput) |  |  |
| events | [EventsOutput](#atomix-multiraft-atomic-map-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-atomic-map-v1-AtomicMapValue"></a>

### AtomicMapValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| index | [uint64](#uint64) |  |  |
| expire | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="atomix-multiraft-atomic-map-v1-ClearInput"></a>

### ClearInput







<a name="atomix-multiraft-atomic-map-v1-ClearOutput"></a>

### ClearOutput







<a name="atomix-multiraft-atomic-map-v1-EntriesInput"></a>

### EntriesInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| watch | [bool](#bool) |  |  |






<a name="atomix-multiraft-atomic-map-v1-EntriesOutput"></a>

### EntriesOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-atomic-map-v1-Entry) |  |  |






<a name="atomix-multiraft-atomic-map-v1-Entry"></a>

### Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Value](#atomix-multiraft-atomic-map-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-map-v1-Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| inserted | [Event.Inserted](#atomix-multiraft-atomic-map-v1-Event-Inserted) |  |  |
| updated | [Event.Updated](#atomix-multiraft-atomic-map-v1-Event-Updated) |  |  |
| removed | [Event.Removed](#atomix-multiraft-atomic-map-v1-Event-Removed) |  |  |






<a name="atomix-multiraft-atomic-map-v1-Event-Inserted"></a>

### Event.Inserted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-atomic-map-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-map-v1-Event-Removed"></a>

### Event.Removed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-atomic-map-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-map-v1-Event-Updated"></a>

### Event.Updated



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| new_value | [Value](#atomix-multiraft-atomic-map-v1-Value) |  |  |
| prev_value | [Value](#atomix-multiraft-atomic-map-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-map-v1-EventsInput"></a>

### EventsInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |






<a name="atomix-multiraft-atomic-map-v1-EventsOutput"></a>

### EventsOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#atomix-multiraft-atomic-map-v1-Event) |  |  |






<a name="atomix-multiraft-atomic-map-v1-GetInput"></a>

### GetInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |






<a name="atomix-multiraft-atomic-map-v1-GetOutput"></a>

### GetOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-atomic-map-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-map-v1-InsertInput"></a>

### InsertInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Value](#atomix-multiraft-atomic-map-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-map-v1-InsertOutput"></a>

### InsertOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-atomic-map-v1-LockInput"></a>

### LockInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| keys | [string](#string) | repeated |  |
| timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="atomix-multiraft-atomic-map-v1-LockOutput"></a>

### LockOutput







<a name="atomix-multiraft-atomic-map-v1-PutInput"></a>

### PutInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Value](#atomix-multiraft-atomic-map-v1-Value) |  |  |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-atomic-map-v1-PutOutput"></a>

### PutOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| prev_value | [Value](#atomix-multiraft-atomic-map-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-map-v1-RemoveInput"></a>

### RemoveInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-atomic-map-v1-RemoveOutput"></a>

### RemoveOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-atomic-map-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-map-v1-SizeInput"></a>

### SizeInput







<a name="atomix-multiraft-atomic-map-v1-SizeOutput"></a>

### SizeOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [uint32](#uint32) |  |  |






<a name="atomix-multiraft-atomic-map-v1-UnlockInput"></a>

### UnlockInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| keys | [string](#string) | repeated |  |






<a name="atomix-multiraft-atomic-map-v1-UnlockOutput"></a>

### UnlockOutput







<a name="atomix-multiraft-atomic-map-v1-UpdateInput"></a>

### UpdateInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Value](#atomix-multiraft-atomic-map-v1-Value) |  |  |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-atomic-map-v1-UpdateOutput"></a>

### UpdateOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| prev_value | [Value](#atomix-multiraft-atomic-map-v1-Value) |  |  |






<a name="atomix-multiraft-atomic-map-v1-Value"></a>

### Value



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| index | [uint64](#uint64) |  |  |
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

