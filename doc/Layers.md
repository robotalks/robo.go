# Layered Robotic Framework

Robotic software is complicated because different pieces in the stack focus
on different tasks with different requirements.
It's clearer to separate the whole stack into layers that each layer has
specific requirements.
And the framework is all about the communication and collaboration between
the layers.

## Layers

- L0: firmware, running on MCUs for accuracy and realtime tasks;
- L1: controller, it's on-robot controlling logic, representing a robot;
- L2: (aka L1+) brain, more complicated controlling logic, 
      running either on-robot or remotely.

### L0 Firmware

Firmware performs the low-level controlling tasks and directly interacts with
hardware. The specific requirements on this layer are:

- Accurate control of each hardware component;
- Perform/react in realtime.

This layer talks _L0 protocol_ with L1 controller, usually through a serial port 
(TTL serial port, USB serial port, or bluetooth, etc.).
The _L0 protocol_ exposes hardware specific primitives, for example, using
pulse width instead of common units (e.g. mm), counters from quadrac encoders
instead of actual distance (e.g. inch/mm).

### L1 Controller

Controller works closely with firmware to hide all 
hardware-specific primitives and talks _L1 protocol_ with L2 (and above)
software components.
The _L1 protocol_ provides hardware-agnostic primitives, for example, using
common units for measurements (e.g. inch/mm for distance, radians/degrees for 
angles). The requirements for this layer are:

- Hardware-agnostic: the primitives exposed are generic and selective based
  on the capability of actual hardware;
- Accuracy: the operations should be measured as accurate as possible.

To be hardware-agnostic, calibration is usually needed after assembly. The
calibration data is specific to individual device and stored as configuration.

To be accurate, L2 software usually sends a batch of operations (plan), or
sends the next operation ahead-of-time before the current operation finishes
to avoid the gap communicating for next step. Because of this, the L1 logic
must be capable of scheduling operations and notifying the L2 software about
the status of execution.

### L2 Brain

To be more accurate, this layer is L1+, as above L1 there can be more layers
depending on specific features of the device. Here, these layers are simply
categorized as L2 which performs controlling logic with unlimited complexity.
It runs on-robot with more computing power, or remotely, or both.

## Registry

A registry is a centralized place (not necessary to be a physically centralized 
node/process) where robot devices (L1 controllers) are registered. This place
serves an endpoint that L2 components can discover the robots and connect to an 
individual robot to perform certain operations.
