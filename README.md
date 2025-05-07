# Racket - A multicast messaging system

Racket is a lightweight, broker-less messaging system that allows sending messages to multiple subscribers.
It is designed to be simple and easy to use, with a focus on performance and reliability.

Messages are sent to a multicast address that is derived from the stream name and on
multiple ports.

## Broker-less

Racket is a broker-less messaging system, which means that there is no central server or broker that
handles the messages. Instead, each sender and receiver is responsible for sending and receiving messages
directly to and from each other. This allows for a more decentralized and scalable architecture, as there is
no single point of failure and no need for a central server to handle the messages.

# Concepts

## Multicast pool

A multicast pool is a set of multicast addresses that are used to send messages to multiple subscribers.
It is defined through a base IP and a netmask. The multicast pool configuration has to be identical on all
nodes in the network.

## Streams

A stream is a named channel that can be used to send and receive messages. Each stream has a unique name
and can be used to send messages to multiple subscribers. The stream name is used to derive the multicast address
by hashing the stream name and mapping it to one address in the multicast pool.

## Subjects

A subject is a dot-separated string that is used to identify a message. Each message needs to have a subject.
When receiving messages, the subject can be used to filter messages and it may contain a wildcard as the last element,
which will match all messages with the same prefix. The wildcard is represented by a single asterisk (`*`).

## Messages

Messages are sent to a stream and can be received by multiple subscribers. Each message is sent to the multicast address
derived from the stream name on all interfaces of the sender.

## Periodic resend

When receivers are joining the network after a message has been sent, they need to receive the last message
per subject. This is done by periodically resending the last message for each subject, with an interval that is
configured in each message.

## Suppress duplicate messages

Because messages are sent periodically, they will be received multiple times by the same receiver.
To suppress duplicate messages, the receiver keeps track of the hash of the last message received
for each subject. If a message is received with the same hash, it is ignored.

# Example

See the [cmd/sender](cmd/sender) and [cmd/receiver](cmd/receiver) directories for examples of how to use Racket.

# License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.