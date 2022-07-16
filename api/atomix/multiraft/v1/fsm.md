# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/v1/fsm.proto](#atomix_multiraft_v1_fsm-proto)
    - [CloseServiceInput](#atomix-multiraft-v1-CloseServiceInput)
    - [CloseServiceOutput](#atomix-multiraft-v1-CloseServiceOutput)
    - [CloseSessionInput](#atomix-multiraft-v1-CloseSessionInput)
    - [CloseSessionOutput](#atomix-multiraft-v1-CloseSessionOutput)
    - [CreateServiceInput](#atomix-multiraft-v1-CreateServiceInput)
    - [CreateServiceOutput](#atomix-multiraft-v1-CreateServiceOutput)
    - [KeepAliveInput](#atomix-multiraft-v1-KeepAliveInput)
    - [KeepAliveInput.OutputStreamSequenceNumsEntry](#atomix-multiraft-v1-KeepAliveInput-OutputStreamSequenceNumsEntry)
    - [KeepAliveOutput](#atomix-multiraft-v1-KeepAliveOutput)
    - [OpenSessionInput](#atomix-multiraft-v1-OpenSessionInput)
    - [OpenSessionOutput](#atomix-multiraft-v1-OpenSessionOutput)
    - [PartitionInput](#atomix-multiraft-v1-PartitionInput)
    - [PartitionOutput](#atomix-multiraft-v1-PartitionOutput)
    - [ServiceInput](#atomix-multiraft-v1-ServiceInput)
    - [ServiceOutput](#atomix-multiraft-v1-ServiceOutput)
    - [SessionInput](#atomix-multiraft-v1-SessionInput)
    - [SessionOutput](#atomix-multiraft-v1-SessionOutput)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_v1_fsm-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/v1/fsm.proto



<a name="atomix-multiraft-v1-CloseServiceInput"></a>

### CloseServiceInput







<a name="atomix-multiraft-v1-CloseServiceOutput"></a>

### CloseServiceOutput







<a name="atomix-multiraft-v1-CloseSessionInput"></a>

### CloseSessionInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-CloseSessionOutput"></a>

### CloseSessionOutput







<a name="atomix-multiraft-v1-CreateServiceInput"></a>

### CreateServiceInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service | [string](#string) |  |  |
| namespace | [string](#string) |  |  |
| name | [string](#string) |  |  |






<a name="atomix-multiraft-v1-CreateServiceOutput"></a>

### CreateServiceOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service_id | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-KeepAliveInput"></a>

### KeepAliveInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [uint64](#uint64) |  |  |
| last_command_sequence_num | [uint64](#uint64) |  |  |
| pending_commands_filter | [bytes](#bytes) |  |  |
| output_stream_sequence_nums | [KeepAliveInput.OutputStreamSequenceNumsEntry](#atomix-multiraft-v1-KeepAliveInput-OutputStreamSequenceNumsEntry) | repeated |  |






<a name="atomix-multiraft-v1-KeepAliveInput-OutputStreamSequenceNumsEntry"></a>

### KeepAliveInput.OutputStreamSequenceNumsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [uint64](#uint64) |  |  |
| value | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-KeepAliveOutput"></a>

### KeepAliveOutput







<a name="atomix-multiraft-v1-OpenSessionInput"></a>

### OpenSessionInput







<a name="atomix-multiraft-v1-OpenSessionOutput"></a>

### OpenSessionOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-PartitionInput"></a>

### PartitionInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| open_session | [OpenSessionInput](#atomix-multiraft-v1-OpenSessionInput) |  |  |
| keep_alive | [KeepAliveInput](#atomix-multiraft-v1-KeepAliveInput) |  |  |
| close_session | [CloseSessionInput](#atomix-multiraft-v1-CloseSessionInput) |  |  |
| session_input | [SessionInput](#atomix-multiraft-v1-SessionInput) |  |  |






<a name="atomix-multiraft-v1-PartitionOutput"></a>

### PartitionOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| open_session | [OpenSessionOutput](#atomix-multiraft-v1-OpenSessionOutput) |  |  |
| keep_alive | [KeepAliveOutput](#atomix-multiraft-v1-KeepAliveOutput) |  |  |
| close_session | [CloseSessionOutput](#atomix-multiraft-v1-CloseSessionOutput) |  |  |
| session_output | [SessionOutput](#atomix-multiraft-v1-SessionOutput) |  |  |






<a name="atomix-multiraft-v1-ServiceInput"></a>

### ServiceInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service_id | [uint64](#uint64) |  |  |
| payload | [bytes](#bytes) |  |  |






<a name="atomix-multiraft-v1-ServiceOutput"></a>

### ServiceOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| payload | [bytes](#bytes) |  |  |






<a name="atomix-multiraft-v1-SessionInput"></a>

### SessionInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [uint64](#uint64) |  |  |
| create_service | [CreateServiceInput](#atomix-multiraft-v1-CreateServiceInput) |  |  |
| close_service | [CloseServiceInput](#atomix-multiraft-v1-CloseServiceInput) |  |  |
| service_input | [ServiceInput](#atomix-multiraft-v1-ServiceInput) |  |  |






<a name="atomix-multiraft-v1-SessionOutput"></a>

### SessionOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| create_service | [CreateServiceOutput](#atomix-multiraft-v1-CreateServiceOutput) |  |  |
| close_service | [CloseServiceOutput](#atomix-multiraft-v1-CloseServiceOutput) |  |  |
| service_output | [ServiceOutput](#atomix-multiraft-v1-ServiceOutput) |  |  |





 

 

 

 



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

