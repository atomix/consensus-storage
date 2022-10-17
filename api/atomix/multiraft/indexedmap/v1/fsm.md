# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/indexedmap/v1/fsm.proto](#atomix_multiraft_indexedmap_v1_fsm-proto)
    - [AppendInput](#atomix-multiraft-indexedmap-v1-AppendInput)
    - [AppendOutput](#atomix-multiraft-indexedmap-v1-AppendOutput)
    - [ClearInput](#atomix-multiraft-indexedmap-v1-ClearInput)
    - [ClearOutput](#atomix-multiraft-indexedmap-v1-ClearOutput)
    - [EntriesInput](#atomix-multiraft-indexedmap-v1-EntriesInput)
    - [EntriesOutput](#atomix-multiraft-indexedmap-v1-EntriesOutput)
    - [Entry](#atomix-multiraft-indexedmap-v1-Entry)
    - [Event](#atomix-multiraft-indexedmap-v1-Event)
    - [Event.Inserted](#atomix-multiraft-indexedmap-v1-Event-Inserted)
    - [Event.Removed](#atomix-multiraft-indexedmap-v1-Event-Removed)
    - [Event.Updated](#atomix-multiraft-indexedmap-v1-Event-Updated)
    - [EventsInput](#atomix-multiraft-indexedmap-v1-EventsInput)
    - [EventsOutput](#atomix-multiraft-indexedmap-v1-EventsOutput)
    - [FirstEntryInput](#atomix-multiraft-indexedmap-v1-FirstEntryInput)
    - [FirstEntryOutput](#atomix-multiraft-indexedmap-v1-FirstEntryOutput)
    - [GetInput](#atomix-multiraft-indexedmap-v1-GetInput)
    - [GetOutput](#atomix-multiraft-indexedmap-v1-GetOutput)
    - [IndexedMapEntry](#atomix-multiraft-indexedmap-v1-IndexedMapEntry)
    - [IndexedMapInput](#atomix-multiraft-indexedmap-v1-IndexedMapInput)
    - [IndexedMapListener](#atomix-multiraft-indexedmap-v1-IndexedMapListener)
    - [IndexedMapOutput](#atomix-multiraft-indexedmap-v1-IndexedMapOutput)
    - [IndexedMapValue](#atomix-multiraft-indexedmap-v1-IndexedMapValue)
    - [LastEntryInput](#atomix-multiraft-indexedmap-v1-LastEntryInput)
    - [LastEntryOutput](#atomix-multiraft-indexedmap-v1-LastEntryOutput)
    - [NextEntryInput](#atomix-multiraft-indexedmap-v1-NextEntryInput)
    - [NextEntryOutput](#atomix-multiraft-indexedmap-v1-NextEntryOutput)
    - [PrevEntryInput](#atomix-multiraft-indexedmap-v1-PrevEntryInput)
    - [PrevEntryOutput](#atomix-multiraft-indexedmap-v1-PrevEntryOutput)
    - [RemoveInput](#atomix-multiraft-indexedmap-v1-RemoveInput)
    - [RemoveOutput](#atomix-multiraft-indexedmap-v1-RemoveOutput)
    - [SizeInput](#atomix-multiraft-indexedmap-v1-SizeInput)
    - [SizeOutput](#atomix-multiraft-indexedmap-v1-SizeOutput)
    - [UpdateInput](#atomix-multiraft-indexedmap-v1-UpdateInput)
    - [UpdateOutput](#atomix-multiraft-indexedmap-v1-UpdateOutput)
    - [Value](#atomix-multiraft-indexedmap-v1-Value)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_indexedmap_v1_fsm-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/indexedmap/v1/fsm.proto



<a name="atomix-multiraft-indexedmap-v1-AppendInput"></a>

### AppendInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bytes](#bytes) |  |  |
| ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="atomix-multiraft-indexedmap-v1-AppendOutput"></a>

### AppendOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-indexedmap-v1-Entry) |  |  |






<a name="atomix-multiraft-indexedmap-v1-ClearInput"></a>

### ClearInput







<a name="atomix-multiraft-indexedmap-v1-ClearOutput"></a>

### ClearOutput







<a name="atomix-multiraft-indexedmap-v1-EntriesInput"></a>

### EntriesInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| watch | [bool](#bool) |  |  |






<a name="atomix-multiraft-indexedmap-v1-EntriesOutput"></a>

### EntriesOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-indexedmap-v1-Entry) |  |  |






<a name="atomix-multiraft-indexedmap-v1-Entry"></a>

### Entry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| index | [uint64](#uint64) |  |  |
| value | [Value](#atomix-multiraft-indexedmap-v1-Value) |  |  |






<a name="atomix-multiraft-indexedmap-v1-Event"></a>

### Event



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| index | [uint64](#uint64) |  |  |
| inserted | [Event.Inserted](#atomix-multiraft-indexedmap-v1-Event-Inserted) |  |  |
| updated | [Event.Updated](#atomix-multiraft-indexedmap-v1-Event-Updated) |  |  |
| removed | [Event.Removed](#atomix-multiraft-indexedmap-v1-Event-Removed) |  |  |






<a name="atomix-multiraft-indexedmap-v1-Event-Inserted"></a>

### Event.Inserted



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-indexedmap-v1-Value) |  |  |






<a name="atomix-multiraft-indexedmap-v1-Event-Removed"></a>

### Event.Removed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-indexedmap-v1-Value) |  |  |
| expired | [bool](#bool) |  |  |






<a name="atomix-multiraft-indexedmap-v1-Event-Updated"></a>

### Event.Updated



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [Value](#atomix-multiraft-indexedmap-v1-Value) |  |  |
| prev_value | [Value](#atomix-multiraft-indexedmap-v1-Value) |  |  |






<a name="atomix-multiraft-indexedmap-v1-EventsInput"></a>

### EventsInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |






<a name="atomix-multiraft-indexedmap-v1-EventsOutput"></a>

### EventsOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#atomix-multiraft-indexedmap-v1-Event) |  |  |






<a name="atomix-multiraft-indexedmap-v1-FirstEntryInput"></a>

### FirstEntryInput







<a name="atomix-multiraft-indexedmap-v1-FirstEntryOutput"></a>

### FirstEntryOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-indexedmap-v1-Entry) |  |  |






<a name="atomix-multiraft-indexedmap-v1-GetInput"></a>

### GetInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-indexedmap-v1-GetOutput"></a>

### GetOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-indexedmap-v1-Entry) |  |  |






<a name="atomix-multiraft-indexedmap-v1-IndexedMapEntry"></a>

### IndexedMapEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| index | [uint64](#uint64) |  |  |
| value | [IndexedMapValue](#atomix-multiraft-indexedmap-v1-IndexedMapValue) |  |  |






<a name="atomix-multiraft-indexedmap-v1-IndexedMapInput"></a>

### IndexedMapInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [SizeInput](#atomix-multiraft-indexedmap-v1-SizeInput) |  |  |
| append | [AppendInput](#atomix-multiraft-indexedmap-v1-AppendInput) |  |  |
| update | [UpdateInput](#atomix-multiraft-indexedmap-v1-UpdateInput) |  |  |
| get | [GetInput](#atomix-multiraft-indexedmap-v1-GetInput) |  |  |
| first_entry | [FirstEntryInput](#atomix-multiraft-indexedmap-v1-FirstEntryInput) |  |  |
| last_entry | [LastEntryInput](#atomix-multiraft-indexedmap-v1-LastEntryInput) |  |  |
| next_entry | [NextEntryInput](#atomix-multiraft-indexedmap-v1-NextEntryInput) |  |  |
| prev_entry | [PrevEntryInput](#atomix-multiraft-indexedmap-v1-PrevEntryInput) |  |  |
| remove | [RemoveInput](#atomix-multiraft-indexedmap-v1-RemoveInput) |  |  |
| clear | [ClearInput](#atomix-multiraft-indexedmap-v1-ClearInput) |  |  |
| entries | [EntriesInput](#atomix-multiraft-indexedmap-v1-EntriesInput) |  |  |
| events | [EventsInput](#atomix-multiraft-indexedmap-v1-EventsInput) |  |  |






<a name="atomix-multiraft-indexedmap-v1-IndexedMapListener"></a>

### IndexedMapListener



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| key | [string](#string) |  |  |






<a name="atomix-multiraft-indexedmap-v1-IndexedMapOutput"></a>

### IndexedMapOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [SizeOutput](#atomix-multiraft-indexedmap-v1-SizeOutput) |  |  |
| append | [AppendOutput](#atomix-multiraft-indexedmap-v1-AppendOutput) |  |  |
| update | [UpdateOutput](#atomix-multiraft-indexedmap-v1-UpdateOutput) |  |  |
| get | [GetOutput](#atomix-multiraft-indexedmap-v1-GetOutput) |  |  |
| first_entry | [FirstEntryOutput](#atomix-multiraft-indexedmap-v1-FirstEntryOutput) |  |  |
| last_entry | [LastEntryOutput](#atomix-multiraft-indexedmap-v1-LastEntryOutput) |  |  |
| next_entry | [NextEntryOutput](#atomix-multiraft-indexedmap-v1-NextEntryOutput) |  |  |
| prev_entry | [PrevEntryOutput](#atomix-multiraft-indexedmap-v1-PrevEntryOutput) |  |  |
| remove | [RemoveOutput](#atomix-multiraft-indexedmap-v1-RemoveOutput) |  |  |
| clear | [ClearOutput](#atomix-multiraft-indexedmap-v1-ClearOutput) |  |  |
| entries | [EntriesOutput](#atomix-multiraft-indexedmap-v1-EntriesOutput) |  |  |
| events | [EventsOutput](#atomix-multiraft-indexedmap-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-indexedmap-v1-IndexedMapValue"></a>

### IndexedMapValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| version | [uint64](#uint64) |  |  |
| expire | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="atomix-multiraft-indexedmap-v1-LastEntryInput"></a>

### LastEntryInput







<a name="atomix-multiraft-indexedmap-v1-LastEntryOutput"></a>

### LastEntryOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-indexedmap-v1-Entry) |  |  |






<a name="atomix-multiraft-indexedmap-v1-NextEntryInput"></a>

### NextEntryInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-indexedmap-v1-NextEntryOutput"></a>

### NextEntryOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-indexedmap-v1-Entry) |  |  |






<a name="atomix-multiraft-indexedmap-v1-PrevEntryInput"></a>

### PrevEntryInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-indexedmap-v1-PrevEntryOutput"></a>

### PrevEntryOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-indexedmap-v1-Entry) |  |  |






<a name="atomix-multiraft-indexedmap-v1-RemoveInput"></a>

### RemoveInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| index | [uint64](#uint64) |  |  |
| prev_version | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-indexedmap-v1-RemoveOutput"></a>

### RemoveOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-indexedmap-v1-Entry) |  |  |






<a name="atomix-multiraft-indexedmap-v1-SizeInput"></a>

### SizeInput







<a name="atomix-multiraft-indexedmap-v1-SizeOutput"></a>

### SizeOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size | [uint32](#uint32) |  |  |






<a name="atomix-multiraft-indexedmap-v1-UpdateInput"></a>

### UpdateInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| index | [uint64](#uint64) |  |  |
| value | [bytes](#bytes) |  |  |
| ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| prev_version | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-indexedmap-v1-UpdateOutput"></a>

### UpdateOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entry | [Entry](#atomix-multiraft-indexedmap-v1-Entry) |  |  |






<a name="atomix-multiraft-indexedmap-v1-Value"></a>

### Value



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [bytes](#bytes) |  |  |
| version | [uint64](#uint64) |  |  |





 

 

 

 



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

