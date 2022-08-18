# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/list/v1/service.proto](#atomix_multiraft_list_v1_service-proto)
    - [AppendRequest](#atomix-multiraft-list-v1-AppendRequest)
    - [AppendResponse](#atomix-multiraft-list-v1-AppendResponse)
    - [ClearRequest](#atomix-multiraft-list-v1-ClearRequest)
    - [ClearResponse](#atomix-multiraft-list-v1-ClearResponse)
    - [ContainsRequest](#atomix-multiraft-list-v1-ContainsRequest)
    - [ContainsResponse](#atomix-multiraft-list-v1-ContainsResponse)
    - [EventsRequest](#atomix-multiraft-list-v1-EventsRequest)
    - [EventsResponse](#atomix-multiraft-list-v1-EventsResponse)
    - [GetRequest](#atomix-multiraft-list-v1-GetRequest)
    - [GetResponse](#atomix-multiraft-list-v1-GetResponse)
    - [InsertRequest](#atomix-multiraft-list-v1-InsertRequest)
    - [InsertResponse](#atomix-multiraft-list-v1-InsertResponse)
    - [ItemsRequest](#atomix-multiraft-list-v1-ItemsRequest)
    - [ItemsResponse](#atomix-multiraft-list-v1-ItemsResponse)
    - [RemoveRequest](#atomix-multiraft-list-v1-RemoveRequest)
    - [RemoveResponse](#atomix-multiraft-list-v1-RemoveResponse)
    - [SetRequest](#atomix-multiraft-list-v1-SetRequest)
    - [SetResponse](#atomix-multiraft-list-v1-SetResponse)
    - [SizeRequest](#atomix-multiraft-list-v1-SizeRequest)
    - [SizeResponse](#atomix-multiraft-list-v1-SizeResponse)
  
    - [List](#atomix-multiraft-list-v1-List)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_list_v1_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/list/v1/service.proto



<a name="atomix-multiraft-list-v1-AppendRequest"></a>

### AppendRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [AppendOutput](#atomix-multiraft-list-v1-AppendOutput) |  |  |






<a name="atomix-multiraft-list-v1-AppendResponse"></a>

### AppendResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [AppendOutput](#atomix-multiraft-list-v1-AppendOutput) |  |  |






<a name="atomix-multiraft-list-v1-ClearRequest"></a>

### ClearRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [ClearOutput](#atomix-multiraft-list-v1-ClearOutput) |  |  |






<a name="atomix-multiraft-list-v1-ClearResponse"></a>

### ClearResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [ClearOutput](#atomix-multiraft-list-v1-ClearOutput) |  |  |






<a name="atomix-multiraft-list-v1-ContainsRequest"></a>

### ContainsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [ContainsOutput](#atomix-multiraft-list-v1-ContainsOutput) |  |  |






<a name="atomix-multiraft-list-v1-ContainsResponse"></a>

### ContainsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [ContainsOutput](#atomix-multiraft-list-v1-ContainsOutput) |  |  |






<a name="atomix-multiraft-list-v1-EventsRequest"></a>

### EventsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [EventsOutput](#atomix-multiraft-list-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-list-v1-EventsResponse"></a>

### EventsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [EventsOutput](#atomix-multiraft-list-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-list-v1-GetRequest"></a>

### GetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [GetInput](#atomix-multiraft-list-v1-GetInput) |  |  |






<a name="atomix-multiraft-list-v1-GetResponse"></a>

### GetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [GetOutput](#atomix-multiraft-list-v1-GetOutput) |  |  |






<a name="atomix-multiraft-list-v1-InsertRequest"></a>

### InsertRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [InsertOutput](#atomix-multiraft-list-v1-InsertOutput) |  |  |






<a name="atomix-multiraft-list-v1-InsertResponse"></a>

### InsertResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [InsertOutput](#atomix-multiraft-list-v1-InsertOutput) |  |  |






<a name="atomix-multiraft-list-v1-ItemsRequest"></a>

### ItemsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [ItemsInput](#atomix-multiraft-list-v1-ItemsInput) |  |  |






<a name="atomix-multiraft-list-v1-ItemsResponse"></a>

### ItemsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [ItemsOutput](#atomix-multiraft-list-v1-ItemsOutput) |  |  |






<a name="atomix-multiraft-list-v1-RemoveRequest"></a>

### RemoveRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [RemoveOutput](#atomix-multiraft-list-v1-RemoveOutput) |  |  |






<a name="atomix-multiraft-list-v1-RemoveResponse"></a>

### RemoveResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [RemoveOutput](#atomix-multiraft-list-v1-RemoveOutput) |  |  |






<a name="atomix-multiraft-list-v1-SetRequest"></a>

### SetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [SetOutput](#atomix-multiraft-list-v1-SetOutput) |  |  |






<a name="atomix-multiraft-list-v1-SetResponse"></a>

### SetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [SetOutput](#atomix-multiraft-list-v1-SetOutput) |  |  |






<a name="atomix-multiraft-list-v1-SizeRequest"></a>

### SizeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [SizeInput](#atomix-multiraft-list-v1-SizeInput) |  |  |






<a name="atomix-multiraft-list-v1-SizeResponse"></a>

### SizeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [SizeOutput](#atomix-multiraft-list-v1-SizeOutput) |  |  |





 

 

 


<a name="atomix-multiraft-list-v1-List"></a>

### List
List is a service for a list primitive

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Size | [SizeRequest](#atomix-multiraft-list-v1-SizeRequest) | [SizeResponse](#atomix-multiraft-list-v1-SizeResponse) | Size gets the number of elements in the list |
| Append | [AppendRequest](#atomix-multiraft-list-v1-AppendRequest) | [AppendResponse](#atomix-multiraft-list-v1-AppendResponse) | Append appends a value to the list |
| Insert | [InsertRequest](#atomix-multiraft-list-v1-InsertRequest) | [InsertResponse](#atomix-multiraft-list-v1-InsertResponse) | Insert inserts a value at a specific index in the list |
| Get | [GetRequest](#atomix-multiraft-list-v1-GetRequest) | [GetResponse](#atomix-multiraft-list-v1-GetResponse) | Get gets the value at an index in the list |
| Set | [SetRequest](#atomix-multiraft-list-v1-SetRequest) | [SetResponse](#atomix-multiraft-list-v1-SetResponse) | Set sets the value at an index in the list |
| Remove | [RemoveRequest](#atomix-multiraft-list-v1-RemoveRequest) | [RemoveResponse](#atomix-multiraft-list-v1-RemoveResponse) | Remove removes an element from the list |
| Clear | [ClearRequest](#atomix-multiraft-list-v1-ClearRequest) | [ClearResponse](#atomix-multiraft-list-v1-ClearResponse) | Clear removes all elements from the list |
| Events | [EventsRequest](#atomix-multiraft-list-v1-EventsRequest) | [EventsResponse](#atomix-multiraft-list-v1-EventsResponse) stream | Events listens for change events |
| Items | [ItemsRequest](#atomix-multiraft-list-v1-ItemsRequest) | [ItemsResponse](#atomix-multiraft-list-v1-ItemsResponse) stream | Items streams all items in the list |

 



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

