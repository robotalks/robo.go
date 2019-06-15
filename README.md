# Robotalks - Robo.go

A layered Robotic framework for Go.
Please read [Layers](doc/Layers.md) for more details about the thoughts behind,
and the general idea of the [Framework](doc/Framework.md).

## Common Messages

A set of commonly used messages for L1 and above are defined in
[robo.proto](https://github.com/robotalks/robo.proto).

## Tools Provided

- `robocli`: an interactive CLI to send commands to controllers;
- `robomon`: a tool to monitor communication on the MQTT broker;
- `joystickd`: a daemon use Joystick to control robots supports Nav2D commands;
- `sim-nav`: a simulated robot implementing Nav2D commands.

## Simulation

Simulation is currently performed with simple 2D visualization provided by 
[see](https://github.com/robotalks/see).
