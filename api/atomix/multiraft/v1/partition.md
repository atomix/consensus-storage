# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/v1/partition.proto](#atomix_multiraft_v1_partition-proto)
    - [CloseSessionRequest](#atomix-multiraft-v1-CloseSessionRequest)
    - [CloseSessionResponse](#atomix-multiraft-v1-CloseSessionResponse)
    - [KeepAliveRequest](#atomix-multiraft-v1-KeepAliveRequest)
    - [KeepAliveResponse](#atomix-multiraft-v1-KeepAliveResponse)
    - [LeaderUpdatedEvent](#atomix-multiraft-v1-LeaderUpdatedEvent)
    - [LogCompactedEvent](#atomix-multiraft-v1-LogCompactedEvent)
    - [LogDBCompactedEvent](#atomix-multiraft-v1-LogDBCompactedEvent)
    - [LogEvent](#atomix-multiraft-v1-LogEvent)
    - [MemberReadyEvent](#atomix-multiraft-v1-MemberReadyEvent)
    - [MembershipChangedEvent](#atomix-multiraft-v1-MembershipChangedEvent)
    - [OpenSessionRequest](#atomix-multiraft-v1-OpenSessionRequest)
    - [OpenSessionResponse](#atomix-multiraft-v1-OpenSessionResponse)
    - [PartitionEvent](#atomix-multiraft-v1-PartitionEvent)
    - [SendSnapshotAbortedEvent](#atomix-multiraft-v1-SendSnapshotAbortedEvent)
    - [SendSnapshotCompletedEvent](#atomix-multiraft-v1-SendSnapshotCompletedEvent)
    - [SendSnapshotStartedEvent](#atomix-multiraft-v1-SendSnapshotStartedEvent)
    - [ServiceConfigChangedEvent](#atomix-multiraft-v1-ServiceConfigChangedEvent)
    - [SnapshotCompactedEvent](#atomix-multiraft-v1-SnapshotCompactedEvent)
    - [SnapshotCreatedEvent](#atomix-multiraft-v1-SnapshotCreatedEvent)
    - [SnapshotReceivedEvent](#atomix-multiraft-v1-SnapshotReceivedEvent)
    - [SnapshotRecoveredEvent](#atomix-multiraft-v1-SnapshotRecoveredEvent)
    - [WatchPartitionRequest](#atomix-multiraft-v1-WatchPartitionRequest)
  
    - [Partition](#atomix-multiraft-v1-Partition)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_v1_partition-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/v1/partition.proto



<a name="atomix-multiraft-v1-CloseSessionRequest"></a>

### CloseSessionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [PartitionRequestHeaders](#atomix-multiraft-v1-PartitionRequestHeaders) |  |  |
| input | [CloseSessionInput](#atomix-multiraft-v1-CloseSessionInput) |  |  |






<a name="atomix-multiraft-v1-CloseSessionResponse"></a>

### CloseSessionResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [PartitionResponseHeaders](#atomix-multiraft-v1-PartitionResponseHeaders) |  |  |
| output | [CloseSessionOutput](#atomix-multiraft-v1-CloseSessionOutput) |  |  |






<a name="atomix-multiraft-v1-KeepAliveRequest"></a>

### KeepAliveRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [PartitionRequestHeaders](#atomix-multiraft-v1-PartitionRequestHeaders) |  |  |
| input | [KeepAliveInput](#atomix-multiraft-v1-KeepAliveInput) |  |  |






<a name="atomix-multiraft-v1-KeepAliveResponse"></a>

### KeepAliveResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [PartitionResponseHeaders](#atomix-multiraft-v1-PartitionResponseHeaders) |  |  |
| output | [KeepAliveOutput](#atomix-multiraft-v1-KeepAliveOutput) |  |  |






<a name="atomix-multiraft-v1-LeaderUpdatedEvent"></a>

### LeaderUpdatedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| term | [uint64](#uint64) |  |  |
| leader | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-LogCompactedEvent"></a>

### LogCompactedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-LogDBCompactedEvent"></a>

### LogDBCompactedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-LogEvent"></a>

### LogEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-MemberReadyEvent"></a>

### MemberReadyEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| node_id | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-MembershipChangedEvent"></a>

### MembershipChangedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| members | [uint64](#uint64) | repeated |  |






<a name="atomix-multiraft-v1-OpenSessionRequest"></a>

### OpenSessionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [PartitionRequestHeaders](#atomix-multiraft-v1-PartitionRequestHeaders) |  |  |
| input | [OpenSessionInput](#atomix-multiraft-v1-OpenSessionInput) |  |  |






<a name="atomix-multiraft-v1-OpenSessionResponse"></a>

### OpenSessionResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [PartitionResponseHeaders](#atomix-multiraft-v1-PartitionResponseHeaders) |  |  |
| output | [OpenSessionOutput](#atomix-multiraft-v1-OpenSessionOutput) |  |  |






<a name="atomix-multiraft-v1-PartitionEvent"></a>

### PartitionEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| partition_id | [uint32](#uint32) |  |  |
| member_ready | [MemberReadyEvent](#atomix-multiraft-v1-MemberReadyEvent) |  |  |
| leader_updated | [LeaderUpdatedEvent](#atomix-multiraft-v1-LeaderUpdatedEvent) |  |  |
| membership_changed | [MembershipChangedEvent](#atomix-multiraft-v1-MembershipChangedEvent) |  |  |
| send_snapshot_started | [SendSnapshotStartedEvent](#atomix-multiraft-v1-SendSnapshotStartedEvent) |  |  |
| send_snapshot_completed | [SendSnapshotCompletedEvent](#atomix-multiraft-v1-SendSnapshotCompletedEvent) |  |  |
| send_snapshot_aborted | [SendSnapshotAbortedEvent](#atomix-multiraft-v1-SendSnapshotAbortedEvent) |  |  |
| snapshot_received | [SnapshotReceivedEvent](#atomix-multiraft-v1-SnapshotReceivedEvent) |  |  |
| snapshot_recovered | [SnapshotRecoveredEvent](#atomix-multiraft-v1-SnapshotRecoveredEvent) |  |  |
| snapshot_created | [SnapshotCreatedEvent](#atomix-multiraft-v1-SnapshotCreatedEvent) |  |  |
| snapshot_compacted | [SnapshotCompactedEvent](#atomix-multiraft-v1-SnapshotCompactedEvent) |  |  |
| log_compacted | [LogCompactedEvent](#atomix-multiraft-v1-LogCompactedEvent) |  |  |
| logdb_compacted | [LogDBCompactedEvent](#atomix-multiraft-v1-LogDBCompactedEvent) |  |  |
| service_config_changed | [ServiceConfigChangedEvent](#atomix-multiraft-v1-ServiceConfigChangedEvent) |  |  |






<a name="atomix-multiraft-v1-SendSnapshotAbortedEvent"></a>

### SendSnapshotAbortedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| to | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-SendSnapshotCompletedEvent"></a>

### SendSnapshotCompletedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| to | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-SendSnapshotStartedEvent"></a>

### SendSnapshotStartedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| to | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-ServiceConfigChangedEvent"></a>

### ServiceConfigChangedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| config | [ServiceConfig](#atomix-multiraft-v1-ServiceConfig) |  |  |






<a name="atomix-multiraft-v1-SnapshotCompactedEvent"></a>

### SnapshotCompactedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-SnapshotCreatedEvent"></a>

### SnapshotCreatedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-SnapshotReceivedEvent"></a>

### SnapshotReceivedEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |
| from | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-SnapshotRecoveredEvent"></a>

### SnapshotRecoveredEvent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-v1-WatchPartitionRequest"></a>

### WatchPartitionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| partition_id | [uint32](#uint32) |  |  |





 

 

 


<a name="atomix-multiraft-v1-Partition"></a>

### Partition


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| OpenSession | [OpenSessionRequest](#atomix-multiraft-v1-OpenSessionRequest) | [OpenSessionResponse](#atomix-multiraft-v1-OpenSessionResponse) |  |
| KeepAlive | [KeepAliveRequest](#atomix-multiraft-v1-KeepAliveRequest) | [KeepAliveResponse](#atomix-multiraft-v1-KeepAliveResponse) |  |
| CloseSession | [CloseSessionRequest](#atomix-multiraft-v1-CloseSessionRequest) | [CloseSessionResponse](#atomix-multiraft-v1-CloseSessionResponse) |  |
| Watch | [WatchPartitionRequest](#atomix-multiraft-v1-WatchPartitionRequest) | [PartitionEvent](#atomix-multiraft-v1-PartitionEvent) stream |  |

 



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

