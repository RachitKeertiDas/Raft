### About Configuration Stores.

We'll separate config from state. This section talks about config.

We'll store configs in JSON format.


### Global Config

These are shared configs that are required by each server. These configs need to be known by each server and must be the same across each server instance. 

Currently, these are:
1. numServers: Number of Servers.
2. RPCMethod: The method of communication - It is ideal to have the same mode of communication. It is technically possible to have each server to have a different mode, but in the interest of simplicity in the codebase, we shall have this the same across all servers.
3. serverInfo:  This is the unique identifying mechanism for the RPC to reach the server.

An example globalConfig.json: 

```
{
  "numServers": 3,
  "RPCMethod": HTTPHandler,
  "serverInfo": {
      "0":{
        "url":""
      },
      "1":{
       "url":""
      },
      "2":{
      "url":""
      }
   }
}

```

### Local config

This stores stuff that is local to each server. i.e. can vary across servers, but should not affect other servers.

0. ServerID : This has to be unique per server. This has to be number 0<=numServers. 
1. LogMethod: This is how the server stores it;s logs internally, this is an abstraction.
2. StateMachine: This is how the server state machine operates internally, again this is an abstraction

