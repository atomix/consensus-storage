# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/v1/headers.proto](#atomix_multiraft_v1_headers-proto)
    - [CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders)
    - [CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders)
    - [OperationRequestHeaders](#atomix-multiraft-v1-OperationRequestHeaders)
    - [OperationResponseHeaders](#atomix-multiraft-v1-OperationResponseHeaders)
    - [PartitionRequestHeaders](#atomix-multiraft-v1-PartitionRequestHeaders)
    - [PartitionResponseHeaders](#atomix-multiraft-v1-PartitionResponseHeaders)
    - [PrimitiveRequestHeaders](#atomix-multiraft-v1-PrimitiveRequestHeaders)
    - [PrimitiveResponseHeaders](#atomix-multiraft-v1-PrimitiveResponseHeaders)
    - [QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders)
    - [QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders)
    - [SessionRequestHeaders](#atomix-multiraft-v1-SessionRequestHeaders)
    - [SessionResponseHeaders](#atomix-multiraft-v1-SessionResponseHeaders)
  
    - [OperationResponseHeaders.Status](#atomix-multiraft-v1-OperationResponseHeaders-Status)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_v1_headers-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/v1/headers.proto



<a name="atomix-multiraft-v1-CommandRequestHeaders"></a>

### CommandRequestHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operation | [OperationRequestHeaders](#atomix-multiraft-v1-OperationRequestHeaders) |  |  |
| sequence_num | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-CommandResponseHeaders"></a>

### CommandResponseHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operation | [OperationResponseHeaders](#atomix-multiraft-v1-OperationResponseHeaders) |  |  |
| output_sequence_num | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-OperationRequestHeaders"></a>

### OperationRequestHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session | [PrimitiveRequestHeaders](#atomix-multiraft-v1-PrimitiveRequestHeaders) |  |  |






<a name="atomix-multiraft-v1-OperationResponseHeaders"></a>

### OperationResponseHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session | [PrimitiveResponseHeaders](#atomix-multiraft-v1-PrimitiveResponseHeaders) |  |  |
| status | [OperationResponseHeaders.Status](#atomix-multiraft-v1-OperationResponseHeaders-Status) |  |  |
| message | [string](#string) |  |  |






<a name="atomix-multiraft-v1-PartitionRequestHeaders"></a>

### PartitionRequestHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| partition_id | [uint32](#uint32) |  |  |






<a name="atomix-multiraft-v1-PartitionResponseHeaders"></a>

### PartitionResponseHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-PrimitiveRequestHeaders"></a>

### PrimitiveRequestHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session | [SessionRequestHeaders](#atomix-multiraft-v1-SessionRequestHeaders) |  |  |
| primitive_id | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-PrimitiveResponseHeaders"></a>

### PrimitiveResponseHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session | [SessionResponseHeaders](#atomix-multiraft-v1-SessionResponseHeaders) |  |  |






<a name="atomix-multiraft-v1-QueryRequestHeaders"></a>

### QueryRequestHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operation | [OperationRequestHeaders](#atomix-multiraft-v1-OperationRequestHeaders) |  |  |
| max_received_index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-QueryResponseHeaders"></a>

### QueryResponseHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operation | [OperationResponseHeaders](#atomix-multiraft-v1-OperationResponseHeaders) |  |  |






<a name="atomix-multiraft-v1-SessionRequestHeaders"></a>

### SessionRequestHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| partition | [PartitionRequestHeaders](#atomix-multiraft-v1-PartitionRequestHeaders) |  |  |
| session_id | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-SessionResponseHeaders"></a>

### SessionResponseHeaders



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| partition | [PartitionResponseHeaders](#atomix-multiraft-v1-PartitionResponseHeaders) |  |  |





 


<a name="atomix-multiraft-v1-OperationResponseHeaders-Status"></a>

### OperationResponseHeaders.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| OK | 1 |  |
| ERROR | 2 |  |
| CANCELED | 3 |  |
| NOT_FOUND | 4 |  |
| ALREADY_EXISTS | 5 |  |
| UNAUTHORIZED | 6 |  |
| FORBIDDEN | 7 |  |
| CONFLICT | 8 |  |
| INVALID | 9 |  |
| UNAVAILABLE | 10 |  |
| NOT_SUPPORTED | 11 |  |
| TIMEOUT | 12 |  |
| INTERNAL | 13 |  |
| FAULT | 14 |  |


 

 

 



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

