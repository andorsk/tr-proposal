# Trust Registry Proposal

This is sample code for the trust registry proposal to be presented to Trust
Over IP on June 15 by Andor Kesselman. It is slated as a WIP, and intended to
spark discussion, not being a hard unmovable anchor.

**STATUS** STRAWMAN

The presentation is here: [slides](https://docs.google.com/presentation/d/1H31sr0JICoNfChTN2F6eNAmaEHeqwTpK9ukwe7CCjm8/edit)

## Folders

- [examples](./examples)
- [samples](./samples)

## Concepts

There are three main concepts introduced in this proposal:

1. Adding a "profiles" section to a did doc service endpoint. The profiles
   section extends the existing service endpoint descriptions to include
   additional color about how to interact with the service endpoint, and what
   the expected data models are. Check out [here](./example/data/profiles) for
   more details.
2. Defining a profiles base model. Profiles are extensible, self descriptive
   models that desribe what the service does, what version it is on, and how to
   interact with it. There is an expectation that some level of convergence will
   be expected, with service providers eventually sitting between many trust
   registries to act as a mediator in case the trust registry itself does not
   want to make any code changes.
3. Finally, this proposal provides a base container model that provides
   additional contextual information and assurances for an interaction. This
   wraps around the core data, is transport agnostic, and is an optional
   requirement for certain trust registries.
