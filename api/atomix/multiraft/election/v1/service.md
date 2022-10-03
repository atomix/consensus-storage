# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [atomix/multiraft/election/v1/service.proto](#atomix_multiraft_election_v1_service-proto)
    - [AnointRequest](#atomix-multiraft-election-v1-AnointRequest)
    - [AnointResponse](#atomix-multiraft-election-v1-AnointResponse)
    - [DemoteRequest](#atomix-multiraft-election-v1-DemoteRequest)
    - [DemoteResponse](#atomix-multiraft-election-v1-DemoteResponse)
    - [EnterRequest](#atomix-multiraft-election-v1-EnterRequest)
    - [EnterResponse](#atomix-multiraft-election-v1-EnterResponse)
    - [EvictRequest](#atomix-multiraft-election-v1-EvictRequest)
    - [EvictResponse](#atomix-multiraft-election-v1-EvictResponse)
    - [GetTermRequest](#atomix-multiraft-election-v1-GetTermRequest)
    - [GetTermResponse](#atomix-multiraft-election-v1-GetTermResponse)
    - [PromoteRequest](#atomix-multiraft-election-v1-PromoteRequest)
    - [PromoteResponse](#atomix-multiraft-election-v1-PromoteResponse)
    - [WatchRequest](#atomix-multiraft-election-v1-WatchRequest)
    - [WatchResponse](#atomix-multiraft-election-v1-WatchResponse)
    - [WithdrawRequest](#atomix-multiraft-election-v1-WithdrawRequest)
    - [WithdrawResponse](#atomix-multiraft-election-v1-WithdrawResponse)
  
    - [LeaderElection](#atomix-multiraft-election-v1-LeaderElection)
  
- [Scalar Value Types](#scalar-value-types)



<a name="atomix_multiraft_election_v1_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## atomix/multiraft/election/v1/service.proto



<a name="atomix-multiraft-election-v1-AnointRequest"></a>

### AnointRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [AnointInput](#atomix-multiraft-election-v1-AnointInput) |  |  |






<a name="atomix-multiraft-election-v1-AnointResponse"></a>

### AnointResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [AnointOutput](#atomix-multiraft-election-v1-AnointOutput) |  |  |






<a name="atomix-multiraft-election-v1-DemoteRequest"></a>

### DemoteRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [DemoteInput](#atomix-multiraft-election-v1-DemoteInput) |  |  |






<a name="atomix-multiraft-election-v1-DemoteResponse"></a>

### DemoteResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [DemoteOutput](#atomix-multiraft-election-v1-DemoteOutput) |  |  |






<a name="atomix-multiraft-election-v1-EnterRequest"></a>

### EnterRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [EnterInput](#atomix-multiraft-election-v1-EnterInput) |  |  |






<a name="atomix-multiraft-election-v1-EnterResponse"></a>

### EnterResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [EnterOutput](#atomix-multiraft-election-v1-EnterOutput) |  |  |






<a name="atomix-multiraft-election-v1-EvictRequest"></a>

### EvictRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [EvictInput](#atomix-multiraft-election-v1-EvictInput) |  |  |






<a name="atomix-multiraft-election-v1-EvictResponse"></a>

### EvictResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [EvictOutput](#atomix-multiraft-election-v1-EvictOutput) |  |  |






<a name="atomix-multiraft-election-v1-GetTermRequest"></a>

### GetTermRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [GetTermInput](#atomix-multiraft-election-v1-GetTermInput) |  |  |






<a name="atomix-multiraft-election-v1-GetTermResponse"></a>

### GetTermResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [GetTermOutput](#atomix-multiraft-election-v1-GetTermOutput) |  |  |






<a name="atomix-multiraft-election-v1-PromoteRequest"></a>

### PromoteRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [PromoteInput](#atomix-multiraft-election-v1-PromoteInput) |  |  |






<a name="atomix-multiraft-election-v1-PromoteResponse"></a>

### PromoteResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [PromoteOutput](#atomix-multiraft-election-v1-PromoteOutput) |  |  |






<a name="atomix-multiraft-election-v1-WatchRequest"></a>

### WatchRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryRequestHeaders](#atomix-multiraft-v1-QueryRequestHeaders) |  |  |
| input | [WatchInput](#atomix-multiraft-election-v1-WatchInput) |  |  |






<a name="atomix-multiraft-election-v1-WatchResponse"></a>

### WatchResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.QueryResponseHeaders](#atomix-multiraft-v1-QueryResponseHeaders) |  |  |
| output | [WatchOutput](#atomix-multiraft-election-v1-WatchOutput) |  |  |






<a name="atomix-multiraft-election-v1-WithdrawRequest"></a>

### WithdrawRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandRequestHeaders](#atomix-multiraft-v1-CommandRequestHeaders) |  |  |
| input | [WithdrawInput](#atomix-multiraft-election-v1-WithdrawInput) |  |  |






<a name="atomix-multiraft-election-v1-WithdrawResponse"></a>

### WithdrawResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [atomix.multiraft.v1.CommandResponseHeaders](#atomix-multiraft-v1-CommandResponseHeaders) |  |  |
| output | [WithdrawOutput](#atomix-multiraft-election-v1-WithdrawOutput) |  |  |





 

 

 


<a name="atomix-multiraft-election-v1-LeaderElection"></a>

### LeaderElection
LeaderElection is a service for a leader election primitive

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Enter | [EnterRequest](#atomix-multiraft-election-v1-EnterRequest) | [EnterResponse](#atomix-multiraft-election-v1-EnterResponse) | Enter enters the leader election |
| Withdraw | [WithdrawRequest](#atomix-multiraft-election-v1-WithdrawRequest) | [WithdrawResponse](#atomix-multiraft-election-v1-WithdrawResponse) | Withdraw withdraws a candidate from the leader election |
| Anoint | [AnointRequest](#atomix-multiraft-election-v1-AnointRequest) | [AnointResponse](#atomix-multiraft-election-v1-AnointResponse) | Anoint anoints a candidate leader |
| Promote | [PromoteRequest](#atomix-multiraft-election-v1-PromoteRequest) | [PromoteResponse](#atomix-multiraft-election-v1-PromoteResponse) | Promote promotes a candidate |
| Demote | [DemoteRequest](#atomix-multiraft-election-v1-DemoteRequest) | [DemoteResponse](#atomix-multiraft-election-v1-DemoteResponse) | Demote demotes a candidate |
| Evict | [EvictRequest](#atomix-multiraft-election-v1-EvictRequest) | [EvictResponse](#atomix-multiraft-election-v1-EvictResponse) | Evict evicts a candidate from the election |
| GetTerm | [GetTermRequest](#atomix-multiraft-election-v1-GetTermRequest) | [GetTermResponse](#atomix-multiraft-election-v1-GetTermResponse) | GetTerm gets the current leadership term |
| Watch | [WatchRequest](#atomix-multiraft-election-v1-WatchRequest) | [WatchResponse](#atomix-multiraft-election-v1-WatchResponse) stream | Watch watches the election for events |

 



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

