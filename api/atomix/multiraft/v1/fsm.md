# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/v1/fsm.proto](#atomix_multiraft_v1_fsm-proto)
    - [ClosePrimitiveInput](#atomix-multiraft-v1-ClosePrimitiveInput)
    - [ClosePrimitiveOutput](#atomix-multiraft-v1-ClosePrimitiveOutput)
    - [CloseSessionInput](#atomix-multiraft-v1-CloseSessionInput)
    - [CloseSessionOutput](#atomix-multiraft-v1-CloseSessionOutput)
    - [CreatePrimitiveInput](#atomix-multiraft-v1-CreatePrimitiveInput)
    - [CreatePrimitiveOutput](#atomix-multiraft-v1-CreatePrimitiveOutput)
    - [Failure](#atomix-multiraft-v1-Failure)
    - [KeepAliveInput](#atomix-multiraft-v1-KeepAliveInput)
    - [KeepAliveInput.LastOutputSequenceNumsEntry](#atomix-multiraft-v1-KeepAliveInput-LastOutputSequenceNumsEntry)
    - [KeepAliveOutput](#atomix-multiraft-v1-KeepAliveOutput)
    - [OpenSessionInput](#atomix-multiraft-v1-OpenSessionInput)
    - [OpenSessionOutput](#atomix-multiraft-v1-OpenSessionOutput)
    - [PrimitiveProposalInput](#atomix-multiraft-v1-PrimitiveProposalInput)
    - [PrimitiveProposalOutput](#atomix-multiraft-v1-PrimitiveProposalOutput)
    - [PrimitiveQueryInput](#atomix-multiraft-v1-PrimitiveQueryInput)
    - [PrimitiveQueryOutput](#atomix-multiraft-v1-PrimitiveQueryOutput)
    - [PrimitiveSnapshot](#atomix-multiraft-v1-PrimitiveSnapshot)
    - [PrimitiveSpec](#atomix-multiraft-v1-PrimitiveSpec)
    - [RaftProposal](#atomix-multiraft-v1-RaftProposal)
    - [SessionProposalInput](#atomix-multiraft-v1-SessionProposalInput)
    - [SessionProposalOutput](#atomix-multiraft-v1-SessionProposalOutput)
    - [SessionProposalSnapshot](#atomix-multiraft-v1-SessionProposalSnapshot)
    - [SessionQueryInput](#atomix-multiraft-v1-SessionQueryInput)
    - [SessionQueryOutput](#atomix-multiraft-v1-SessionQueryOutput)
    - [SessionSnapshot](#atomix-multiraft-v1-SessionSnapshot)
    - [Snapshot](#atomix-multiraft-v1-Snapshot)
    - [StateMachineProposalInput](#atomix-multiraft-v1-StateMachineProposalInput)
    - [StateMachineProposalOutput](#atomix-multiraft-v1-StateMachineProposalOutput)
    - [StateMachineQueryInput](#atomix-multiraft-v1-StateMachineQueryInput)
    - [StateMachineQueryOutput](#atomix-multiraft-v1-StateMachineQueryOutput)
  
    - [Failure.Status](#atomix-multiraft-v1-Failure-Status)
    - [SessionProposalSnapshot.Phase](#atomix-multiraft-v1-SessionProposalSnapshot-Phase)
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






<a name="atomix-multiraft-v1-PrimitiveProposalInput"></a>

### PrimitiveProposalInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| primitive_id | [uint64](#uint64) |  |  |
| payload | [bytes](#bytes) |  |  |






<a name="atomix-multiraft-v1-PrimitiveProposalOutput"></a>

### PrimitiveProposalOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| payload | [bytes](#bytes) |  |  |






<a name="atomix-multiraft-v1-PrimitiveQueryInput"></a>

### PrimitiveQueryInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| primitive_id | [uint64](#uint64) |  |  |
| payload | [bytes](#bytes) |  |  |






<a name="atomix-multiraft-v1-PrimitiveQueryOutput"></a>

### PrimitiveQueryOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| payload | [bytes](#bytes) |  |  |






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






<a name="atomix-multiraft-v1-RaftProposal"></a>

### RaftProposal



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| term | [uint64](#uint64) |  |  |
| sequence_num | [uint64](#uint64) |  |  |
| proposal | [StateMachineProposalInput](#atomix-multiraft-v1-StateMachineProposalInput) |  |  |






<a name="atomix-multiraft-v1-SessionProposalInput"></a>

### SessionProposalInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [uint64](#uint64) |  |  |
| sequence_num | [uint64](#uint64) |  |  |
| deadline | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| create_primitive | [CreatePrimitiveInput](#atomix-multiraft-v1-CreatePrimitiveInput) |  |  |
| close_primitive | [ClosePrimitiveInput](#atomix-multiraft-v1-ClosePrimitiveInput) |  |  |
| proposal | [PrimitiveProposalInput](#atomix-multiraft-v1-PrimitiveProposalInput) |  |  |






<a name="atomix-multiraft-v1-SessionProposalOutput"></a>

### SessionProposalOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sequence_num | [uint64](#uint64) |  |  |
| failure | [Failure](#atomix-multiraft-v1-Failure) |  |  |
| create_primitive | [CreatePrimitiveOutput](#atomix-multiraft-v1-CreatePrimitiveOutput) |  |  |
| close_primitive | [ClosePrimitiveOutput](#atomix-multiraft-v1-ClosePrimitiveOutput) |  |  |
| proposal | [PrimitiveProposalOutput](#atomix-multiraft-v1-PrimitiveProposalOutput) |  |  |






<a name="atomix-multiraft-v1-SessionProposalSnapshot"></a>

### SessionProposalSnapshot



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| phase | [SessionProposalSnapshot.Phase](#atomix-multiraft-v1-SessionProposalSnapshot-Phase) |  |  |
| input | [SessionProposalInput](#atomix-multiraft-v1-SessionProposalInput) |  |  |
| pending_outputs | [SessionProposalOutput](#atomix-multiraft-v1-SessionProposalOutput) | repeated |  |
| last_output_sequence_num | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-SessionQueryInput"></a>

### SessionQueryInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [uint64](#uint64) |  |  |
| deadline | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| query | [PrimitiveQueryInput](#atomix-multiraft-v1-PrimitiveQueryInput) |  |  |






<a name="atomix-multiraft-v1-SessionQueryOutput"></a>

### SessionQueryOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| failure | [Failure](#atomix-multiraft-v1-Failure) |  |  |
| query | [PrimitiveQueryOutput](#atomix-multiraft-v1-PrimitiveQueryOutput) |  |  |






<a name="atomix-multiraft-v1-SessionSnapshot"></a>

### SessionSnapshot



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [uint64](#uint64) |  |  |
| state | [SessionSnapshot.State](#atomix-multiraft-v1-SessionSnapshot-State) |  |  |
| timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| last_updated | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="atomix-multiraft-v1-Snapshot"></a>

### Snapshot



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="atomix-multiraft-v1-StateMachineProposalInput"></a>

### StateMachineProposalInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| open_session | [OpenSessionInput](#atomix-multiraft-v1-OpenSessionInput) |  |  |
| keep_alive | [KeepAliveInput](#atomix-multiraft-v1-KeepAliveInput) |  |  |
| close_session | [CloseSessionInput](#atomix-multiraft-v1-CloseSessionInput) |  |  |
| proposal | [SessionProposalInput](#atomix-multiraft-v1-SessionProposalInput) |  |  |






<a name="atomix-multiraft-v1-StateMachineProposalOutput"></a>

### StateMachineProposalOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| open_session | [OpenSessionOutput](#atomix-multiraft-v1-OpenSessionOutput) |  |  |
| keep_alive | [KeepAliveOutput](#atomix-multiraft-v1-KeepAliveOutput) |  |  |
| close_session | [CloseSessionOutput](#atomix-multiraft-v1-CloseSessionOutput) |  |  |
| proposal | [SessionProposalOutput](#atomix-multiraft-v1-SessionProposalOutput) |  |  |






<a name="atomix-multiraft-v1-StateMachineQueryInput"></a>

### StateMachineQueryInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| max_received_index | [uint64](#uint64) |  |  |
| query | [SessionQueryInput](#atomix-multiraft-v1-SessionQueryInput) |  |  |






<a name="atomix-multiraft-v1-StateMachineQueryOutput"></a>

### StateMachineQueryOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| query | [SessionQueryOutput](#atomix-multiraft-v1-SessionQueryOutput) |  |  |





 


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



<a name="atomix-multiraft-v1-SessionProposalSnapshot-Phase"></a>

### SessionProposalSnapshot.Phase


| Name | Number | Description |
| ---- | ------ | ----------- |
| PENDING | 0 |  |
| RUNNING | 1 |  |
| COMPLETE | 2 |  |
| CANCELED | 3 |  |



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

