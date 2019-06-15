# MQTT as Communication Channel

MQTT is used as a communication channel between L1 and L1+ components.

## Topic Schema

The base topic is derived from MQTT broker URL, e.g.

```
mqtt://localhost:1883/robo/
```

Where `robo/` is served as the prefix of all topics.
The base topic for a specific robot (L1 controller) is constructed using
_Type_ and _ID_:

- _Type_: represents the class of the device;
- _ID_: the unique ID of a particular device.

For example, the following topic base is used for Nav2D simulation robot:

```
robo/sim-nav/0/
```

Where _Type_ is `sim-nav`, _ID_ is `0` (usually, machine ID is used here.)

The following topics are used to communicate with the robot:

- `robo/sim-nav/0/cmd`: (protobuf binary format)
  for _Commander_ to write command messages to _Controller_;
- `robo/sim-nav/0/msg`: (protobuf binary format)
  for _Controller_ to write responses and events to _Commander_;
- `robo/sim-nav/0/meta`: (JSON format)
  robot metadata, used for discovery and online state.

## Discovery and Online State

Specially, `/meta` suffix in topic is used identify the online state of the 
robot. When the _Controller_ starts and connects to MQTT broker, it publishes
a _retained_ message with JSON content describing some basic information of
the robot. It should also establish the MQTT connection with _will_ enabled.
It's also used as the _will_ topic and an empty message as _will_ content.
If the robot is disconnected unexpectedly, the _retained_ value of this topic
will be removed, and _Commanders_ will be notified about this offline event.
When the _Controller_ gracefully shuts down, it should remove this _retained_
value to indicate the offline state.

As the message is _retained_, it will serve the purpose of discovery. When
a _Commander_ connects to the MQTT broker and subscribes to topic, it will be
notified with a valid JSON content if the _Controller_ is online.
