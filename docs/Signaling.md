# Session signaling

Lets assume two Wireguard peers `Pa` & `Pb` are seeking to establish a ICE session.

The smaller public key (PK) of the two peers takes the role of the controlling agent.
In this example PA has the role of the controlling agent as: `PK(PA) < PK(PB)`.

```text
PA                         PB

  --- initial offer     -->        id=SID_Pa, version=0, candidates=[], eoc=false
  <-- initial offer     ---        id=SID_Pb, version=0, candidates=[], eoc=false

  --- subsequent offers -->        id=SID_Pa, version=1, candidates=[C1_Pa], eoc=false
  <-- subsequent offers ---        id=SID_Pb, version=1, candidates=[C1_Pb], eoc=false

  --- subsequent offers -->        id=SID_Pa, version=2, candidates=[C1_Pa, C2_Pa], eoc=false
  <-- subsequent offers ---        id=SID_Pb, version=2, candidates=[C1_Pb, C2_Pb], eoc=false

  ---  eoc. offer       -->        id=SID_Pa, version=3, candidates=[C1_Pa, C2_Pa], eoc=true
  <--  eoc. offer       ---        id=SID_Pb, version=3, candidates=[C1_Pb, C2_Pb], eoc=true
```

## Restart

Agent will restart

- if
  - `last_recv.id` has been set
  - `recv.id!=last_recv.id`
  - `recv.version==0`

- then
  - set
    - `local.id=rand()`
    - `local.version=0`
    - `local.candidates=[]`
  - publish new offer
  - wait for first offer including candidates from remote
  - (re)start agent
  - add first received
  - start gathering candidates
    - send an offers for each candidate `c`:
      - `candidates=local.candidates.append(c)`
      - `id=local.id`
      - `rid=local.rid`
      - `version=local.version++`

## Offer

Offers are exchanged of the signaling backend as packed Protobuf messages:

## Backends

WICE can support multiple backends for signaling session information such as session IDs, ICE candidates, public keys and STUN credentials.

### Available backends

Currently HTTP REST, HTTP WebSockets, Kubernetes and libp2p are supported as backends.
Checkout the `Backend` interface in `wice/backend/backend.go` for implementing your own backend.
