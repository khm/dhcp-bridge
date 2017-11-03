# dhcp-bridge
work in progress tools to break dhcp out of the box

## impetus

Existing tools -- even 'cloud-native' tools -- mostly work by manually
interfering with text files, databases, dbus endpoints, or other externalities
built into existing dhcp and dns servers:  dnsmasq, isc-dhcpd, unbound, bind,
etc.  This means you get to maintain whatever system inventory management tools
AND THEN you also get to keep track of (often fragile) state for your
infrastructure services.

the goal here is to smash the state -- every dhcp request should be answered 
from a Single Source of Truth, so we don't have to play stupid games
tricking existing programs into working in ways they weren't designed
to.

A lot of this means we drop some RFC compliance as well, since the DHCP
RFCs are opinionated about implementation details.  The goal here is to
make DHCP do what we want, not bend IETF-compliant programs out of
shape.

## alternatives
there are some tools that don't fit this mold -- consider:

name | url | problem
---- | --- | -------
microdhcpd | https://github.com/google/microdhcpd | requires buy-in to the entire gRPC stack, including vmregistry.
etcdhcp | https://github.com/lclarkmichalek/etcdhcp | is a classical dhcp service with a cloud-native backing store.
shiva | https://github.com/nlamirault/shiva | another classical dhcp service, with several backing store options.

## direction
None of these solutions are trivial to adapt to what I'm after -- the ability
to have a service that takes the requesting mac address, throw everything else
in the trash, and ask some other service what to send back.  That's half the
battle.

The other half is the other end of the bridge -- a service that answers the
questions asked by the previously-described program, and does little else.  

## excuses
For this proof of concept I'm using JSON over HTTP, because the web stack is by 
far the most exercised networking toolset in the world.  With these tools, it's
trivial to insert tons of web stuff in between the provider and the rest of the
world.  You could have a simple CGI service looking things up in your config
management apparatus.  You could have a ton of static files you track in a
repository.  The test code in this directory does a static lookup for a couple
of mac addresses -- but the intention is for it to send a real live DHCP
DISCOVER packet and smnuggle the answers back along the bridge.

## begging
I'm not a programmer.  If you have better code, please send it.
