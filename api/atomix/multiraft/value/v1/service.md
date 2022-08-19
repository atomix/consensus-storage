# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/value/v1/service.proto](#atomix_multiraft_value_v1_service-proto)
    - [DeleteRequest](#atomix-multiraft-value-v1-DeleteRequest)
    - [DeleteResponse](#atomix-multiraft-value-v1-DeleteResponse)
    - [EventsRequest](#atomix-multiraft-value-v1-EventsRequest)
    - [EventsResponse](#atomix-multiraft-value-v1-EventsResponse)
    - [GetRequest](#atomix-multiraft-value-v1-GetRequest)
    - [GetResponse](#atomix-multiraft-value-v1-GetResponse)
    - [InsertRequest](#atomix-multiraft-value-v1-InsertRequest)
    - [InsertResponse](#atomix-multiraft-value-v1-InsertResponse)
    - [SetRequest](#atomix-multiraft-value-v1-SetRequest)
    - [SetResponse](#atomix-multiraft-value-v1-SetResponse)
    - [UpdateRequest](#atomix-multiraft-value-v1-UpdateRequest)
    - [UpdateResponse](#atomix-multiraft-value-v1-UpdateResponse)
    - [WatchRequest](#atomix-multiraft-value-v1-WatchRequest)
    - [WatchResponse](#atomix-multiraft-value-v1-WatchResponse)
  
    - [Value](#atomix-multiraft-value-v1-Value)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_value_v1_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/value/v1/service.proto



<a name="atomix-multiraft-value-v1-DeleteRequest"></a>

### DeleteRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [DeleteInput](#atomix-multiraft-value-v1-DeleteInput) |  |  |






<a name="atomix-multiraft-value-v1-DeleteResponse"></a>

### DeleteResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [DeleteOutput](#atomix-multiraft-value-v1-DeleteOutput) |  |  |






<a name="atomix-multiraft-value-v1-EventsRequest"></a>

### EventsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [EventsInput](#atomix-multiraft-value-v1-EventsInput) |  |  |






<a name="atomix-multiraft-value-v1-EventsResponse"></a>

### EventsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [EventsOutput](#atomix-multiraft-value-v1-EventsOutput) |  |  |






<a name="atomix-multiraft-value-v1-GetRequest"></a>

### GetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [GetInput](#atomix-multiraft-value-v1-GetInput) |  |  |






<a name="atomix-multiraft-value-v1-GetResponse"></a>

### GetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [GetOutput](#atomix-multiraft-value-v1-GetOutput) |  |  |






<a name="atomix-multiraft-value-v1-InsertRequest"></a>

### InsertRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [InsertInput](#atomix-multiraft-value-v1-InsertInput) |  |  |






<a name="atomix-multiraft-value-v1-InsertResponse"></a>

### InsertResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [InsertOutput](#atomix-multiraft-value-v1-InsertOutput) |  |  |






<a name="atomix-multiraft-value-v1-SetRequest"></a>

### SetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [SetInput](#atomix-multiraft-value-v1-SetInput) |  |  |






<a name="atomix-multiraft-value-v1-SetResponse"></a>

### SetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [SetOutput](#atomix-multiraft-value-v1-SetOutput) |  |  |






<a name="atomix-multiraft-value-v1-UpdateRequest"></a>

### UpdateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [UpdateInput](#atomix-multiraft-value-v1-UpdateInput) |  |  |






<a name="atomix-multiraft-value-v1-UpdateResponse"></a>

### UpdateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [UpdateOutput](#atomix-multiraft-value-v1-UpdateOutput) |  |  |






<a name="atomix-multiraft-value-v1-WatchRequest"></a>

### WatchRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [WatchInput](#atomix-multiraft-value-v1-WatchInput) |  |  |






<a name="atomix-multiraft-value-v1-WatchResponse"></a>

### WatchResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [WatchOutput](#atomix-multiraft-value-v1-WatchOutput) |  |  |





 

 

 


<a name="atomix-multiraft-value-v1-Value"></a>

### Value
Value is a service for a map primitive

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Set | [SetRequest](#atomix-multiraft-value-v1-SetRequest) | [SetResponse](#atomix-multiraft-value-v1-SetResponse) | Set sets the value |
| Insert | [InsertRequest](#atomix-multiraft-value-v1-InsertRequest) | [InsertResponse](#atomix-multiraft-value-v1-InsertResponse) | Insert inserts the value |
| Update | [UpdateRequest](#atomix-multiraft-value-v1-UpdateRequest) | [UpdateResponse](#atomix-multiraft-value-v1-UpdateResponse) | Update updates the value |
| Get | [GetRequest](#atomix-multiraft-value-v1-GetRequest) | [GetResponse](#atomix-multiraft-value-v1-GetResponse) | Get gets the value |
| Delete | [DeleteRequest](#atomix-multiraft-value-v1-DeleteRequest) | [DeleteResponse](#atomix-multiraft-value-v1-DeleteResponse) | Delete deletes the value |
| Watch | [WatchRequest](#atomix-multiraft-value-v1-WatchRequest) | [WatchResponse](#atomix-multiraft-value-v1-WatchResponse) stream | Watch watches the value |
| Events | [EventsRequest](#atomix-multiraft-value-v1-EventsRequest) | [EventsResponse](#atomix-multiraft-value-v1-EventsResponse) stream | Events watches for value change events |

 



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

