# UE4网络同步-Actor同步流程梳理

> Actor 的同步主要是属性同步到客户端

# 1 Server同步Actor到客户端

## 1.1 主函数 UNetDriver::ServerReplicateActors

> 服务器在 NetDiver 的 TickFlush 里面，每一帧都会去执行 ServerReplicateActors 来同步 Actor 的相关内容给客户端。

![UNetDriver::ServerReplicateActors](https://note.youdao.com/yws/res/156416/WEBRESOURCE9dadc1ea615cc121b9e6c526ba2cfa01)

**方法名称**：`UNetDriver::ServerReplicateActors`

**参数**：
- `const float DeltaSeconds`：这是从上次更新以来经过的时间，以秒为单位。这个参数被用于最后计算、更新连接的 "LastUpdateTime"，以及在计算 "ActorChannelBudget" 时作为一个重要因素。

**返回值**：
`int32`：返回一个整数，表示在此方法执行过程中，成功更新的网络连接数量。

**主要流程**：

1.  ServerReplicateActors\_BuildConsiderList生成需要同步的Actor列表，
2.  针对每个Client Connection 做如下流程：
    *   对ConsiderList中的Actors 做相关性检查，之后按照NetPriority排序，结果放入PriorityActors。
    *   遍历PriorityActors，针对每个Actor执行replicated(Channel->ReplicateActor())， 复制属性自身和子对象以及组件

### 1.1.1 对象复制 UActorChannel::ReplicateActor

> 负责 Actor 对象的复制

![UActorChannel::ReplicateActor](https://note.youdao.com/yws/res/156431/WEBRESOURCE15f37c5a843773c739a3dbc8abe4f598)

**方法名称**：`UActorChannel::ReplicateActor`

**参数**：
（无）

**返回值**：
`bool` ：复制成功返回true，否则返回false


**主要流程**：

1.  **如果是初始情况，发送初始数据**：如果刚开始进行复制（即`RepFlags.bNetInitial`为真），代码会发送关于Actor的初始信息，包括Actor的属性，其子对象的初始信息。

2.  **复制Actor和Component的属性以及RPCs**：代码在`if (!bIsNewlyReplicationPaused)`的判断里面，根据复制的标志（RepFlags）复制了Actor和其子对象的属性及RPCs。同时，会检查并处理删除的子对象。

    ```cpp
    if (!bIsNewlyReplicationPaused)
    {
        // 复制Actor的属性
        WroteSomethingImportant |= ActorReplicator->ReplicateProperties(Bunch, RepFlags);
        // 复制子对象的属性
        WroteSomethingImportant |= Actor->ReplicateSubobjects(this, &Bunch, &RepFlags);

        // 处理已删除的子对象
        for (auto RepComp = ReplicationMap.CreateIterator(); RepComp; ++RepComp)
        {
            // ...处理已删除子对象的逻辑
        }
    }
    ```

3.  **如果需要，发送复制结果**：在复制过程中如果有重要的信息（WroteSomethingImportant为真），则将复制结果发送出去。

    ```cpp
    if (WroteSomethingImportant)
    {
        // 发送复制的结果
        FPacketIdRange PacketRange = SendBunch( &Bunch, 1 );
        ……
    }
    ```

#### 1.1.1.1 属性复制 FObjectReplicator::ReplicateProperties

> 负责对象属性的复制

![FObjectReplicator::ReplicateProperties](https://note.youdao.com/yws/res/156452/WEBRESOURCEbf820045e3e3b237b456889c21061b09)

**方法名称**：`FObjectReplicator::ReplicateProperties`

**参数**：
- `FOutBunch & Bunch`：用于发送数据的包。
- `FReplicationFlags RepFlags`：复制的标志。

**返回值**：
`bool` ：复制成功返回true，否则返回false

**主要流程**：

1.  **初始化**：本函数开始时，获取需要复制的对象并进行一些错误检查。例如检查复制对象是否为`nullptr`。同时还检查了`OwningChannel`，`RepLayout`，`RepState`，`ChangelistMgr`等是否有效，这些对象将在后续复制过程中使用到。
    ```cpp
       Object = GetObject();
       if ( Object == nullptr )
       {
           return false;
       }
       
       check(OwningChannel);
       check(RepLayout.IsValid());
       check(RepState.IsValid());
    ```

2.  **更新改变列表**：通过调用`FNetSerializeCB::UpdateChangelistMgr`函数，其中会对比对象的属性，并将对象的属性变化信息更新到`ChangelistMgr`中。`ChangelistMgr`会存储所有已改变的属性列表。如果在更新改变列表时出现错误，会将网络连接设置为等待关闭的状态，并结束函数。

    ```cpp
    const ERepLayoutResult UpdateResult = FNetSerializeCB::UpdateChangelistMgr(*RepLayout, SendingRepState, *ChangelistMgr, Object, Connection->Driver->ReplicationFrame, RepFlags, OwningChannel->bForceCompareProperties || bUseCheckpointRepState);
     
    if (UNLIKELY(ERepLayoutResult::FatalError == UpdateResult))
    {
        Connection->SetPendingCloseDueToReplicationFailure();
        return false;
    }
    ```

3.  **复制属性**：`RepLayout`将结合`ChangelistMgr`中的改变列表，找出发生变化的属性，并从对象中获取这些属性的当前值，将这些值写入`Writer`以供发送。`RepLayout->ReplicateProperties`函数主要负责这个步骤。

    ```cpp
    bHasRepLayout = RepLayout->ReplicateProperties(SendingRepState, ChangelistMgr->GetRepChangelistState(), (uint8*)Object, ObjectClass, OwningChannel, Writer, RepFlags);
    ```

4. **复制自定义Delta属性**：此步骤将处理那些需要自定义复制逻辑的属性，例如某些动态数组。对于这些属性，我们使用"delta compression"技术，只复制实际发生变化的部分，从而大大减少了网络流量。

   ```cpp
   ReplicateCustomDeltaProperties(Writer, RepFlags);
   ```

5. **复制队列的远程函数**：`Writer.SerializeBits( RemoteFunctions->GetData(), RemoteFunctions->GetNumBits() );` 这部分代码将队列中的远程函数（Unreliable Functions）写入`Writer`。这些远程函数是那些被标记为`unreliable`的函数，它们可能会被丢弃，不保证会被接收方接收到。这里的`SerializeBits`操作就是将这些函数序列化为二进制数据，以便通过网络发送。

   ```cpp
   Writer.SerializeBits( RemoteFunctions->GetData(), RemoteFunctions->GetNumBits() );
   ```

6.  **发送数据**：在完成所有属性的复制后，本函数会检查是否有写入`Writer`的数据。如果有，调用`OwningChannel->WriteContentBlockPayload`将数据写入网络包。若存在写入的重要数据（即`Writer`中包含数据），函数返回true，否则返回false。


# 2 客户端接收同步流程
（待补充）


参考链接：
1. [UE4网络同步-Actor同步流程详解 - 知乎](https://zhuanlan.zhihu.com/p/533713823)
