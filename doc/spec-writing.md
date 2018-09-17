# Changing the Ignition spec

This goes over guidelines for making changes to the Ignition spec. It does not cover the mechanical
aspects (see development.md) but rather what to check to ensure the spec stays backwards compatible,
forwards compatible, and declarative.

## Background on config parsing, appending, etc

The config provided by the platform (e.g. ec2 metadata) is not the config that gets executed.
Internally Ignition has a base config which defines the root partition. A platform specific config
gets appended to that and the platform provided config gets appended to that. This platform
provided config can append other configs as well or even replace itself with another config.

Any of these configs can be any version. Internally when a config is fetched, it's first translated
to the latest experimental config before being appended. This applies to the internal configs as
well, so from Ignition's perspective it's only ever dealing with appending/replacing the latest
config spec.

## Backwards compatibility

Ignition maintains backwards compatibility in two ways:

1) Any old spec can be translated to a new spec with the same meaning. This is important because
all configs are translated to the latest version internally. If this weren't the case, operating
systems with newer Ignition versions would do different things which would go against the
declarative nature of the Ignition spec.

This is implemented by having each spec version know how to translate from the previous, then
repeatedly translating the config until it is at the latest version.

2) Within a major version, the version number can be bumped without changing the meaning of the
config. Minor version bumps should not include breaking changes. For a spec that means changing the
version does not require changing anything else.

## Declarative not imperative

The Ignition spec should be completely declarative. An Ignition config does not describe steps to
provision a machine, it describes the state the machine should be in after provisioning.

## Challenges and gotchas

Maintaining backwards compatability is the biggest challenge, both ensuring that the translation
from the previous spec is lossless and ensuring that the new spec doesn't make creating future spec
versions harder.

Another major challenge is ensuring the spec is declarative. This means the order in which items
are specified should not have an impact on the meaning of the spec. There can be a specific order
in which items are processed, but it must be dependent on the contents of the items, not the order
specified. For example, directories are created from shallowest (e.g. /a) to deepest (/a/b/c/d).

PXE systems that have a stateful partition are another complicating factor. In this case every boot
is first boot, but there might be important data that should not be wiped. This was the reason for
introducing the 2.1.0 filesystem and 2.3.0-exp partition semantics.
