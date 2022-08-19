# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/countermap/v1/service.proto](#atomix_multiraft_countermap_v1_service-proto)
    - [ClearRequest](#atomix-multiraft-countermap-v1-ClearRequest)
    - [ClearResponse](#atomix-multiraft-countermap-v1-ClearResponse)
    - [DecrementRequest](#atomix-multiraft-countermap-v1-DecrementRequest)
    - [DecrementResponse](#atomix-multiraft-countermap-v1-DecrementResponse)
    - [EntriesRequest](#atomix-multiraft-countermap-v1-EntriesRequest)
    - [EntriesResponse](#atomix-multiraft-countermap-v1-EntriesResponse)
    - [EventsRequest](#atomix-multiraft-countermap-v1-EventsRequest)
    - [EventsResponse](#atomix-multiraft-countermap-v1-EventsResponse)
    - [GetRequest](#atomix-multiraft-countermap-v1-GetRequest)
    - [GetResponse](#atomix-multiraft-countermap-v1-GetResponse)
    - [IncrementRequest](#atomix-multiraft-countermap-v1-IncrementRequest)
    - [IncrementResponse](#atomix-multiraft-countermap-v1-IncrementResponse)
    - [InsertRequest](#atomix-multiraft-countermap-v1-InsertRequest)
    - [InsertResponse](#atomix-multiraft-countermap-v1-InsertResponse)
    - [LockRequest](#atomix-multiraft-countermap-v1-LockRequest)
    - [LockResponse](#atomix-multiraft-countermap-v1-LockResponse)
    - [RemoveRequest](#atomix-multiraft-countermap-v1-RemoveRequest)
    - [RemoveResponse](#atomix-multiraft-countermap-v1-RemoveResponse)
    - [SetRequest](#atomix-multiraft-countermap-v1-SetRequest)
    - [SetResponse](#atomix-multiraft-countermap-v1-SetResponse)
    - [SizeRequest](#atomix-multiraft-countermap-v1-SizeRequest)
    - [SizeResponse](#atomix-multiraft-countermap-v1-SizeResponse)
    - [UnlockRequest](#atomix-multiraft-countermap-v1-UnlockRequest)
    - [UnlockResponse](#atomix-multiraft-countermap-v1-UnlockResponse)
    - [UpdateRequest](#atomix-multiraft-countermap-v1-UpdateRequest)
    - [UpdateResponse](#atomix-multiraft-countermap-v1-UpdateResponse)
  
    - [CounterMap](#atomix-multiraft-countermap-v1-CounterMap)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_countermap_v1_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/countermap/v1/service.proto



<a name="atomix-multiraft-countermap-v1-ClearRequest"></a>

### ClearRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [ClearInput](#atomix-multiraft-countermap-v1-ClearInput) |  |  |






<a name="atomix-multiraft-countermap-v1-ClearResponse"></a>

### ClearResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [ClearOutput](#atomix-multiraft-countermap-v1-ClearOutput) |  |  |






<a name="atomix-multiraft-countermap-v1-DecrementRequest"></a>

### DecrementRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [DecrementInput](#atomix-multiraft-countermap-v1-DecrementInput) |  |  |






<a name="atomix-multiraft-countermap-v1-DecrementResponse"></a>

### DecrementResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [DecrementOutput](#atomix-multiraft-countermap-v1-DecrementOutput) |  |  |






<a name="atomix-multiraft-countermap-v1-EntriesRequest"></a>

### EntriesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [EntriesInput](#atomix-multiraft-countermap-v1-EntriesInput) |  |  |






<a name="atomix-multiraft-countermap-v1-EntriesResponse"></a>

### EntriesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [EntriesOutput](#atomix-multiraft-countermap-v1-EntriesOutput) |  |  |






<a name="atomix-multiraft-countermap-v1-EventsRequest"></a>

### EventsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [EventsInput](#atomix-multiraft-countermap-v1-EventsInput) |  |  |






<a name="atomix-multiraft-countermap-v1-EventsResponse"></a>

### EventsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [EventsOutput](#atomix-multiraft-countermap-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-countermap-v1-GetRequest"></a>

### GetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [GetInput](#atomix-multiraft-countermap-v1-GetInput) |  |  |






<a name="atomix-multiraft-countermap-v1-GetResponse"></a>

### GetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [GetOutput](#atomix-multiraft-countermap-v1-GetOutput) |  |  |






<a name="atomix-multiraft-countermap-v1-IncrementRequest"></a>

### IncrementRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [IncrementInput](#atomix-multiraft-countermap-v1-IncrementInput) |  |  |






<a name="atomix-multiraft-countermap-v1-IncrementResponse"></a>

### IncrementResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [IncrementOutput](#atomix-multiraft-countermap-v1-IncrementOutput) |  |  |






<a name="atomix-multiraft-countermap-v1-InsertRequest"></a>

### InsertRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [InsertInput](#atomix-multiraft-countermap-v1-InsertInput) |  |  |






<a name="atomix-multiraft-countermap-v1-InsertResponse"></a>

### InsertResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [InsertOutput](#atomix-multiraft-countermap-v1-InsertOutput) |  |  |






<a name="atomix-multiraft-countermap-v1-LockRequest"></a>

### LockRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [LockInput](#atomix-multiraft-countermap-v1-LockInput) |  |  |






<a name="atomix-multiraft-countermap-v1-LockResponse"></a>

### LockResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [LockOutput](#atomix-multiraft-countermap-v1-LockOutput) |  |  |






<a name="atomix-multiraft-countermap-v1-RemoveRequest"></a>

### RemoveRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [RemoveInput](#atomix-multiraft-countermap-v1-RemoveInput) |  |  |






<a name="atomix-multiraft-countermap-v1-RemoveResponse"></a>

### RemoveResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [RemoveOutput](#atomix-multiraft-countermap-v1-RemoveOutput) |  |  |






<a name="atomix-multiraft-countermap-v1-SetRequest"></a>

### SetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [SetInput](#atomix-multiraft-countermap-v1-SetInput) |  |  |






<a name="atomix-multiraft-countermap-v1-SetResponse"></a>

### SetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [SetOutput](#atomix-multiraft-countermap-v1-SetOutput) |  |  |






<a name="atomix-multiraft-countermap-v1-SizeRequest"></a>

### SizeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [SizeInput](#atomix-multiraft-countermap-v1-SizeInput) |  |  |






<a name="atomix-multiraft-countermap-v1-SizeResponse"></a>

### SizeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [SizeOutput](#atomix-multiraft-countermap-v1-SizeOutput) |  |  |






<a name="atomix-multiraft-countermap-v1-UnlockRequest"></a>

### UnlockRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [UnlockInput](#atomix-multiraft-countermap-v1-UnlockInput) |  |  |






<a name="atomix-multiraft-countermap-v1-UnlockResponse"></a>

### UnlockResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [UnlockOutput](#atomix-multiraft-countermap-v1-UnlockOutput) |  |  |






<a name="atomix-multiraft-countermap-v1-UpdateRequest"></a>

### UpdateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [UpdateInput](#atomix-multiraft-countermap-v1-UpdateInput) |  |  |






<a name="atomix-multiraft-countermap-v1-UpdateResponse"></a>

### UpdateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [UpdateOutput](#atomix-multiraft-countermap-v1-UpdateOutput) |  |  |





 

 

 


<a name="atomix-multiraft-countermap-v1-CounterMap"></a>

### CounterMap
CounterMap is a service for a counter map primitive

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Size | [SizeRequest](#atomix-multiraft-countermap-v1-SizeRequest) | [SizeResponse](#atomix-multiraft-countermap-v1-SizeResponse) | Size returns the size of the map |
| Set | [SetRequest](#atomix-multiraft-countermap-v1-SetRequest) | [SetResponse](#atomix-multiraft-countermap-v1-SetResponse) | Set sets an entry into the map |
| Insert | [InsertRequest](#atomix-multiraft-countermap-v1-InsertRequest) | [InsertResponse](#atomix-multiraft-countermap-v1-InsertResponse) | Insert inserts an entry into the map |
| Update | [UpdateRequest](#atomix-multiraft-countermap-v1-UpdateRequest) | [UpdateResponse](#atomix-multiraft-countermap-v1-UpdateResponse) | Update updates an entry in the map |
| Increment | [IncrementRequest](#atomix-multiraft-countermap-v1-IncrementRequest) | [IncrementResponse](#atomix-multiraft-countermap-v1-IncrementResponse) | Increment increments a counter in the map |
| Decrement | [DecrementRequest](#atomix-multiraft-countermap-v1-DecrementRequest) | [DecrementResponse](#atomix-multiraft-countermap-v1-DecrementResponse) | Decrement decrements a counter in the map |
| Get | [GetRequest](#atomix-multiraft-countermap-v1-GetRequest) | [GetResponse](#atomix-multiraft-countermap-v1-GetResponse) | Get gets the entry for a key |
| Remove | [RemoveRequest](#atomix-multiraft-countermap-v1-RemoveRequest) | [RemoveResponse](#atomix-multiraft-countermap-v1-RemoveResponse) | Remove removes an entry from the map |
| Clear | [ClearRequest](#atomix-multiraft-countermap-v1-ClearRequest) | [ClearResponse](#atomix-multiraft-countermap-v1-ClearResponse) | Clear removes all entries from the map |
| Lock | [LockRequest](#atomix-multiraft-countermap-v1-LockRequest) | [LockResponse](#atomix-multiraft-countermap-v1-LockResponse) | Lock locks a key in the map |
| Unlock | [UnlockRequest](#atomix-multiraft-countermap-v1-UnlockRequest) | [UnlockResponse](#atomix-multiraft-countermap-v1-UnlockResponse) | Unlock unlocks a key in the map |
| Events | [EventsRequest](#atomix-multiraft-countermap-v1-EventsRequest) | [EventsResponse](#atomix-multiraft-countermap-v1-EventsResponse) stream | Events listens for change events |
| Entries | [EntriesRequest](#atomix-multiraft-countermap-v1-EntriesRequest) | [EntriesResponse](#atomix-multiraft-countermap-v1-EntriesResponse) stream | Entries lists all entries in the map |

 



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

