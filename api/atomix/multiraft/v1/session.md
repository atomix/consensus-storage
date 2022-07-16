# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/v1/session.proto](#atomix_multiraft_v1_session-proto)
    - [ClosePrimitiveRequest](#atomix-multiraft-v1-ClosePrimitiveRequest)
    - [ClosePrimitiveResponse](#atomix-multiraft-v1-ClosePrimitiveResponse)
    - [CreatePrimitiveRequest](#atomix-multiraft-v1-CreatePrimitiveRequest)
    - [CreatePrimitiveResponse](#atomix-multiraft-v1-CreatePrimitiveResponse)
  
    - [Session](#atomix-multiraft-v1-Session)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_v1_session-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/v1/session.proto



<a name="atomix-multiraft-v1-ClosePrimitiveRequest"></a>

### ClosePrimitiveRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [SessionRequestHeaders](#atomix-multiraft-v1-SessionRequestHeaders) |  |  |
| input | [ClosePrimitiveInput](#atomix-multiraft-v1-ClosePrimitiveInput) |  |  |






<a name="atomix-multiraft-v1-ClosePrimitiveResponse"></a>

### ClosePrimitiveResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [SessionResponseHeaders](#atomix-multiraft-v1-SessionResponseHeaders) |  |  |
| output | [ClosePrimitiveOutput](#atomix-multiraft-v1-ClosePrimitiveOutput) |  |  |






<a name="atomix-multiraft-v1-CreatePrimitiveRequest"></a>

### CreatePrimitiveRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [SessionRequestHeaders](#atomix-multiraft-v1-SessionRequestHeaders) |  |  |
| input | [CreatePrimitiveInput](#atomix-multiraft-v1-CreatePrimitiveInput) |  |  |






<a name="atomix-multiraft-v1-CreatePrimitiveResponse"></a>

### CreatePrimitiveResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [SessionResponseHeaders](#atomix-multiraft-v1-SessionResponseHeaders) |  |  |
| output | [CreatePrimitiveOutput](#atomix-multiraft-v1-CreatePrimitiveOutput) |  |  |





 

 

 


<a name="atomix-multiraft-v1-Session"></a>

### Session


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreatePrimitive | [CreatePrimitiveRequest](#atomix-multiraft-v1-CreatePrimitiveRequest) | [CreatePrimitiveResponse](#atomix-multiraft-v1-CreatePrimitiveResponse) |  |
| ClosePrimitive | [ClosePrimitiveRequest](#atomix-multiraft-v1-ClosePrimitiveRequest) | [ClosePrimitiveResponse](#atomix-multiraft-v1-ClosePrimitiveResponse) |  |

 



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

