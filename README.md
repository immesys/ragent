# BOSSWAVE Remote Agent (ragent)

Firstly, before you use this tool, ask yourself if you NEED it. The preferred
way of connecting to the bosswave network is to run your own agent. You should
only consider connecting to a remote agent if you are running a very resource
constrained device (e.g a cellphone). Even a device like a raspberry pi can run
its own agent.

# Overview

What `ragent server` does is listen for connections secured by TLS. When it
receives a connection, it first proves to the client that it is in fact the
VK that the client is expecting (this is so that ragent cannot be MITM'd even
though it has a self-signed certificate). Then the client proves that it is
some VK. Then the server looks for a DOT verifying that the client is permitted
to use this ragent server. See below for the DOT details

# Running a server

To run a ragent server, create an entity for the server (clients will need
the VK of this entity to connect). Then run
```
  ragent server <entity> <listenaddr> <agentaddr>
```
e.g
```
  ragent server myserverentity.ent 0.0.0.0:28590 127.0.0.1:28589
```

# Running a client

The client will provide an OOB socket the same as a real agent, but will
do so by forwarding all the requests to a remote server. To run a client,
run:

```
  ragent client <entity> <server> <servervk> <listenaddr>
```

Note that the full VK must be given because the client cannot resolve an
alias until it has an agent! e.g

```
  ragent client mycliententity.ent ragent.cal-sdb.org:28590 \
    gdIHa4kskW9_gAKm4liWnLPN7lQ8N4L2oqCCdK112fA= \
    127.0.0.1:28589
```

Note that you will not get any indication of success or failure until a
client attempts to open an OOB socket.

# Granting a DOT

In order for a client to use a given ragent server, the entity they use must
have been given permission to use the ragent. This takes the form of a DOT
granted from the server VK to the client VK (directly, with no intermediaries)
granting any permissions on the URI "ragent/1.0/full". e.g

```
  bw2 mkd -f server.ent -t clientVK -u ragent/1.0/full
```

We may later add finer grained permissions on API versions greater than 1.0.
Any revocations or expiries of the DOT or either VK will invalidate the
permissions and future connections will be refused.

# Small print

This is still an alpha. The security model is sound, but the code panics on
literally any error. I'll work on it again in the future to fix that.
