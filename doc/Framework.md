# The Framework

As described in [Layers](Layers.md), the framework is all about communication
and collaboration between layers.

The framework defines a simple communication model and a in-process message 
dispatching mechanism to help build software components.
As L0 firmware is specific to hardware, the framework only covers the 
communication protocol, without defining how to build the firmware.

## Communication Model

The overall communication model uses a point-to-point channel (vs server/client
model) between the layers. For example, between L0 and L1, a single channel is 
used over a serial port.

Over the communication channel, two types of messages are transmitted:

- Commands: a message is sent as a request and a response is expected;
- Events: a message is sent for notification without expecting a response.

The lower end of the communication channel is a _Controller_ which accepts and
executes _Commands_, and also emits _Events_.
The upper end of the communication channel is a _Commander_ which is responsible
for the logic of controlling certain behaviors by sending _Commands_ to the
_Controller_ and consumes _Events_ from the _Controller_ for understanding the
status of the device.

It's possible on L1 and above, a TCP like (server/client model) communication 
will be used (e.g the _Controller_ exposes a TCP port or runs as an RPC server 
over HTTP). So physically, there are N:1 mapping between _Commanders_ and
the _Controller_. When implementing the _Controller_, keep in mind to treat all
the _Commanders_ virtually as a single _Commander_. A common practice for that
is broadcast all _Events_ to all clients, and responses are only sent to the
client which sent the corresponding requests.

### Communication Channels

#### Serial Port

This channel is used for communication between L0 and L1. The protocol introduced
in this framework focused on robustness of data transferring that the communication
is recoverable at certain level from errors (e.g. data corruption due to signal 
interferencing). 
A simple sequence based synchronization mechanism is used for detecting the case
when L0 and L1 are out of sync and re-synchronize to recover the communication.
It's limited without being able to detect every bit alteration, and parity bits can
be enabled for that purpose if needed.

#### MQTT

MQTT is currently supported for communication between L1 and L2 (aka L1+). At this
level, the following features are required from the channel:

- Discovery: allow clients to discover L1 controllers;
- Online/Offline detection: allow clients to detect online/offline L1 controllers;
- Messaging: send commands and receive responses and events.

Please read [MQTT](MQTT.md) for more details. Other alternatives will be added later.

## In-process Messaging

The Framework provides a simple message loop to coordinates multiple parts in
a _Controller_ process without concerning complicated multi-thread/concurrent
problems. The loop runs periodically, and each time, the loop runs, an
_Iteration_ is executed to cover the logic in the _Controller_. The next
_Iteration_ will be schedule after a short time of delay or immediately if
needed. Within each _Iteration_, multiple parts of controlling logic are 
organized as a few list of controlling tasks with different priorities.
And they are executed according to the priority. There are fixed number 
(e.g. 16) of priorities and can be used to control the order of tasks executed
with the following natures:

- Sensing: collect and update states of individual components;
- Controlling: process the states and make decisions;
- Acuation: perform actions based on decisions.

Though it's not necessary to categorize tasks in the way above.

Each _Iteration_ carries a list of messages received between the end of last
iteration and the start of this iteration. Commands are also represented as
messages generally with a specific type. When processing the messages, 
a message can be taken (removed from the iteration) or left for the upcoming
tasks in the same iteration. New messages can be appended in the same 
iteration for upcoming tasks or posted to the next iteration.
