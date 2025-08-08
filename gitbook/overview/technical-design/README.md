# Technical Design

This chapter presents a concise overview of the Skyline Bridge architecture. As previously emphasized, Skyline is explicitly designed to facilitate secure, decentralized, and bi-directional transfers of native blockchain assets - particularly ADA and Apex tokens - thereby ensuring seamless interoperability between the Cardano and Apex Fusion blockchain networks.

Skyline's architectural design incorporates robust mechanisms for securely locking native tokens on their source blockchains, simultaneously unlocking their wrapped token equivalents on the destination blockchain. Through this secure, multisignature-based locking and unlocking approach, asset integrity and value preservation are consistently maintained throughout the bridging process. Consequently, users gain effortless access to cross-chain applications, enabling the utilization of ADA tokens within the Apex Fusion ecosystem and Apex tokens within the Cardano blockchain's dApps.

At its core, Skyline functions as a dedicated, independent blockchain network with its own decentralized validator nodes. These validators play an essential dual role:

1.  Block production

    Validators produce new blocks on the Skyline blockchain, maintaining the network's decentralized consensus and security.
2.  Information collection and validation

    Validators employ specialized, secure components - referred to as trusted components - to collect accurate bridging-related data from both Cardano and Apex Fusion ecosystems. Validators subsequently leverage Skyline's consensus protocol to validate and verify the authenticity, correctness, and integrity of this cross-chain information.

Moreover, these trusted components are instrumental in creating structured Batches of validated bridging requests. Each Batch encapsulates multiple validated bridging requests, optimally organized for efficient submission to the destination blockchain. After validator consensus and cryptographic signing, Batches are securely relayed to the target blockchain network.

The following sections will provide explanations of Skyline's key architectural components, clearly outlining their roles, interdependencies, and the specific dataflows occurring within the system.
