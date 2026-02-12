---
name: LiveKit SFU Management
description: Manage LiveKit rooms, participants, and data channels.
---

# LiveKit SFU Decision Tree

This skill provides expertise in managing LiveKit API resources and SFU data routing.

1.  **Room Management**
    *   **Trigger**: Create, List, or Delete Rooms.
    *   **Action**: Use `lksdk.RoomServiceClient`.
    *   **Tools**:
        *   `CreateRoom(ctx, req)`
        *   `RemoveParticipant(ctx, req)`
    *   **Constraint**: Always enforce `EmptyTimeout` and `MaxParticipants` limits if applicable.

2.  **Authentication & Access Control**
    *   **Trigger**: Generating access tokens for clients.
    *   **Action**: Use `auth.NewAccessToken(apiKey, apiSecret)`.
    *   **Logic**:
        *   Grant `VideoGrant{RoomJoin: true, Room: "name"}`.
        *   Use `SetIdentity(identity)`.
        *   Use `SetValidFor(duration)`.
        *   Optional: Set `CanPublish: false` or `SIPGrant` if needed.

3.  **Data Channel Routing**
    *   **Trigger**: Sending custom binary payloads or messages.
    *   **Action**: Use `localParticipant.PublishDataPacket(payload, options)`.
    *   **Evaluation Table**:
        *   **Reliable (SCTP)**: For Chat, RPC, Config Sync.
            *   Flag: `WithDataPublishReliable(true)`
            *   Max Size: **15 KiB**
        *   **Lossy (Default)**: For Telemetry, Cursors, transient data.
            *   Flag: `WithDataPublishReliable(false)`
            *   Max Size: **1300 Bytes** (MTU limit).

4.  **State Synchronization**
    *   **Trigger**: Synchronizing participant metadata.
    *   **Warning**: `Participant.Metadata` is limited to **64 KiB** and incurs overhead.
    *   **Action**: Use `ParticipantAttributesChangedFunc` callback.
    *   **Alternative**: Use specific Data Packets for high-frequency updates.
