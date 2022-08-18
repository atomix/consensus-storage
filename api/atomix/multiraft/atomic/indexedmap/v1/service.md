# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/atomic/indexedmap/v1/service.proto](#atomix_multiraft_atomic_indexedmap_v1_service-proto)
    - [AppendRequest](#atomix-multiraft-atomic-indexedmap-v1-AppendRequest)
    - [AppendResponse](#atomix-multiraft-atomic-indexedmap-v1-AppendResponse)
    - [ClearRequest](#atomix-multiraft-atomic-indexedmap-v1-ClearRequest)
    - [ClearResponse](#atomix-multiraft-atomic-indexedmap-v1-ClearResponse)
    - [EntriesRequest](#atomix-multiraft-atomic-indexedmap-v1-EntriesRequest)
    - [EntriesResponse](#atomix-multiraft-atomic-indexedmap-v1-EntriesResponse)
    - [EventsRequest](#atomix-multiraft-atomic-indexedmap-v1-EventsRequest)
    - [EventsResponse](#atomix-multiraft-atomic-indexedmap-v1-EventsResponse)
    - [FirstEntryRequest](#atomix-multiraft-atomic-indexedmap-v1-FirstEntryRequest)
    - [FirstEntryResponse](#atomix-multiraft-atomic-indexedmap-v1-FirstEntryResponse)
    - [GetRequest](#atomix-multiraft-atomic-indexedmap-v1-GetRequest)
    - [GetResponse](#atomix-multiraft-atomic-indexedmap-v1-GetResponse)
    - [LastEntryRequest](#atomix-multiraft-atomic-indexedmap-v1-LastEntryRequest)
    - [LastEntryResponse](#atomix-multiraft-atomic-indexedmap-v1-LastEntryResponse)
    - [NextEntryRequest](#atomix-multiraft-atomic-indexedmap-v1-NextEntryRequest)
    - [NextEntryResponse](#atomix-multiraft-atomic-indexedmap-v1-NextEntryResponse)
    - [PrevEntryRequest](#atomix-multiraft-atomic-indexedmap-v1-PrevEntryRequest)
    - [PrevEntryResponse](#atomix-multiraft-atomic-indexedmap-v1-PrevEntryResponse)
    - [RemoveRequest](#atomix-multiraft-atomic-indexedmap-v1-RemoveRequest)
    - [RemoveResponse](#atomix-multiraft-atomic-indexedmap-v1-RemoveResponse)
    - [SizeRequest](#atomix-multiraft-atomic-indexedmap-v1-SizeRequest)
    - [SizeResponse](#atomix-multiraft-atomic-indexedmap-v1-SizeResponse)
    - [UpdateRequest](#atomix-multiraft-atomic-indexedmap-v1-UpdateRequest)
    - [UpdateResponse](#atomix-multiraft-atomic-indexedmap-v1-UpdateResponse)
  
    - [IndexedMap](#atomix-multiraft-atomic-indexedmap-v1-IndexedMap)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_atomic_indexedmap_v1_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/atomic/indexedmap/v1/service.proto



<a name="atomix-multiraft-atomic-indexedmap-v1-AppendRequest"></a>

### AppendRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [AppendInput](#atomix-multiraft-atomic-indexedmap-v1-AppendInput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-AppendResponse"></a>

### AppendResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [AppendOutput](#atomix-multiraft-atomic-indexedmap-v1-AppendOutput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-ClearRequest"></a>

### ClearRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [ClearInput](#atomix-multiraft-atomic-indexedmap-v1-ClearInput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-ClearResponse"></a>

### ClearResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [ClearOutput](#atomix-multiraft-atomic-indexedmap-v1-ClearOutput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-EntriesRequest"></a>

### EntriesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [EntriesInput](#atomix-multiraft-atomic-indexedmap-v1-EntriesInput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-EntriesResponse"></a>

### EntriesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [EntriesOutput](#atomix-multiraft-atomic-indexedmap-v1-EntriesOutput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-EventsRequest"></a>

### EventsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [EventsInput](#atomix-multiraft-atomic-indexedmap-v1-EventsInput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-EventsResponse"></a>

### EventsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [EventsOutput](#atomix-multiraft-atomic-indexedmap-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-FirstEntryRequest"></a>

### FirstEntryRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [FirstEntryInput](#atomix-multiraft-atomic-indexedmap-v1-FirstEntryInput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-FirstEntryResponse"></a>

### FirstEntryResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [FirstEntryOutput](#atomix-multiraft-atomic-indexedmap-v1-FirstEntryOutput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-GetRequest"></a>

### GetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [GetInput](#atomix-multiraft-atomic-indexedmap-v1-GetInput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-GetResponse"></a>

### GetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [GetOutput](#atomix-multiraft-atomic-indexedmap-v1-GetOutput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-LastEntryRequest"></a>

### LastEntryRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [LastEntryInput](#atomix-multiraft-atomic-indexedmap-v1-LastEntryInput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-LastEntryResponse"></a>

### LastEntryResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [LastEntryOutput](#atomix-multiraft-atomic-indexedmap-v1-LastEntryOutput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-NextEntryRequest"></a>

### NextEntryRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [NextEntryInput](#atomix-multiraft-atomic-indexedmap-v1-NextEntryInput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-NextEntryResponse"></a>

### NextEntryResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [NextEntryOutput](#atomix-multiraft-atomic-indexedmap-v1-NextEntryOutput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-PrevEntryRequest"></a>

### PrevEntryRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [PrevEntryInput](#atomix-multiraft-atomic-indexedmap-v1-PrevEntryInput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-PrevEntryResponse"></a>

### PrevEntryResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [PrevEntryOutput](#atomix-multiraft-atomic-indexedmap-v1-PrevEntryOutput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-RemoveRequest"></a>

### RemoveRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [RemoveInput](#atomix-multiraft-atomic-indexedmap-v1-RemoveInput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-RemoveResponse"></a>

### RemoveResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [RemoveOutput](#atomix-multiraft-atomic-indexedmap-v1-RemoveOutput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-SizeRequest"></a>

### SizeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [SizeInput](#atomix-multiraft-atomic-indexedmap-v1-SizeInput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-SizeResponse"></a>

### SizeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [SizeOutput](#atomix-multiraft-atomic-indexedmap-v1-SizeOutput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-UpdateRequest"></a>

### UpdateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [UpdateInput](#atomix-multiraft-atomic-indexedmap-v1-UpdateInput) |  |  |






<a name="atomix-multiraft-atomic-indexedmap-v1-UpdateResponse"></a>

### UpdateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [UpdateOutput](#atomix-multiraft-atomic-indexedmap-v1-UpdateOutput) |  |  |





 

 

 


<a name="atomix-multiraft-atomic-indexedmap-v1-IndexedMap"></a>

### IndexedMap
IndexedMap is a service for an indexed map primitive

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Size | [SizeRequest](#atomix-multiraft-atomic-indexedmap-v1-SizeRequest) | [SizeResponse](#atomix-multiraft-atomic-indexedmap-v1-SizeResponse) | Size returns the size of the map |
| Append | [AppendRequest](#atomix-multiraft-atomic-indexedmap-v1-AppendRequest) | [AppendResponse](#atomix-multiraft-atomic-indexedmap-v1-AppendResponse) | Append appends an entry to the map |
| Update | [UpdateRequest](#atomix-multiraft-atomic-indexedmap-v1-UpdateRequest) | [UpdateResponse](#atomix-multiraft-atomic-indexedmap-v1-UpdateResponse) | Update updates an entry in the map |
| Get | [GetRequest](#atomix-multiraft-atomic-indexedmap-v1-GetRequest) | [GetResponse](#atomix-multiraft-atomic-indexedmap-v1-GetResponse) | Get gets the entry for a key |
| FirstEntry | [FirstEntryRequest](#atomix-multiraft-atomic-indexedmap-v1-FirstEntryRequest) | [FirstEntryResponse](#atomix-multiraft-atomic-indexedmap-v1-FirstEntryResponse) | FirstEntry gets the first entry in the map |
| LastEntry | [LastEntryRequest](#atomix-multiraft-atomic-indexedmap-v1-LastEntryRequest) | [LastEntryResponse](#atomix-multiraft-atomic-indexedmap-v1-LastEntryResponse) | LastEntry gets the last entry in the map |
| PrevEntry | [PrevEntryRequest](#atomix-multiraft-atomic-indexedmap-v1-PrevEntryRequest) | [PrevEntryResponse](#atomix-multiraft-atomic-indexedmap-v1-PrevEntryResponse) | PrevEntry gets the previous entry in the map |
| NextEntry | [NextEntryRequest](#atomix-multiraft-atomic-indexedmap-v1-NextEntryRequest) | [NextEntryResponse](#atomix-multiraft-atomic-indexedmap-v1-NextEntryResponse) | NextEntry gets the next entry in the map |
| Clear | [ClearRequest](#atomix-multiraft-atomic-indexedmap-v1-ClearRequest) | [ClearResponse](#atomix-multiraft-atomic-indexedmap-v1-ClearResponse) | Clear removes all entries from the map |
| Events | [EventsRequest](#atomix-multiraft-atomic-indexedmap-v1-EventsRequest) | [EventsResponse](#atomix-multiraft-atomic-indexedmap-v1-EventsResponse) stream | Events listens for change events |
| Entries | [EntriesRequest](#atomix-multiraft-atomic-indexedmap-v1-EntriesRequest) | [EntriesResponse](#atomix-multiraft-atomic-indexedmap-v1-EntriesResponse) stream | Entries lists all entries in the map |

 



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

