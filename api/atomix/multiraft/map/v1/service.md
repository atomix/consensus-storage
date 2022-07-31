# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/map/v1/service.proto](#atomix_multiraft_map_v1_service-proto)
    - [ClearRequest](#atomix-multiraft-map-v1-ClearRequest)
    - [ClearResponse](#atomix-multiraft-map-v1-ClearResponse)
    - [EntriesRequest](#atomix-multiraft-map-v1-EntriesRequest)
    - [EntriesResponse](#atomix-multiraft-map-v1-EntriesResponse)
    - [EventsRequest](#atomix-multiraft-map-v1-EventsRequest)
    - [EventsResponse](#atomix-multiraft-map-v1-EventsResponse)
    - [GetRequest](#atomix-multiraft-map-v1-GetRequest)
    - [GetResponse](#atomix-multiraft-map-v1-GetResponse)
    - [PutRequest](#atomix-multiraft-map-v1-PutRequest)
    - [PutResponse](#atomix-multiraft-map-v1-PutResponse)
    - [RemoveRequest](#atomix-multiraft-map-v1-RemoveRequest)
    - [RemoveResponse](#atomix-multiraft-map-v1-RemoveResponse)
    - [SizeRequest](#atomix-multiraft-map-v1-SizeRequest)
    - [SizeResponse](#atomix-multiraft-map-v1-SizeResponse)
  
    - [Map](#atomix-multiraft-map-v1-Map)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_map_v1_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/map/v1/service.proto



<a name="atomix-multiraft-map-v1-ClearRequest"></a>

### ClearRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [ClearInput](#atomix-multiraft-map-v1-ClearInput) |  |  |






<a name="atomix-multiraft-map-v1-ClearResponse"></a>

### ClearResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [ClearOutput](#atomix-multiraft-map-v1-ClearOutput) |  |  |






<a name="atomix-multiraft-map-v1-EntriesRequest"></a>

### EntriesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [EntriesInput](#atomix-multiraft-map-v1-EntriesInput) |  |  |






<a name="atomix-multiraft-map-v1-EntriesResponse"></a>

### EntriesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [EntriesOutput](#atomix-multiraft-map-v1-EntriesOutput) |  |  |






<a name="atomix-multiraft-map-v1-EventsRequest"></a>

### EventsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [EventsInput](#atomix-multiraft-map-v1-EventsInput) |  |  |






<a name="atomix-multiraft-map-v1-EventsResponse"></a>

### EventsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [EventsOutput](#atomix-multiraft-map-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-map-v1-GetRequest"></a>

### GetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [GetInput](#atomix-multiraft-map-v1-GetInput) |  |  |






<a name="atomix-multiraft-map-v1-GetResponse"></a>

### GetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [GetOutput](#atomix-multiraft-map-v1-GetOutput) |  |  |






<a name="atomix-multiraft-map-v1-PutRequest"></a>

### PutRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [PutInput](#atomix-multiraft-map-v1-PutInput) |  |  |






<a name="atomix-multiraft-map-v1-PutResponse"></a>

### PutResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [PutOutput](#atomix-multiraft-map-v1-PutOutput) |  |  |






<a name="atomix-multiraft-map-v1-RemoveRequest"></a>

### RemoveRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [RemoveInput](#atomix-multiraft-map-v1-RemoveInput) |  |  |






<a name="atomix-multiraft-map-v1-RemoveResponse"></a>

### RemoveResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [RemoveOutput](#atomix-multiraft-map-v1-RemoveOutput) |  |  |






<a name="atomix-multiraft-map-v1-SizeRequest"></a>

### SizeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [SizeInput](#atomix-multiraft-map-v1-SizeInput) |  |  |






<a name="atomix-multiraft-map-v1-SizeResponse"></a>

### SizeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [SizeOutput](#atomix-multiraft-map-v1-SizeOutput) |  |  |





 

 

 


<a name="atomix-multiraft-map-v1-Map"></a>

### Map
Map is a service for a map primitive

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Size | [SizeRequest](#atomix-multiraft-map-v1-SizeRequest) | [SizeResponse](#atomix-multiraft-map-v1-SizeResponse) | Size returns the size of the map |
| Put | [PutRequest](#atomix-multiraft-map-v1-PutRequest) | [PutResponse](#atomix-multiraft-map-v1-PutResponse) | Put puts an entry into the map |
| Get | [GetRequest](#atomix-multiraft-map-v1-GetRequest) | [GetResponse](#atomix-multiraft-map-v1-GetResponse) | Get gets the entry for a key |
| Remove | [RemoveRequest](#atomix-multiraft-map-v1-RemoveRequest) | [RemoveResponse](#atomix-multiraft-map-v1-RemoveResponse) | Remove removes an entry from the map |
| Clear | [ClearRequest](#atomix-multiraft-map-v1-ClearRequest) | [ClearResponse](#atomix-multiraft-map-v1-ClearResponse) | Clear removes all entries from the map |
| Events | [EventsRequest](#atomix-multiraft-map-v1-EventsRequest) | [EventsResponse](#atomix-multiraft-map-v1-EventsResponse) stream | Events listens for change events |
| Entries | [EntriesRequest](#atomix-multiraft-map-v1-EntriesRequest) | [EntriesResponse](#atomix-multiraft-map-v1-EntriesResponse) stream | Entries lists all entries in the map |

 



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

