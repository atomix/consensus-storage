# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/set/v1/service.proto](#atomix_multiraft_set_v1_service-proto)
    - [AddRequest](#atomix-multiraft-set-v1-AddRequest)
    - [AddResponse](#atomix-multiraft-set-v1-AddResponse)
    - [ClearRequest](#atomix-multiraft-set-v1-ClearRequest)
    - [ClearResponse](#atomix-multiraft-set-v1-ClearResponse)
    - [ContainsRequest](#atomix-multiraft-set-v1-ContainsRequest)
    - [ContainsResponse](#atomix-multiraft-set-v1-ContainsResponse)
    - [ElementsRequest](#atomix-multiraft-set-v1-ElementsRequest)
    - [ElementsResponse](#atomix-multiraft-set-v1-ElementsResponse)
    - [EventsRequest](#atomix-multiraft-set-v1-EventsRequest)
    - [EventsResponse](#atomix-multiraft-set-v1-EventsResponse)
    - [RemoveRequest](#atomix-multiraft-set-v1-RemoveRequest)
    - [RemoveResponse](#atomix-multiraft-set-v1-RemoveResponse)
    - [SizeRequest](#atomix-multiraft-set-v1-SizeRequest)
    - [SizeResponse](#atomix-multiraft-set-v1-SizeResponse)
  
    - [Set](#atomix-multiraft-set-v1-Set)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_set_v1_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/set/v1/service.proto



<a name="atomix-multiraft-set-v1-AddRequest"></a>

### AddRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [AddInput](#atomix-multiraft-set-v1-AddInput) |  |  |






<a name="atomix-multiraft-set-v1-AddResponse"></a>

### AddResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [AddOutput](#atomix-multiraft-set-v1-AddOutput) |  |  |






<a name="atomix-multiraft-set-v1-ClearRequest"></a>

### ClearRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [ClearInput](#atomix-multiraft-set-v1-ClearInput) |  |  |






<a name="atomix-multiraft-set-v1-ClearResponse"></a>

### ClearResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [ClearOutput](#atomix-multiraft-set-v1-ClearOutput) |  |  |






<a name="atomix-multiraft-set-v1-ContainsRequest"></a>

### ContainsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [ContainsInput](#atomix-multiraft-set-v1-ContainsInput) |  |  |






<a name="atomix-multiraft-set-v1-ContainsResponse"></a>

### ContainsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [ContainsOutput](#atomix-multiraft-set-v1-ContainsOutput) |  |  |






<a name="atomix-multiraft-set-v1-ElementsRequest"></a>

### ElementsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [ElementsInput](#atomix-multiraft-set-v1-ElementsInput) |  |  |






<a name="atomix-multiraft-set-v1-ElementsResponse"></a>

### ElementsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [ElementsOutput](#atomix-multiraft-set-v1-ElementsOutput) |  |  |






<a name="atomix-multiraft-set-v1-EventsRequest"></a>

### EventsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [EventsInput](#atomix-multiraft-set-v1-EventsInput) |  |  |






<a name="atomix-multiraft-set-v1-EventsResponse"></a>

### EventsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [EventsOutput](#atomix-multiraft-set-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-set-v1-RemoveRequest"></a>

### RemoveRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [RemoveInput](#atomix-multiraft-set-v1-RemoveInput) |  |  |






<a name="atomix-multiraft-set-v1-RemoveResponse"></a>

### RemoveResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [RemoveOutput](#atomix-multiraft-set-v1-RemoveOutput) |  |  |






<a name="atomix-multiraft-set-v1-SizeRequest"></a>

### SizeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [SizeInput](#atomix-multiraft-set-v1-SizeInput) |  |  |






<a name="atomix-multiraft-set-v1-SizeResponse"></a>

### SizeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [SizeOutput](#atomix-multiraft-set-v1-SizeOutput) |  |  |





 

 

 


<a name="atomix-multiraft-set-v1-Set"></a>

### Set
Set is a service for a set primitive

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Size | [SizeRequest](#atomix-multiraft-set-v1-SizeRequest) | [SizeResponse](#atomix-multiraft-set-v1-SizeResponse) | Size gets the number of elements in the set |
| Contains | [ContainsRequest](#atomix-multiraft-set-v1-ContainsRequest) | [ContainsResponse](#atomix-multiraft-set-v1-ContainsResponse) | Contains returns whether the set contains a value |
| Add | [AddRequest](#atomix-multiraft-set-v1-AddRequest) | [AddResponse](#atomix-multiraft-set-v1-AddResponse) | Add adds a value to the set |
| Remove | [RemoveRequest](#atomix-multiraft-set-v1-RemoveRequest) | [RemoveResponse](#atomix-multiraft-set-v1-RemoveResponse) | Remove removes a value from the set |
| Clear | [ClearRequest](#atomix-multiraft-set-v1-ClearRequest) | [ClearResponse](#atomix-multiraft-set-v1-ClearResponse) | Clear removes all values from the set |
| Events | [EventsRequest](#atomix-multiraft-set-v1-EventsRequest) | [EventsResponse](#atomix-multiraft-set-v1-EventsResponse) stream | Events listens for set change events |
| Elements | [ElementsRequest](#atomix-multiraft-set-v1-ElementsRequest) | [ElementsResponse](#atomix-multiraft-set-v1-ElementsResponse) stream | Elements lists all elements in the set |

 



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

