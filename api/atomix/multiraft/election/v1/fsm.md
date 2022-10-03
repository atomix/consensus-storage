# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/election/v1/fsm.proto](#atomix_multiraft_election_v1_fsm-proto)
    - [AnointInput](#atomix-multiraft-election-v1-AnointInput)
    - [AnointOutput](#atomix-multiraft-election-v1-AnointOutput)
    - [DemoteInput](#atomix-multiraft-election-v1-DemoteInput)
    - [DemoteOutput](#atomix-multiraft-election-v1-DemoteOutput)
    - [EnterInput](#atomix-multiraft-election-v1-EnterInput)
    - [EnterOutput](#atomix-multiraft-election-v1-EnterOutput)
    - [EvictInput](#atomix-multiraft-election-v1-EvictInput)
    - [EvictOutput](#atomix-multiraft-election-v1-EvictOutput)
    - [GetTermInput](#atomix-multiraft-election-v1-GetTermInput)
    - [GetTermOutput](#atomix-multiraft-election-v1-GetTermOutput)
    - [LeaderElectionCandidate](#atomix-multiraft-election-v1-LeaderElectionCandidate)
    - [LeaderElectionInput](#atomix-multiraft-election-v1-LeaderElectionInput)
    - [LeaderElectionOutput](#atomix-multiraft-election-v1-LeaderElectionOutput)
    - [LeaderElectionSnapshot](#atomix-multiraft-election-v1-LeaderElectionSnapshot)
    - [PromoteInput](#atomix-multiraft-election-v1-PromoteInput)
    - [PromoteOutput](#atomix-multiraft-election-v1-PromoteOutput)
    - [Term](#atomix-multiraft-election-v1-Term)
    - [WatchInput](#atomix-multiraft-election-v1-WatchInput)
    - [WatchOutput](#atomix-multiraft-election-v1-WatchOutput)
    - [WithdrawInput](#atomix-multiraft-election-v1-WithdrawInput)
    - [WithdrawOutput](#atomix-multiraft-election-v1-WithdrawOutput)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_election_v1_fsm-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/election/v1/fsm.proto



<a name="atomix-multiraft-election-v1-AnointInput"></a>

### AnointInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| candidate | [string](#string) |  |  |






<a name="atomix-multiraft-election-v1-AnointOutput"></a>

### AnointOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| term | [Term](#atomix-multiraft-election-v1-Term) |  |  |






<a name="atomix-multiraft-election-v1-DemoteInput"></a>

### DemoteInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| candidate | [string](#string) |  |  |






<a name="atomix-multiraft-election-v1-DemoteOutput"></a>

### DemoteOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| term | [Term](#atomix-multiraft-election-v1-Term) |  |  |






<a name="atomix-multiraft-election-v1-EnterInput"></a>

### EnterInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| candidate | [string](#string) |  |  |






<a name="atomix-multiraft-election-v1-EnterOutput"></a>

### EnterOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| term | [Term](#atomix-multiraft-election-v1-Term) |  |  |






<a name="atomix-multiraft-election-v1-EvictInput"></a>

### EvictInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| candidate | [string](#string) |  |  |






<a name="atomix-multiraft-election-v1-EvictOutput"></a>

### EvictOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| term | [Term](#atomix-multiraft-election-v1-Term) |  |  |






<a name="atomix-multiraft-election-v1-GetTermInput"></a>

### GetTermInput







<a name="atomix-multiraft-election-v1-GetTermOutput"></a>

### GetTermOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| term | [Term](#atomix-multiraft-election-v1-Term) |  |  |






<a name="atomix-multiraft-election-v1-LeaderElectionCandidate"></a>

### LeaderElectionCandidate



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| session_id | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-election-v1-LeaderElectionInput"></a>

### LeaderElectionInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enter | [EnterInput](#atomix-multiraft-election-v1-EnterInput) |  |  |
| withdraw | [WithdrawInput](#atomix-multiraft-election-v1-WithdrawInput) |  |  |
| anoint | [AnointInput](#atomix-multiraft-election-v1-AnointInput) |  |  |
| promote | [PromoteInput](#atomix-multiraft-election-v1-PromoteInput) |  |  |
| demote | [DemoteInput](#atomix-multiraft-election-v1-DemoteInput) |  |  |
| evict | [EvictInput](#atomix-multiraft-election-v1-EvictInput) |  |  |
| get_term | [GetTermInput](#atomix-multiraft-election-v1-GetTermInput) |  |  |
| watch | [WatchInput](#atomix-multiraft-election-v1-WatchInput) |  |  |






<a name="atomix-multiraft-election-v1-LeaderElectionOutput"></a>

### LeaderElectionOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enter | [EnterOutput](#atomix-multiraft-election-v1-EnterOutput) |  |  |
| withdraw | [WithdrawOutput](#atomix-multiraft-election-v1-WithdrawOutput) |  |  |
| anoint | [AnointOutput](#atomix-multiraft-election-v1-AnointOutput) |  |  |
| promote | [PromoteOutput](#atomix-multiraft-election-v1-PromoteOutput) |  |  |
| demote | [DemoteOutput](#atomix-multiraft-election-v1-DemoteOutput) |  |  |
| evict | [EvictOutput](#atomix-multiraft-election-v1-EvictOutput) |  |  |
| get_term | [GetTermOutput](#atomix-multiraft-election-v1-GetTermOutput) |  |  |
| watch | [WatchOutput](#atomix-multiraft-election-v1-WatchOutput) |  |  |






<a name="atomix-multiraft-election-v1-LeaderElectionSnapshot"></a>

### LeaderElectionSnapshot



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| term | [uint64](#uint64) |  |  |
| leader | [LeaderElectionCandidate](#atomix-multiraft-election-v1-LeaderElectionCandidate) |  |  |
| candidates | [LeaderElectionCandidate](#atomix-multiraft-election-v1-LeaderElectionCandidate) | repeated |  |






<a name="atomix-multiraft-election-v1-PromoteInput"></a>

### PromoteInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| candidate | [string](#string) |  |  |






<a name="atomix-multiraft-election-v1-PromoteOutput"></a>

### PromoteOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| term | [Term](#atomix-multiraft-election-v1-Term) |  |  |






<a name="atomix-multiraft-election-v1-Term"></a>

### Term



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| leader | [string](#string) |  |  |
| candidates | [string](#string) | repeated |  |
| index | [uint64](#uint64) |  |  |






<a name="atomix-multiraft-election-v1-WatchInput"></a>

### WatchInput







<a name="atomix-multiraft-election-v1-WatchOutput"></a>

### WatchOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| term | [Term](#atomix-multiraft-election-v1-Term) |  |  |






<a name="atomix-multiraft-election-v1-WithdrawInput"></a>

### WithdrawInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| candidate | [string](#string) |  |  |






<a name="atomix-multiraft-election-v1-WithdrawOutput"></a>

### WithdrawOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| term | [Term](#atomix-multiraft-election-v1-Term) |  |  |





 

 

 

 



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

