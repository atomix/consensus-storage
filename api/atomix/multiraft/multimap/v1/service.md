# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/multimap/v1/service.proto](#atomix_multiraft_multimap_v1_service-proto)
    - [ClearRequest](#atomix-multiraft-multimap-v1-ClearRequest)
    - [ClearResponse](#atomix-multiraft-multimap-v1-ClearResponse)
    - [ContainsRequest](#atomix-multiraft-multimap-v1-ContainsRequest)
    - [ContainsResponse](#atomix-multiraft-multimap-v1-ContainsResponse)
    - [EntriesRequest](#atomix-multiraft-multimap-v1-EntriesRequest)
    - [EntriesResponse](#atomix-multiraft-multimap-v1-EntriesResponse)
    - [EventsRequest](#atomix-multiraft-multimap-v1-EventsRequest)
    - [EventsResponse](#atomix-multiraft-multimap-v1-EventsResponse)
    - [GetRequest](#atomix-multiraft-multimap-v1-GetRequest)
    - [GetResponse](#atomix-multiraft-multimap-v1-GetResponse)
    - [PutAllRequest](#atomix-multiraft-multimap-v1-PutAllRequest)
    - [PutAllResponse](#atomix-multiraft-multimap-v1-PutAllResponse)
    - [PutEntriesRequest](#atomix-multiraft-multimap-v1-PutEntriesRequest)
    - [PutEntriesResponse](#atomix-multiraft-multimap-v1-PutEntriesResponse)
    - [PutRequest](#atomix-multiraft-multimap-v1-PutRequest)
    - [PutResponse](#atomix-multiraft-multimap-v1-PutResponse)
    - [RemoveAllRequest](#atomix-multiraft-multimap-v1-RemoveAllRequest)
    - [RemoveAllResponse](#atomix-multiraft-multimap-v1-RemoveAllResponse)
    - [RemoveEntriesRequest](#atomix-multiraft-multimap-v1-RemoveEntriesRequest)
    - [RemoveEntriesResponse](#atomix-multiraft-multimap-v1-RemoveEntriesResponse)
    - [RemoveRequest](#atomix-multiraft-multimap-v1-RemoveRequest)
    - [RemoveResponse](#atomix-multiraft-multimap-v1-RemoveResponse)
    - [ReplaceRequest](#atomix-multiraft-multimap-v1-ReplaceRequest)
    - [ReplaceResponse](#atomix-multiraft-multimap-v1-ReplaceResponse)
    - [SizeRequest](#atomix-multiraft-multimap-v1-SizeRequest)
    - [SizeResponse](#atomix-multiraft-multimap-v1-SizeResponse)
  
    - [MultiMap](#atomix-multiraft-multimap-v1-MultiMap)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_multimap_v1_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/multimap/v1/service.proto



<a name="atomix-multiraft-multimap-v1-ClearRequest"></a>

### ClearRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [ClearInput](#atomix-multiraft-multimap-v1-ClearInput) |  |  |






<a name="atomix-multiraft-multimap-v1-ClearResponse"></a>

### ClearResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [ClearOutput](#atomix-multiraft-multimap-v1-ClearOutput) |  |  |






<a name="atomix-multiraft-multimap-v1-ContainsRequest"></a>

### ContainsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [ContainsInput](#atomix-multiraft-multimap-v1-ContainsInput) |  |  |






<a name="atomix-multiraft-multimap-v1-ContainsResponse"></a>

### ContainsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [ContainsOutput](#atomix-multiraft-multimap-v1-ContainsOutput) |  |  |






<a name="atomix-multiraft-multimap-v1-EntriesRequest"></a>

### EntriesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [EntriesInput](#atomix-multiraft-multimap-v1-EntriesInput) |  |  |






<a name="atomix-multiraft-multimap-v1-EntriesResponse"></a>

### EntriesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [EntriesOutput](#atomix-multiraft-multimap-v1-EntriesOutput) |  |  |






<a name="atomix-multiraft-multimap-v1-EventsRequest"></a>

### EventsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [EventsInput](#atomix-multiraft-multimap-v1-EventsInput) |  |  |






<a name="atomix-multiraft-multimap-v1-EventsResponse"></a>

### EventsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [EventsOutput](#atomix-multiraft-multimap-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-multimap-v1-GetRequest"></a>

### GetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [GetInput](#atomix-multiraft-multimap-v1-GetInput) |  |  |






<a name="atomix-multiraft-multimap-v1-GetResponse"></a>

### GetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [GetOutput](#atomix-multiraft-multimap-v1-GetOutput) |  |  |






<a name="atomix-multiraft-multimap-v1-PutAllRequest"></a>

### PutAllRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [PutAllInput](#atomix-multiraft-multimap-v1-PutAllInput) |  |  |






<a name="atomix-multiraft-multimap-v1-PutAllResponse"></a>

### PutAllResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [PutAllOutput](#atomix-multiraft-multimap-v1-PutAllOutput) |  |  |






<a name="atomix-multiraft-multimap-v1-PutEntriesRequest"></a>

### PutEntriesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [PutEntriesInput](#atomix-multiraft-multimap-v1-PutEntriesInput) |  |  |






<a name="atomix-multiraft-multimap-v1-PutEntriesResponse"></a>

### PutEntriesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [PutEntriesOutput](#atomix-multiraft-multimap-v1-PutEntriesOutput) |  |  |






<a name="atomix-multiraft-multimap-v1-PutRequest"></a>

### PutRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [PutInput](#atomix-multiraft-multimap-v1-PutInput) |  |  |






<a name="atomix-multiraft-multimap-v1-PutResponse"></a>

### PutResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [PutOutput](#atomix-multiraft-multimap-v1-PutOutput) |  |  |






<a name="atomix-multiraft-multimap-v1-RemoveAllRequest"></a>

### RemoveAllRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [RemoveAllInput](#atomix-multiraft-multimap-v1-RemoveAllInput) |  |  |






<a name="atomix-multiraft-multimap-v1-RemoveAllResponse"></a>

### RemoveAllResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [RemoveAllOutput](#atomix-multiraft-multimap-v1-RemoveAllOutput) |  |  |






<a name="atomix-multiraft-multimap-v1-RemoveEntriesRequest"></a>

### RemoveEntriesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [RemoveEntriesInput](#atomix-multiraft-multimap-v1-RemoveEntriesInput) |  |  |






<a name="atomix-multiraft-multimap-v1-RemoveEntriesResponse"></a>

### RemoveEntriesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [RemoveEntriesOutput](#atomix-multiraft-multimap-v1-RemoveEntriesOutput) |  |  |






<a name="atomix-multiraft-multimap-v1-RemoveRequest"></a>

### RemoveRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [RemoveInput](#atomix-multiraft-multimap-v1-RemoveInput) |  |  |






<a name="atomix-multiraft-multimap-v1-RemoveResponse"></a>

### RemoveResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [RemoveOutput](#atomix-multiraft-multimap-v1-RemoveOutput) |  |  |






<a name="atomix-multiraft-multimap-v1-ReplaceRequest"></a>

### ReplaceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [ReplaceInput](#atomix-multiraft-multimap-v1-ReplaceInput) |  |  |






<a name="atomix-multiraft-multimap-v1-ReplaceResponse"></a>

### ReplaceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [ReplaceOutput](#atomix-multiraft-multimap-v1-ReplaceOutput) |  |  |






<a name="atomix-multiraft-multimap-v1-SizeRequest"></a>

### SizeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [SizeInput](#atomix-multiraft-multimap-v1-SizeInput) |  |  |






<a name="atomix-multiraft-multimap-v1-SizeResponse"></a>

### SizeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [SizeOutput](#atomix-multiraft-multimap-v1-SizeOutput) |  |  |





 

 

 


<a name="atomix-multiraft-multimap-v1-MultiMap"></a>

### MultiMap
MultiMap is a service for a multimap primitive

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Size | [SizeRequest](#atomix-multiraft-multimap-v1-SizeRequest) | [SizeResponse](#atomix-multiraft-multimap-v1-SizeResponse) | Size returns the size of the multimap |
| Put | [PutRequest](#atomix-multiraft-multimap-v1-PutRequest) | [PutResponse](#atomix-multiraft-multimap-v1-PutResponse) | Put puts an entry into the multimap |
| PutAll | [PutAllRequest](#atomix-multiraft-multimap-v1-PutAllRequest) | [PutAllResponse](#atomix-multiraft-multimap-v1-PutAllResponse) | PutAll puts an entry into the multimap |
| PutEntries | [PutEntriesRequest](#atomix-multiraft-multimap-v1-PutEntriesRequest) | [PutEntriesResponse](#atomix-multiraft-multimap-v1-PutEntriesResponse) | PutEntries puts a set of entries into the multimap |
| Replace | [ReplaceRequest](#atomix-multiraft-multimap-v1-ReplaceRequest) | [ReplaceResponse](#atomix-multiraft-multimap-v1-ReplaceResponse) | Replace replaces an entry in the multimap |
| Contains | [ContainsRequest](#atomix-multiraft-multimap-v1-ContainsRequest) | [ContainsResponse](#atomix-multiraft-multimap-v1-ContainsResponse) | Contains checks if the multimap contains an entry |
| Get | [GetRequest](#atomix-multiraft-multimap-v1-GetRequest) | [GetResponse](#atomix-multiraft-multimap-v1-GetResponse) | Get gets the entry for a key |
| Remove | [RemoveRequest](#atomix-multiraft-multimap-v1-RemoveRequest) | [RemoveResponse](#atomix-multiraft-multimap-v1-RemoveResponse) | Remove removes an entry from the multimap |
| RemoveAll | [RemoveAllRequest](#atomix-multiraft-multimap-v1-RemoveAllRequest) | [RemoveAllResponse](#atomix-multiraft-multimap-v1-RemoveAllResponse) | RemoveAll removes a key from the multimap |
| RemoveEntries | [RemoveEntriesRequest](#atomix-multiraft-multimap-v1-RemoveEntriesRequest) | [RemoveEntriesResponse](#atomix-multiraft-multimap-v1-RemoveEntriesResponse) | RemoveEntries removes a set of entries from the multimap |
| Clear | [ClearRequest](#atomix-multiraft-multimap-v1-ClearRequest) | [ClearResponse](#atomix-multiraft-multimap-v1-ClearResponse) | Clear removes all entries from the multimap |
| Events | [EventsRequest](#atomix-multiraft-multimap-v1-EventsRequest) | [EventsResponse](#atomix-multiraft-multimap-v1-EventsResponse) stream | Events listens for change events |
| Entries | [EntriesRequest](#atomix-multiraft-multimap-v1-EntriesRequest) | [EntriesResponse](#atomix-multiraft-multimap-v1-EntriesResponse) stream | Entries lists all entries in the multimap |

 



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

