# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/v1/fsm.proto](#atomix_multiraft_v1_fsm-proto)
    - [ClosePrimitiveInput](#atomix-multiraft-v1-ClosePrimitiveInput)
    - [ClosePrimitiveOutput](#atomix-multiraft-v1-ClosePrimitiveOutput)
    - [CloseSessionInput](#atomix-multiraft-v1-CloseSessionInput)
    - [CloseSessionOutput](#atomix-multiraft-v1-CloseSessionOutput)
    - [CommandInput](#atomix-multiraft-v1-CommandInput)
    - [CommandOutput](#atomix-multiraft-v1-CommandOutput)
    - [CommandSnapshot](#atomix-multiraft-v1-CommandSnapshot)
    - [CreatePrimitiveInput](#atomix-multiraft-v1-CreatePrimitiveInput)
    - [CreatePrimitiveOutput](#atomix-multiraft-v1-CreatePrimitiveOutput)
    - [Failure](#atomix-multiraft-v1-Failure)
    - [KeepAliveInput](#atomix-multiraft-v1-KeepAliveInput)
    - [KeepAliveInput.LastOutputSequenceNumsEntry](#atomix-multiraft-v1-KeepAliveInput-LastOutputSequenceNumsEntry)
    - [KeepAliveOutput](#atomix-multiraft-v1-KeepAliveOutput)
    - [OpenSessionInput](#atomix-multiraft-v1-OpenSessionInput)
    - [OpenSessionOutput](#atomix-multiraft-v1-OpenSessionOutput)
    - [OperationInput](#atomix-multiraft-v1-OperationInput)
    - [OperationOutput](#atomix-multiraft-v1-OperationOutput)
    - [PartitionQueryInput](#atomix-multiraft-v1-PartitionQueryInput)
    - [PartitionQueryOutput](#atomix-multiraft-v1-PartitionQueryOutput)
    - [PrimitiveOperationInput](#atomix-multiraft-v1-PrimitiveOperationInput)
    - [PrimitiveOperationOutput](#atomix-multiraft-v1-PrimitiveOperationOutput)
    - [PrimitiveSnapshot](#atomix-multiraft-v1-PrimitiveSnapshot)
    - [PrimitiveSpec](#atomix-multiraft-v1-PrimitiveSpec)
    - [QueryInput](#atomix-multiraft-v1-QueryInput)
    - [QueryOutput](#atomix-multiraft-v1-QueryOutput)
    - [SessionCommandInput](#atomix-multiraft-v1-SessionCommandInput)
    - [SessionCommandOutput](#atomix-multiraft-v1-SessionCommandOutput)
    - [SessionQueryInput](#atomix-multiraft-v1-SessionQueryInput)
    - [SessionQueryOutput](#atomix-multiraft-v1-SessionQueryOutput)
    - [SessionSnapshot](#atomix-multiraft-v1-SessionSnapshot)
    - [Snapshot](#atomix-multiraft-v1-Snapshot)
  
    - [CommandSnapshot.State](#atomix-multiraft-v1-CommandSnapshot-State)
    - [Failure.Status](#atomix-multiraft-v1-Failure-Status)
    - [SessionSnapshot.State](#atomix-multiraft-v1-SessionSnapshot-State)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_v1_fsm-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/v1/fsm.proto



<a name="atomix-multiraft-v1-ClosePrimitiveInput"></a>

### ClosePrimitiveInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| primitive_id | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-ClosePrimitiveOutput"></a>

### ClosePrimitiveOutput







<a name="atomix-multiraft-v1-CloseSessionInput"></a>

### CloseSessionInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-CloseSessionOutput"></a>

### CloseSessionOutput







<a name="atomix-multiraft-v1-CommandInput"></a>

### CommandInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| open_session | [OpenSessionInput](#atomix-multiraft-v1-OpenSessionInput) |  |  |
| keep_alive | [KeepAliveInput](#atomix-multiraft-v1-KeepAliveInput) |  |  |
| close_session | [CloseSessionInput](#atomix-multiraft-v1-CloseSessionInput) |  |  |
| session_command | [SessionCommandInput](#atomix-multiraft-v1-SessionCommandInput) |  |  |






<a name="atomix-multiraft-v1-CommandOutput"></a>

### CommandOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| open_session | [OpenSessionOutput](#atomix-multiraft-v1-OpenSessionOutput) |  |  |
| keep_alive | [KeepAliveOutput](#atomix-multiraft-v1-KeepAliveOutput) |  |  |
| close_session | [CloseSessionOutput](#atomix-multiraft-v1-CloseSessionOutput) |  |  |
| session_command | [SessionCommandOutput](#atomix-multiraft-v1-SessionCommandOutput) |  |  |






<a name="atomix-multiraft-v1-CommandSnapshot"></a>

### CommandSnapshot



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| state | [CommandSnapshot.State](#atomix-multiraft-v1-CommandSnapshot-State) |  |  |
| input | [SessionCommandInput](#atomix-multiraft-v1-SessionCommandInput) |  |  |
| pending_outputs | [SessionCommandOutput](#atomix-multiraft-v1-SessionCommandOutput) | repeated |  |
| last_output_sequence_num | [uint64](#uint64) |  |  |
| deadline | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="atomix-multiraft-v1-CreatePrimitiveInput"></a>

### CreatePrimitiveInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| spec | [PrimitiveSpec](#atomix-multiraft-v1-PrimitiveSpec) |  |  |






<a name="atomix-multiraft-v1-CreatePrimitiveOutput"></a>

### CreatePrimitiveOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| primitive_id | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-Failure"></a>

### Failure



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [Failure.Status](#atomix-multiraft-v1-Failure-Status) |  |  |
| message | [string](#string) |  |  |






<a name="atomix-multiraft-v1-KeepAliveInput"></a>

### KeepAliveInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [uint64](#uint64) |  |  |
| input_filter | [bytes](#bytes) |  |  |
| last_input_sequence_num | [uint64](#uint64) |  |  |
| last_output_sequence_nums | [KeepAliveInput.LastOutputSequenceNumsEntry](#atomix-multiraft-v1-KeepAliveInput-LastOutputSequenceNumsEntry) | repeated |  |






<a name="atomix-multiraft-v1-KeepAliveInput-LastOutputSequenceNumsEntry"></a>

### KeepAliveInput.LastOutputSequenceNumsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [uint64](#uint64) |  |  |
| value | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-KeepAliveOutput"></a>

### KeepAliveOutput







<a name="atomix-multiraft-v1-OpenSessionInput"></a>

### OpenSessionInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="atomix-multiraft-v1-OpenSessionOutput"></a>

### OpenSessionOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-OperationInput"></a>

### OperationInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| payload | [bytes](#bytes) |  |  |






<a name="atomix-multiraft-v1-OperationOutput"></a>

### OperationOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| payload | [bytes](#bytes) |  |  |






<a name="atomix-multiraft-v1-PartitionQueryInput"></a>

### PartitionQueryInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| partition_id | [uint32](#uint32) |  |  |
| query | [QueryInput](#atomix-multiraft-v1-QueryInput) |  |  |
| sync | [bool](#bool) |  |  |






<a name="atomix-multiraft-v1-PartitionQueryOutput"></a>

### PartitionQueryOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| query | [QueryOutput](#atomix-multiraft-v1-QueryOutput) |  |  |






<a name="atomix-multiraft-v1-PrimitiveOperationInput"></a>

### PrimitiveOperationInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| primitive_id | [uint64](#uint64) |  |  |
| input | [OperationInput](#atomix-multiraft-v1-OperationInput) |  |  |






<a name="atomix-multiraft-v1-PrimitiveOperationOutput"></a>

### PrimitiveOperationOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| output | [OperationOutput](#atomix-multiraft-v1-OperationOutput) |  |  |






<a name="atomix-multiraft-v1-PrimitiveSnapshot"></a>

### PrimitiveSnapshot



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| primitive_id | [uint64](#uint64) |  |  |
| spec | [PrimitiveSpec](#atomix-multiraft-v1-PrimitiveSpec) |  |  |






<a name="atomix-multiraft-v1-PrimitiveSpec"></a>

### PrimitiveSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service | [string](#string) |  |  |
| namespace | [string](#string) |  |  |
| name | [string](#string) |  |  |






<a name="atomix-multiraft-v1-QueryInput"></a>

### QueryInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| max_received_index | [uint64](#uint64) |  |  |
| session_query | [SessionQueryInput](#atomix-multiraft-v1-SessionQueryInput) |  |  |






<a name="atomix-multiraft-v1-QueryOutput"></a>

### QueryOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| session_query | [SessionQueryOutput](#atomix-multiraft-v1-SessionQueryOutput) |  |  |






<a name="atomix-multiraft-v1-SessionCommandInput"></a>

### SessionCommandInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [uint64](#uint64) |  |  |
| sequence_num | [uint64](#uint64) |  |  |
| deadline | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| create_primitive | [CreatePrimitiveInput](#atomix-multiraft-v1-CreatePrimitiveInput) |  |  |
| close_primitive | [ClosePrimitiveInput](#atomix-multiraft-v1-ClosePrimitiveInput) |  |  |
| operation | [PrimitiveOperationInput](#atomix-multiraft-v1-PrimitiveOperationInput) |  |  |






<a name="atomix-multiraft-v1-SessionCommandOutput"></a>

### SessionCommandOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sequence_num | [uint64](#uint64) |  |  |
| failure | [Failure](#atomix-multiraft-v1-Failure) |  |  |
| create_primitive | [CreatePrimitiveOutput](#atomix-multiraft-v1-CreatePrimitiveOutput) |  |  |
| close_primitive | [ClosePrimitiveOutput](#atomix-multiraft-v1-ClosePrimitiveOutput) |  |  |
| operation | [PrimitiveOperationOutput](#atomix-multiraft-v1-PrimitiveOperationOutput) |  |  |






<a name="atomix-multiraft-v1-SessionQueryInput"></a>

### SessionQueryInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [uint64](#uint64) |  |  |
| operation | [PrimitiveOperationInput](#atomix-multiraft-v1-PrimitiveOperationInput) |  |  |






<a name="atomix-multiraft-v1-SessionQueryOutput"></a>

### SessionQueryOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| failure | [Failure](#atomix-multiraft-v1-Failure) |  |  |
| operation | [PrimitiveOperationOutput](#atomix-multiraft-v1-PrimitiveOperationOutput) |  |  |






<a name="atomix-multiraft-v1-SessionSnapshot"></a>

### SessionSnapshot



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [uint64](#uint64) |  |  |
| state | [SessionSnapshot.State](#atomix-multiraft-v1-SessionSnapshot-State) |  |  |
| timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| last_updated | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| commands | [CommandSnapshot](#atomix-multiraft-v1-CommandSnapshot) | repeated |  |






<a name="atomix-multiraft-v1-Snapshot"></a>

### Snapshot



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |





 


<a name="atomix-multiraft-v1-CommandSnapshot-State"></a>

### CommandSnapshot.State


| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 |  |
| RUNNING | 1 |  |
| COMPLETE | 2 |  |



<a name="atomix-multiraft-v1-Failure-Status"></a>

### Failure.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| ERROR | 1 |  |
| CANCELED | 2 |  |
| NOT_FOUND | 3 |  |
| ALREADY_EXISTS | 4 |  |
| UNAUTHORIZED | 5 |  |
| FORBIDDEN | 6 |  |
| CONFLICT | 7 |  |
| INVALID | 8 |  |
| UNAVAILABLE | 9 |  |
| NOT_SUPPORTED | 10 |  |
| TIMEOUT | 11 |  |
| FAULT | 12 |  |
| INTERNAL | 13 |  |



<a name="atomix-multiraft-v1-SessionSnapshot-State"></a>

### SessionSnapshot.State


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| OPEN | 1 |  |
| CLOSED | 2 |  |


 

 

 



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

