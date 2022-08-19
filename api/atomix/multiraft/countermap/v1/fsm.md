# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/countermap/v1/fsm.proto](#atomix_multiraft_countermap_v1_fsm-proto)
    - [ClearInput](#atomix-multiraft-countermap-v1-ClearInput)
    - [ClearOutput](#atomix-multiraft-countermap-v1-ClearOutput)
    - [CounterMapInput](#atomix-multiraft-countermap-v1-CounterMapInput)
    - [CounterMapListener](#atomix-multiraft-countermap-v1-CounterMapListener)
    - [CounterMapOutput](#atomix-multiraft-countermap-v1-CounterMapOutput)
    - [DecrementInput](#atomix-multiraft-countermap-v1-DecrementInput)
    - [DecrementOutput](#atomix-multiraft-countermap-v1-DecrementOutput)
    - [EntriesInput](#atomix-multiraft-countermap-v1-EntriesInput)
    - [EntriesOutput](#atomix-multiraft-countermap-v1-EntriesOutput)
    - [Entry](#atomix-multiraft-countermap-v1-Entry)
    - [Event](#atomix-multiraft-countermap-v1-Event)
    - [Event.Inserted](#atomix-multiraft-countermap-v1-Event-Inserted)
    - [Event.Removed](#atomix-multiraft-countermap-v1-Event-Removed)
    - [Event.Updated](#atomix-multiraft-countermap-v1-Event-Updated)
    - [EventsInput](#atomix-multiraft-countermap-v1-EventsInput)
    - [EventsOutput](#atomix-multiraft-countermap-v1-EventsOutput)
    - [GetInput](#atomix-multiraft-countermap-v1-GetInput)
    - [GetOutput](#atomix-multiraft-countermap-v1-GetOutput)
    - [IncrementInput](#atomix-multiraft-countermap-v1-IncrementInput)
    - [IncrementOutput](#atomix-multiraft-countermap-v1-IncrementOutput)
    - [InsertInput](#atomix-multiraft-countermap-v1-InsertInput)
    - [InsertOutput](#atomix-multiraft-countermap-v1-InsertOutput)
    - [LockInput](#atomix-multiraft-countermap-v1-LockInput)
    - [LockOutput](#atomix-multiraft-countermap-v1-LockOutput)
    - [RemoveInput](#atomix-multiraft-countermap-v1-RemoveInput)
    - [RemoveOutput](#atomix-multiraft-countermap-v1-RemoveOutput)
    - [SetInput](#atomix-multiraft-countermap-v1-SetInput)
    - [SetOutput](#atomix-multiraft-countermap-v1-SetOutput)
    - [SizeInput](#atomix-multiraft-countermap-v1-SizeInput)
    - [SizeOutput](#atomix-multiraft-countermap-v1-SizeOutput)
    - [UnlockInput](#atomix-multiraft-countermap-v1-UnlockInput)
    - [UnlockOutput](#atomix-multiraft-countermap-v1-UnlockOutput)
    - [UpdateInput](#atomix-multiraft-countermap-v1-UpdateInput)
    - [UpdateOutput](#atomix-multiraft-countermap-v1-UpdateOutput)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_countermap_v1_fsm-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/countermap/v1/fsm.proto



<a name="atomix-multiraft-countermap-v1-ClearInput"></a>

### ClearInput







<a name="atomix-multiraft-countermap-v1-ClearOutput"></a>

### ClearOutput







<a name="atomix-multiraft-countermap-v1-CounterMapInput"></a>

### CounterMapInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [SizeInput](#atomix-multiraft-countermap-v1-SizeInput) |  |  |
| set | [SetInput](#atomix-multiraft-countermap-v1-SetInput) |  |  |
| insert | [InsertInput](#atomix-multiraft-countermap-v1-InsertInput) |  |  |
| update | [UpdateInput](#atomix-multiraft-countermap-v1-UpdateInput) |  |  |
| increment | [IncrementInput](#atomix-multiraft-countermap-v1-IncrementInput) |  |  |
| decrement | [DecrementInput](#atomix-multiraft-countermap-v1-DecrementInput) |  |  |
| get | [GetInput](#atomix-multiraft-countermap-v1-GetInput) |  |  |
| remove | [RemoveInput](#atomix-multiraft-countermap-v1-RemoveInput) |  |  |
| clear | [ClearInput](#atomix-multiraft-countermap-v1-ClearInput) |  |  |
| lock | [LockInput](#atomix-multiraft-countermap-v1-LockInput) |  |  |
| unlock | [UnlockInput](#atomix-multiraft-countermap-v1-UnlockInput) |  |  |
| entries | [EntriesInput](#atomix-multiraft-countermap-v1-EntriesInput) |  |  |
| events | [EventsInput](#atomix-multiraft-countermap-v1-EventsInput) |  |  |






<a name="atomix-multiraft-countermap-v1-CounterMapListener"></a>

### CounterMapListener



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| key | [string](#string) |  |  |






<a name="atomix-multiraft-countermap-v1-CounterMapOutput"></a>

### CounterMapOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [SizeOutput](#atomix-multiraft-countermap-v1-SizeOutput) |  |  |
| set | [SetOutput](#atomix-multiraft-countermap-v1-SetOutput) |  |  |
| insert | [InsertOutput](#atomix-multiraft-countermap-v1-InsertOutput) |  |  |
| update | [UpdateOutput](#atomix-multiraft-countermap-v1-UpdateOutput) |  |  |
| increment | [IncrementOutput](#atomix-multiraft-countermap-v1-IncrementOutput) |  |  |
| decrement | [DecrementOutput](#atomix-multiraft-countermap-v1-DecrementOutput) |  |  |
| get | [GetOutput](#atomix-multiraft-countermap-v1-GetOutput) |  |  |
| remove | [RemoveOutput](#atomix-multiraft-countermap-v1-RemoveOutput) |  |  |
| clear | [ClearOutput](#atomix-multiraft-countermap-v1-ClearOutput) |  |  |
| lock | [LockOutput](#atomix-multiraft-countermap-v1-LockOutput) |  |  |
| unlock | [UnlockOutput](#atomix-multiraft-countermap-v1-UnlockOutput) |  |  |
| entries | [EntriesOutput](#atomix-multiraft-countermap-v1-EntriesOutput) |  |  |
| events | [EventsOutput](#atomix-multiraft-countermap-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-countermap-v1-DecrementInput"></a>

### DecrementInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| delta | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-DecrementOutput"></a>

### DecrementOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| prev_value | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-EntriesInput"></a>

### EntriesInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| watch | [bool](#bool) |  |  |






<a name="atomix-multiraft-countermap-v1-EntriesOutput"></a>

### EntriesOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-countermap-v1-Entry) |  |  |






<a name="atomix-multiraft-countermap-v1-Entry"></a>

### Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| inserted | [Event.Inserted](#atomix-multiraft-countermap-v1-Event-Inserted) |  |  |
| updated | [Event.Updated](#atomix-multiraft-countermap-v1-Event-Updated) |  |  |
| removed | [Event.Removed](#atomix-multiraft-countermap-v1-Event-Removed) |  |  |






<a name="atomix-multiraft-countermap-v1-Event-Inserted"></a>

### Event.Inserted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-Event-Removed"></a>

### Event.Removed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-Event-Updated"></a>

### Event.Updated



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [int64](#int64) |  |  |
| prev_value | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-EventsInput"></a>

### EventsInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |






<a name="atomix-multiraft-countermap-v1-EventsOutput"></a>

### EventsOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#atomix-multiraft-countermap-v1-Event) |  |  |






<a name="atomix-multiraft-countermap-v1-GetInput"></a>

### GetInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |






<a name="atomix-multiraft-countermap-v1-GetOutput"></a>

### GetOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-IncrementInput"></a>

### IncrementInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| delta | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-IncrementOutput"></a>

### IncrementOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| prev_value | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-InsertInput"></a>

### InsertInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-InsertOutput"></a>

### InsertOutput







<a name="atomix-multiraft-countermap-v1-LockInput"></a>

### LockInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| keys | [string](#string) | repeated |  |
| timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="atomix-multiraft-countermap-v1-LockOutput"></a>

### LockOutput







<a name="atomix-multiraft-countermap-v1-RemoveInput"></a>

### RemoveInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| prev_value | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-RemoveOutput"></a>

### RemoveOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-SetInput"></a>

### SetInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-SetOutput"></a>

### SetOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| prev_value | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-SizeInput"></a>

### SizeInput







<a name="atomix-multiraft-countermap-v1-SizeOutput"></a>

### SizeOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [uint32](#uint32) |  |  |






<a name="atomix-multiraft-countermap-v1-UnlockInput"></a>

### UnlockInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| keys | [string](#string) | repeated |  |






<a name="atomix-multiraft-countermap-v1-UnlockOutput"></a>

### UnlockOutput







<a name="atomix-multiraft-countermap-v1-UpdateInput"></a>

### UpdateInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [int64](#int64) |  |  |
| prev_value | [int64](#int64) |  |  |






<a name="atomix-multiraft-countermap-v1-UpdateOutput"></a>

### UpdateOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| prev_value | [int64](#int64) |  |  |





 

 

 

 



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

